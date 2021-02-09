package routers

import (
	"github.com/gin-gonic/gin"
	"goblog/api"
)

func InitWebRouter(Router *gin.RouterGroup) {
	BlogRouter := Router.Group("web")
	{
		BlogRouter.GET("blog/list", api.GetBlogs)
		BlogRouter.GET("blog/detail/:id", api.GetBlogDetail)
		BlogRouter.GET("tag/list", api.GetTags)
		BlogRouter.GET("classification/list", api.GetClassifications)
		BlogRouter.GET("links/list", api.GetLinks)
	}
}
