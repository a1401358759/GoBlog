package web

import (
	"github.com/gin-gonic/gin"
	"goblog/core/global"
	"goblog/modules/model"
	"goblog/utils"
	"gorm.io/gorm/clause"
)

// GetBlogList 博客列表
func GetBlogList(c *gin.Context, pageNum, pageSize int) ([]model.Article, int64) {
	db := global.GDb

	title := c.DefaultQuery("title", "")
	var count int64

	// 计算总数
	db = db.Order("id desc").Model(&model.Article{}).Preload(clause.Associations).Where("status = ?", utils.BlogStatus.PUBLISHED)
	if title != "" {
		db = db.Where("title LIKE ?", "%"+title+"%")
	}
	db.Count(&count)
	// 分页
	var blogList []model.Article
	db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&blogList)

	return blogList, count
}

// BlogDetail 博客详情
func BlogDetail(blogID string) (blog model.Article) {
	global.GDb.Preload(clause.Associations).First(&blog, blogID)
	return
}

// 标签列表
func GetTagList() (tagList []map[string]interface{}) {
	global.GDb.Model(&model.Tag{}).Select("id", "name").Scan(&tagList)
	return
}

// 分类列表
func GetClassificationList() (classifications []map[string]interface{}) {
	global.GDb.Model(&model.Classification{}).Select("id", "name").Scan(&classifications)
	return
}

// 友情链接列表
func GetLinkList() (links []map[string]interface{}) {
	global.GDb.Model(&model.Links{}).Select("name", "link", "avatar", "desc").Scan(&links)
	utils.Shuffle(links) // 随机打乱顺序
	return
}
