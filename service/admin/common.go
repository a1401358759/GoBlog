package admin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GetPageParams 获取分页信息
func GetPageParams(c *gin.Context) (pageNum, pageSize int) {
	pageNum, _ = strconv.Atoi(c.DefaultQuery("PageNumber", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("PageSize", "20"))
	return
}

// GetUserInfo 获取登录用户信息
func GetUserInfo(c *gin.Context) (userID int, userEmail string) {
	userID = c.GetInt("UID")
	userEmail = c.GetString("Email")
	return
}

// GenSqlStr 生成数据库查询可以使用的sql格式
func GenSqlStr(param []interface{}) string {
	/*
		[]int{1,2,3} ==> "1","2","3"
	*/
	var paramList []string
	for i := 0; i < len(param); i++ {
		if val, ok := param[i].(int); ok {
			paramList = append(paramList, "'"+strconv.Itoa(val)+"'")
		} else if val, ok := param[i].(string); ok {
			paramList = append(paramList, "'"+val+"'")
		}
	}
	return strings.Join(paramList, ",")
}

func GenSqlStrInt(param []int) string {
	/*
		[]int{1,2,3} ==> "1","2","3"
	*/
	var paramList []string
	for i := 0; i < len(param); i++ {
		paramList = append(paramList, "'"+strconv.Itoa(param[i])+"'")
	}
	return strings.Join(paramList, ",")
}

func GenSqlStrString(param []string) string {
	/*
		[]string{"1","2","3"} ==> "1","2","3"
	*/
	var paramList []string
	for i := 0; i < len(param); i++ {
		paramList = append(paramList, "'"+param[i]+"'")
	}
	return strings.Join(paramList, ",")
}

// Interface2String 报表导出使用，将interface全部转为string
func Interface2String(param interface{}) string {
	if param == nil {
		return ""
	}
	switch param.(type) {
	case string:
		return param.(string)
	case int:
		return strconv.Itoa(param.(int))
	case time.Time:
		return param.(time.Time).Add(time.Hour * 8).Format("2006-01-02 15:04:05")
	case float64:
		return strconv.Itoa(int(param.(float64)))
	case int8:
		return strconv.Itoa(int(param.(int8)))
	case int32:
		return strconv.Itoa(int(param.(int32)))
	case int64:
		return strconv.Itoa(int(param.(int64)))
	default:
		return ""
	}
}

// MinutesToTime 分钟数转化为时分秒
func MinutesToTime(minutes int) string {
	if minutes == 0 {
		return ""
	}
	seconds := minutes * 60

	m := seconds / 60
	s := seconds % 60

	h := m / 60
	m = m % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
