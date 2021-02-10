package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	db := global.GDb

	db.Preload(clause.Associations).First(&blog, blogID)

	var tags []model.Tag
	sql := fmt.Sprintf("select tag.* from tag join article_tags a on tag.id = a.tag_id where article_id = %s;", blogID)
	if err := db.Raw(sql).Find(&tags).Error; err != nil {
		global.GLog.Error("BlogDetail", zap.Any("err", err))
	}
	blog.Tags = tags

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
