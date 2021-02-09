package api

import (
	"github.com/gin-gonic/gin"
	"goblog/core/response"
	"goblog/modules/entity"
	"goblog/service/web"
	"goblog/utils"
)

// GetBlogs 文章列表
func GetBlogs(c *gin.Context) {
	resp := entity.NewResponse()
	pageNum, pageSize := utils.GetPageParams(c)

	data, count := web.GetBlogList(c, pageNum, pageSize)

	resp.SetData(data).SetMeta(map[string]int64{"Count": count})
	response.Ok(resp, c)
	return
}

// GetBlogDetail 博客详情
func GetBlogDetail(c *gin.Context) {
	resp := entity.NewResponse()

	blogID := c.Param("id")
	data := web.BlogDetail(blogID)

	resp.SetData(data)
	response.Ok(resp, c)
	return
}

// GetTags 标签列表
func GetTags(c *gin.Context) {
	resp := entity.NewResponse()

	data := web.GetTagList()

	resp.SetData(data)
	response.Ok(resp, c)
	return
}

// GetClassifications 分类列表
func GetClassifications(c *gin.Context) {
	resp := entity.NewResponse()

	data := web.GetClassificationList()

	resp.SetData(data)
	response.Ok(resp, c)
	return
}

// GetLinks 友情链接列表
func GetLinks(c *gin.Context) {
	resp := entity.NewResponse()

	data := web.GetLinkList()

	resp.SetData(data)
	response.Ok(resp, c)
	return
}
