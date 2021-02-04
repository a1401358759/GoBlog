package routers

import (
	"github.com/gin-gonic/gin"
	"goblog/api"
)

func InitAdminRouter(Router *gin.RouterGroup) {
	// user
	UserRouter := Router.Group("user")
	{
		UserRouter.POST("login", api.Login)               // 登录
		UserRouter.POST("change_password", api.ChangePwd) // 修改密码
		UserRouter.GET("logout", api.Logout)              // 登出
	}
	// admin
	AdminRouter := Router.Group("admin")
	{
		AdminRouter.GET("update/:revisionId", api.Revision) // 更新详情
	}
	// report
	ReportRouter := Router.Group("report")
	{
		ReportRouter.GET("mapping", api.GetReportMapping) // 自定义报表字段mapping信息
	}
	// config
	ConfigRouter := Router.Group("config")
	{
		ConfigRouter.GET("home", api.HomePage) // 首页统计

	}
}
