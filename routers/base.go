package routers

import (
	"github.com/gin-gonic/gin"
	"goblog/core/global"
	"goblog/middleware"
)

func Routers() *gin.Engine {
	var Router = gin.Default()
	// 全局中间件
	Router.Use(middleware.Cors())
	// api 路由组及中间件
	ApiGroup := Router.Group("api/v2/")
	ApiGroup.Use(middleware.GinRecovery(true))
	ApiGroup.Use(middleware.LoginAuth())
	{
		InitAdminRouter(ApiGroup)
	}
	global.GLog.Info("router register success")
	return Router
}
