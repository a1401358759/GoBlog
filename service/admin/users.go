package admin

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
	"goblog/core/global"
	"goblog/modules/model"
	"strconv"
	"time"
)

//JWTSign
type JWTSign struct {
	sign []byte
}

//NewSign
func NewSign() *JWTSign {
	return &JWTSign{
		[]byte("cus-be-better"),
	}
}

//GetToken
func (j JWTSign) GetToken(param JWTParams) (token string, err error) {
	tokenHandler := jwt.NewWithClaims(jwt.SigningMethodHS256, param)
	token, err = tokenHandler.SignedString(j.sign)
	return
}

//ParseToken
func (j JWTSign) ParseToken(token string) (*JWTParams, error) {
	t, err := jwt.ParseWithClaims(token, &JWTParams{}, func(token *jwt.Token) (interface{}, error) {
		return j.sign, nil
	})
	if err != nil {

	}
	if params, ok := t.Claims.(*JWTParams); ok && t.Valid {
		if !CheckTokenExpires(params.UID, token) {
			return nil, fmt.Errorf("token过期，请重新登录。")
		}
		RefreshTokenExpires(params.UID, token, false)
		return params, nil
	}
	return nil, fmt.Errorf("token无效")
}

//JWTParams
type JWTParams struct {
	UID      int    `json:"uid"`
	Email    string `json:"Email"`
	Password string `json:"Password"`
	jwt.StandardClaims
}

//GenerateToken
func GenerateToken(email string, pwd string, uid int) string {
	sign := NewSign()
	tokenParam := JWTParams{
		Password: pwd,
		Email:    email,
		UID:      uid,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "cus-be-better",
			NotBefore: int64(time.Now().Unix() - 1000),
		},
	}
	token, err := sign.GetToken(tokenParam)
	if err != nil {

	}
	RefreshTokenExpires(uid, token, true)
	return token
}

//RefreshTokenExpires
func RefreshTokenExpires(uid int, token string, login bool) {
	// 打开单用户登录模式，则每次删除所有token，然后再设置本次登录的token
	// 只有登录时做此操作，不然同时操作可能会导致token竞态
	if global.GConfig.CUS.OnlyOneUser && login {
		tokenKeys := global.GRedis.Keys("token-*").Val()
		for _, tokenKey := range tokenKeys {
			global.GRedis.Del(tokenKey)
		}
	}
	err := global.GRedis.Set("token-"+strconv.Itoa(uid), token, 15*time.Minute).Err()
	if err != nil {
		fmt.Println("token 有效期设置失败")
	}
}

//CheckTokenExpires
func CheckTokenExpires(uid int, token string) bool {
	t, _ := global.GRedis.Get("token-" + strconv.Itoa(uid)).Result()
	if t == "" {
		return false
	}
	// 顶号登录，旧token失效
	if t != token {
		return false
	}
	return true
}

func PwdEncrypt(str string) string {
	//enc := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	sum := sha256.Sum256([]byte(str + "omega"))
	res := fmt.Sprintf("%x", sum)
	return res
}

func LoginCheck(email, password string) (int, string, error) {
	var user model.User
	global.GDb.Where("Email=?", email).First(&user)
	pwd := PwdEncrypt(password)
	if pwd == user.Password {
		token := GenerateToken(email, password, user.UserID)
		return user.UserID, token, nil
	} else {
		return 0, "", errors.New("密码错误")
	}
}

func ChangePwd(email, password, newPassword string) error {
	changer := global.GDb.Table("user").Where("Email=?", email).Where("Password=?", PwdEncrypt(password)).Update("Password", PwdEncrypt(newPassword))
	if changer.Error != nil {
		return changer.Error
	}
	if changer.RowsAffected == 0 {
		return errors.New("账号或密码错误")
	} else {
		var u model.User
		if err := global.GDb.Table("user").Where("Email=?", email).First(&u).Error; err != nil {
			global.GLog.Error("用户查找失败:", zap.Any("err", err))
			return err
		}
		Logout(u.UserID)
		return nil
	}
}

func Logout(uid int) {
	global.GRedis.Del("token-" + strconv.Itoa(uid))
}
