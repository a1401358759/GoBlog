package routers

import (
	"goblog/api"

	"github.com/gin-gonic/gin"
)

func InitCuspRouter(Router *gin.RouterGroup) {
	CuspRouter := Router.Group("")
	{
		CuspRouter.POST("/ClientWebService/client.asmx", api.WebService)
		CuspRouter.POST("/SimpleAuthWebService/SimpleAuth.asmx", api.WebService)
		//CuspRouter.POST("/ReportingWebService/ReportingWebService.asmx", api.WebService)  // 和cusss共用一个URL
		CuspRouter.POST("/Content/:foldername/:filename", api.DownloadFiles)
	}
}
