package routers

import (
	"github.com/gin-gonic/gin"
	"goblog/api"
)

func InitAdminRouter(Router *gin.RouterGroup) {
	// user
	UserRouter := Router.Group("user")
	{
		// UserRouter.POST("login", api.Login)               // 登录
		// UserRouter.POST("change_password", api.ChangePwd) // 修改密码
		UserRouter.GET("logout", api.Logout) // 登出
	}
	// admin
	AdminRouter := Router.Group("admin")
	{
		AdminRouter.GET("update/:revisionId", api.Test) // 更新详情
	}
}
