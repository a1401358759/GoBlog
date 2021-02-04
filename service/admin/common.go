package admin

import (
	"encoding/json"
	"fmt"
	"goblog/core/global"
	"goblog/modules/model"
	"goblog/service"
	"goblog/utils"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 获取产品分类和对应的title {16: "Silverlight", 22: "CMGE V2020-L", 32: "CMGE V0-G", 58: "CMGE V0-H"}
func GetProductClassification() (map[int]string, map[int]string) {
	db := global.GDb
	var revisions []model.Revision
	db.Select("ProductRevisionID, ClassificationRevisionID").
		Where("ProductRevisionID is not null and ClassificationRevisionID is not null").
		Where("UpdateType not in ('Category', 'Detectoid')").Find(&revisions)

	var productRevisions []int
	var classificationRevisions []int
	for _, item := range revisions {
		productRevisions = append(productRevisions, item.ProductRevisionID)
		classificationRevisions = append(classificationRevisions, item.ClassificationRevisionID)
	}

	allRevisionIds := append(productRevisions, classificationRevisions...)
	revisionTitleDict := GetRevisionTitles(allRevisionIds)

	var productTitle = make(map[int]string)
	var classificationTitle = make(map[int]string)
	for key, value := range revisionTitleDict {
		if utils.ContainInt(key, productRevisions) {
			productTitle[key] = value
		} else {
			classificationTitle[key] = value
		}
	}

	return productTitle, classificationTitle
}

// GetRevisionTitles 根据RevisionID获取title
func GetRevisionTitles(allRevisionIds []int) map[int]string {
	db := global.GDb
	revisionTitleDict := make(map[int]string, 0)
	var propertites []struct {
		RevisionID int
		Title      string
	}

	sql := fmt.Sprintf("select RevisionID, Title from property where Language in ('zh-cn', 'en') and RevisionID in (%s) order by RevisionID, Language asc;", GenSqlStrInt(allRevisionIds))
	db.Raw(sql).Scan(&propertites)

	for i := 0; i < len(propertites); i++ {
		revisionTitleDict[propertites[i].RevisionID] = propertites[i].Title
	}
	return revisionTitleDict
}

func GetAllGroupsDict() map[string]string {
	/*
		获取除所有计算机组和下游服务器组外的所有计算机组
		return: {"6D88DF96-F239-4FA0-B419-1B90DA0C044A": "测试", "B73CA6ED-5727-47F3-84DE-015E03F6A88A": "待分配组"}
	*/
	db := global.GDb
	var groups []model.ComputerTargetGroup
	db.Order("IsBuiltin desc").Select("TargetGroupID, TargetGroupName").Where("TargetGroupID not in (?)", []string{utils.UUIDAllComputer, utils.UUIDGroupDss}).Find(&groups)

	var groupDict = make(map[string]string)
	for _, item := range groups {
		groupDict[item.TargetGroupID] = item.TargetGroupName
	}
	groupDict[utils.UUIDGroupUnassigned] = "待分配组"

	return groupDict
}

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

// GenOperateRecord 生成操作记录
func GenOperateRecord(operate int, desc string, result int, userID int, userEmail string) {
	var record = model.OperateRecord{
		OperateID:     userID,
		Operator:      userEmail,
		Operate:       operate,
		OperateDesc:   desc,
		OperateResult: result,
	}
	if err := global.GDb.Create(&record).Error; err != nil {
		global.GLog.Error("GenOperateRecord", zap.Any("err", err))
	}
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

func GetFullDomainName() string {
	serverConf := service.GetGlobalServerConfig()
	fullDomainName := serverConf.FullDomainName
	return fullDomainName
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

//获取被bundle的map
func BundledMap() (map[int]string, error) {
	var bundledMap = make(map[int]string)
	redisDB := global.GRedis
	bundleList := redisDB.HGetAll("b_r").Val()
	for bundleId, revisionIdStr := range bundleList {
		var revisionIds []int
		if err := json.Unmarshal([]byte(revisionIdStr), &revisionIds); err != nil {
			return nil, err
		}
		for _, revisionId := range revisionIds {
			bundledMap[revisionId] = bundleId
		}
	}
	return bundledMap, nil
}

// 计算导入后数据库变化
func GetImportNumber(oldUpdateCount, oldRevisionCount, oldLanguageCount, oldFileCount int64) (newUpdateCount, newRevisionCount, newLanguageCount, newFileCount int64) {
	var updateCount, revisionCount, languageCount, fileCount int64
	db := global.GDb
	db.Table("update").Count(&updateCount)
	db.Table("revision").Count(&revisionCount)
	db.Table("update_language").Count(&languageCount)
	db.Table("files").Count(&fileCount)
	newUpdateCount = updateCount - oldUpdateCount
	newRevisionCount = revisionCount - oldRevisionCount
	newLanguageCount = languageCount - oldLanguageCount
	newFileCount = fileCount - oldFileCount

	return newUpdateCount, newRevisionCount, newLanguageCount, newFileCount
}
