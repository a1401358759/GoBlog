package cus

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"goblog/core/global"
	initialize "goblog/core/init"
	"goblog/modules/model"
	admin "goblog/service/admin"
	"goblog/utils"
	"strconv"
)

func AllUser() {
	var users []model.User
	global.GDb.Find(&users)
	fmt.Println(aurora.Green("User information: "))
	fmt.Println(aurora.Blue("total: " + strconv.Itoa(len(users))))
	for i := 0; i < len(users); i++ {
		fmt.Println(aurora.Sprintf("%d |  %s", aurora.Magenta(i), aurora.Yellow(users[i].Email)))
	}
	return
}

func DelUser(email string) {
	if email == "admin-omega@cmgos.com" {
		fmt.Println(aurora.Red("管理员用户不可删除."))
		return
	}
	var user model.User
	var err error
	if err = global.GDb.Where("email=?", email).First(&user).Error; err != nil {
		fmt.Println(aurora.Red("用户删除失败:" + err.Error()))
	}
	if err = global.GDb.Where("user_id=?", user.UserID).Delete(&model.CustomReportRules{}).Error; err != nil {
		fmt.Println(aurora.Red("用户删除失败:" + err.Error()))
	}
	if err = global.GDb.Where("user_id=?", user.UserID).Delete(&model.AutoApproveRules{}).Error; err != nil {
		fmt.Println(aurora.Red("用户删除失败:" + err.Error()))
	}
	if err = global.GDb.Where("id=?", user.UserID).Delete(&model.User{}).Error; err != nil {
		fmt.Println(aurora.Red("用户删除失败:" + err.Error()))
	}
	fmt.Println(aurora.Green("用户删除成功."))
	return
}

func CreateUser(email string, password string) {
	var userExists int64
	if err := global.GDb.Table("user").Where("email=?", email).Count(&userExists).Error; err != nil {
		fmt.Println(aurora.Red("用户验证失败."))
		return
	}
	if userExists > 0 {
		fmt.Println(aurora.Red("用户邮箱已被占用，创建失败."))
		return
	}
	encryptPwd := admin.PwdEncrypt(utils.MD5V([]byte(password)))
	if err := global.GDb.Create(&model.User{
		Email:    email,
		Password: encryptPwd,
		Active:   true,
	}).Error; err != nil {
		fmt.Println(aurora.Red("用户创建失败：" + err.Error()))
		return
	}
	fmt.Println(aurora.Green("用户创建成功."))
	return
}

func ChangePwd(email string, password string) {
	var user model.User
	if err := global.GDb.Where("email=?", email).First(&user).Error; err != nil {
		fmt.Println(aurora.Red("用户查找失败:" + err.Error()))
		return
	}
	encryptPwd := admin.PwdEncrypt(utils.MD5V([]byte(password)))
	user.Password = encryptPwd
	if err := global.GDb.Save(&user).Error; err != nil {
		fmt.Println(aurora.Red("密码修改失败:" + err.Error()))
		return
	}
	fmt.Println(aurora.Green("密码修改成功."))
	return
}

func CleanAllComputer() {
	if err := global.GDb.Exec(
		"TRUNCATE TABLE computer_summary_for_microsoft_updates;" +
			"TRUNCATE TABLE computer_in_group;" +
			"TRUNCATE TABLE update_status_per_computer;" +
			"TRUNCATE TABLE computer_statement;" +
			"TRUNCATE TABLE computer_revision_stats;" +
			"TRUNCATE TABLE computer_target;").Error; err != nil {
		fmt.Println(aurora.Red("清空所有计算机及相关信息失败:" + err.Error()))
		return
	}
	// 初始化redis服务
	initialize.Redis()
	global.GRedis.Del("c_i*")
	global.GRedis.Del("m_h:*")
}
