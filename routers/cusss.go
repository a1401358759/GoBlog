package routers

import (
	"goblog/api"

	"github.com/gin-gonic/gin"
)

func InitCusssRouter(Router *gin.RouterGroup) {
	CusssRouter := Router.Group("")
	{
		CusssRouter.POST("/ServerSyncWebService/ServerSyncWebService.asmx", api.SyncService)
		CusssRouter.POST("/DssAuthWebService/DssAuthWebService.asmx", api.SyncService)
		CusssRouter.POST("/ReportingWebService/ReportingWebService.asmx", api.SyncService)
	}
}
