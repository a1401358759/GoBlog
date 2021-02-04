package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
	"go.uber.org/zap"
	"goblog/core/global"
	"goblog/core/response"
	"goblog/modules/model"
	"goblog/service"
	"goblog/utils"
	"goblog/utils/goset"
	"gorm.io/gorm"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ReportRuleList 自定义规则报表列表
func ReportRuleList(userID, pageNum, pageSize int) (newRuleList []map[string]interface{}, count int64) {
	db := global.GDb
	// 计算总数
	db = db.Model(&model.CustomReportRules{}).Where("user_id = ? or built_in is true", userID)
	db.Count(&count)
	// 分页
	var ruleList []model.CustomReportRules
	db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&ruleList)

	// 展示列表前对计算机组进行判断
	ruleIdList := RulesListInvalid(ruleList)
	// 查出所有分页内的计算机更新安装状态报表rule相关的更新
	var relationsDict = make(map[int][]int)

	if len(ruleIdList) > 0 {
		var relations = make([]struct {
			RevisionID, RuleID int
		}, 0)
		if err := db.Model(&model.ComputerUpdateRelation{}).Select("revision_id, rule_id").Where("rule_id in (?) and is_valid is true", ruleIdList).Scan(&relations).Error; err != nil {
			global.GLog.Error("ReportRuleList", zap.Any("err", err))
		}
		for _, item := range relations {
			if _, ok := relationsDict[item.RuleID]; !ok {
				relationsDict[item.RuleID] = []int{item.RevisionID}
			} else {
				relationsDict[item.RuleID] = append(relationsDict[item.RuleID], item.RevisionID)
			}
		}
	}

	newRuleList = make([]map[string]interface{}, 0)
	for _, item := range ruleList {
		var show []string
		_ = json.Unmarshal([]byte(item.Show), &show)
		var screen map[string]map[string][]string
		_ = json.Unmarshal([]byte(item.Screen), &screen)

		newRuleList = append(newRuleList, map[string]interface{}{
			"rule_id":       strconv.Itoa(int(item.ID)),
			"name":          item.Name,
			"service_type":  item.ServiceType,
			"show":          show,
			"screen":        screen,
			"built_in":      item.BuiltIn,
			"desc":          item.Desc,
			"TargetGroupID": item.TargetGroupID,
			"ActionID":      strconv.Itoa(item.ActionID),
			"is_valid":      item.IsValid,
			"revision_ids":  relationsDict[int(item.ID)],
		})
	}
	return newRuleList, count
}

// ReportRuleCreate 自定义报表规则创建
func ReportRuleCreate(c *gin.Context) (code int, msg string, ruleID int) {
	userID, userEmail := GetUserInfo(c)
	params := ReportRuleAdd{}
	err := c.BindJSON(&params)
	if err != nil {
		global.GLog.Error("ReportRuleCreate", zap.Any("err", err))
		return response.FAILED, "参数获取失败", -1
	}
	if params.ServerType == ServiceType.ComputerUpdatesStatus && len(params.RevisionIDs) == 0 {
		return response.ParamNOTEnough, response.ErrorMsg[response.ParamNOTEnough], -1
	}
	showJson, _ := json.Marshal(params.Show)
	screenJson, _ := json.Marshal(params.Screen)
	builtIn := false
	rule := model.CustomReportRules{
		UserID:        userID,
		Name:          params.Name,
		ServiceType:   params.ServerType,
		Show:          string(showJson),
		Screen:        string(screenJson),
		BuiltIn:       &builtIn,
		Desc:          params.Desc,
		TargetGroupID: params.TargetGroupID,
		ActionID:      params.ActionID,
	}

	db := global.GDb
	err = db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(&rule).Error; err != nil {
			return err
		}
		// 创建relations
		if params.ServerType == ServiceType.ComputerUpdatesStatus && len(params.RevisionIDs) > 0 {
			var relations []model.ComputerUpdateRelation
			for _, item := range params.RevisionIDs {
				relations = append(relations, model.ComputerUpdateRelation{RuleID: rule.ID, RevisionID: item, TargetGroupID: params.TargetGroupID})
			}
			if err = tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		return nil
	})
	// 创建操作记录
	if err != nil {
		global.GLog.Error("AddReportRules", zap.Any("err", err))
		detail := "创建自定义报表规则失败"
		GenOperateRecord(OperateAction.CreateReport, detail, ResultStatus.FAILED, userID, userEmail)
		return response.FAILED, response.ErrorMsg[response.FAILED], -1
	} else {
		operateDesc := fmt.Sprintf("创建报表【%s】", params.Name)
		GenOperateRecord(OperateAction.CreateReport, operateDesc, ResultStatus.SUCCESS, userID, userEmail)
		return response.SUCCESS, response.ErrorMsg[response.SUCCESS], int(rule.ID)
	}
}

// ReportRuleEdit 自定义报表规则编辑
func ReportRuleEdit(c *gin.Context) (code int, msg string, ruleID int) {
	userID, _ := GetUserInfo(c)
	params := ReportRuleAdd{}
	err := c.BindJSON(&params)
	if err != nil {
		global.GLog.Error("ReportRuleEdit", zap.Any("err", err))
		return response.FAILED, "参数获取失败", -1
	}
	if params.ServerType == ServiceType.ComputerUpdatesStatus && len(params.RevisionIDs) == 0 {
		return response.ParamNOTEnough, response.ErrorMsg[response.ParamNOTEnough], -1
	}
	showJson, _ := json.Marshal(params.Show)
	screenJson, _ := json.Marshal(params.Screen)
	builtIn := false
	rule := model.CustomReportRules{
		UserID:        userID,
		Name:          params.Name,
		ServiceType:   params.ServerType,
		Show:          string(showJson),
		Screen:        string(screenJson),
		BuiltIn:       &builtIn,
		Desc:          params.Desc,
		TargetGroupID: params.TargetGroupID,
		ActionID:      params.ActionID,
	}

	db := global.GDb
	err = db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Where("id = ? and user_id = ? and built_in is false", params.RuleID, userID).Updates(&rule).Error; err != nil {
			return err
		}
		// 计算机更新安装统计报表需要处理关联表
		if params.ServerType == ServiceType.ComputerUpdatesStatus && len(params.RevisionIDs) > 0 {
			// 删除表中原有的和被编辑规则相关的数据
			if err = tx.Where("rule_id = ?", params.RuleID).Delete(model.ComputerUpdateRelation{}).Error; err != nil {
				return err
			}
			// 新建更新和规则的关系数据
			var relations []model.ComputerUpdateRelation
			for _, item := range params.RevisionIDs {
				relations = append(relations, model.ComputerUpdateRelation{RuleID: rule.ID, RevisionID: item, TargetGroupID: params.TargetGroupID})
			}
			if err = tx.Create(&relations).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		global.GLog.Error("EditReportRules", zap.Any("err", err))
		return response.FAILED, "修改自定义报表规则失败", -1
	} else {
		return response.SUCCESS, response.ErrorMsg[response.SUCCESS], params.RuleID
	}
}

// ReportRuleDel 自定义报表规则删除
func ReportRuleDel(c *gin.Context) (code int, msg string) {
	userID, userEmail := GetUserInfo(c)
	var err error

	var params struct {
		RuleID   int    `json:"rule_id"`
		RuleName string `json:"rule_name"`
	}
	err = c.BindJSON(&params)
	if err != nil {
		global.GLog.Error("ReportRuleDel", zap.Any("err", err))
		return response.FAILED, "参数获取失败"
	}

	db := global.GDb
	err = db.Transaction(func(tx *gorm.DB) error {
		// 如果被删除报表是计算机更新安装状态报表，则删除关系表对应数据
		if err = tx.Where("rule_id = ?", params.RuleID).Delete(model.ComputerUpdateRelation{}).Error; err != nil {
			return err
		}
		// 删除报表数据
		if err = tx.Where("id = ? and user_id = ? and built_in is false", params.RuleID, userID).Delete(model.CustomReportRules{}).Error; err != nil {
			return err
		}
		return nil
	})
	// 创建操作记录
	if err == nil {
		operateDesc := fmt.Sprintf("删除报表【%s】", params.RuleName)
		GenOperateRecord(OperateAction.DeleteReport, operateDesc, ResultStatus.SUCCESS, userID, userEmail)
		return response.SUCCESS, response.ErrorMsg[response.SUCCESS]
	} else {
		detail := "删除自定义报表规则失败"
		GenOperateRecord(OperateAction.DeleteReport, detail, ResultStatus.FAILED, userID, userEmail)
		return response.FAILED, "删除自定义报表规则失败"
	}
}

// RulesListInvalid 报表列表过滤，查看是否失效
func RulesListInvalid(ruleList []model.CustomReportRules) []int {
	var ruleIdList []int
	var groupIdList []string

	db := global.GDb
	if err := db.Model(&model.ComputerTargetGroup{}).Where("TargetGroupID not in (?)", []string{utils.UUIDAllComputer, utils.UUIDGroupDss}).Pluck("TargetGroupID", &groupIdList).Error; err != nil {
		global.GLog.Error("RulesListInvalid", zap.Any("err", err))
	}

	for _, rule := range ruleList {
		// 计算机更新安装统计报表 如果相关计算机组删除则失效
		if rule.ServiceType == ServiceType.ComputerUpdatesStatus {
			ruleIdList = append(ruleIdList, int(rule.ID))
			if rule.IsValid && utils.ContainStr(rule.TargetGroupID, groupIdList) {
				rule.IsValid = false
				db.Save(&rule)
			}
		}
		// 如果为更新报表，则从报表规则中删除相关计算机组id
		if rule.ServiceType == ServiceType.Updates && rule.Screen != "" {
			// 更新删除组，可不修改数据库
			var screen map[string]map[string][]string
			_ = json.Unmarshal([]byte(rule.Screen), &screen)
			if inFilter, ok := screen["in"]; ok {
				if targetGroupId, ok := inFilter["TargetGroupID"]; ok {
					// 将被删除的计算机组和报表本身存储的计算机组做交集
					newGroupIds := goset.StrIntersect(targetGroupId, groupIdList).List()
					if len(newGroupIds) == 0 {
						delete(screen["in"], "TargetGroupID")
						// 如果删除TargetGroupID后，整个in中不包含别的筛选条件，则将整个in删除
						if len(screen["in"]) == 0 {
							delete(screen, "in")
						}
					} else {
						screen["in"]["TargetGroupID"] = newGroupIds
					}
					screenBytes, _ := json.Marshal(screen)
					rule.Screen = string(screenBytes)
					db.Save(&rule)
				}
			}
		}
	}
	return ruleIdList
}

// GenFilterSql 生成报表过滤SQL
func GenFilterSql(screen map[string]map[string][]string) string {
	var filterSql = ""
	if len(screen) > 0 {
		if inSearch, ok := screen["in"]; ok {
			for key, val := range inSearch {
				filterSql += fmt.Sprintf(" AND %s IN (%s)", key, GenSqlStrString(val))
			}
		}
		if rangeSearch, ok := screen["range"]; ok {
			for key, val := range rangeSearch {
				if key == "CreationDate" {
					// 开始时间减8小时
					tmp, _ := time.Parse("2006-01-02 15:04:05", val[0])
					val[0] = tmp.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
					// 结束时间减8小时
					tmp1, _ := time.Parse("2006-01-02 15:04:05", val[1])
					val[1] = tmp1.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
				}
				filterSql += fmt.Sprintf(" AND %s >= '%s' AND %s <= '%s'", key, val[0], key, val[1])
			}
		}
		filterSql = " WHERE" + strings.TrimPrefix(filterSql, " AND")
	}
	return filterSql
}

func ReportCommonSql(tableName string, screen map[string]map[string][]string, show []string, pageNum, pageSize int) (sql string, count int64, statsTime string) {
	filterSql := GenFilterSql(screen)
	fieldStr := strings.Join(show, ",")
	limitSql := fmt.Sprintf(" LIMIT " + strconv.Itoa((pageNum-1)*pageSize) + "," + strconv.Itoa(pageSize)) // 分页

	var statsSql, timeSql string
	if tableName == (model.RevisionStatement{}.TableName()) {
		sql = fmt.Sprintf("SELECT DISTINCT %s FROM %s%s%s;", fieldStr, tableName, filterSql, limitSql)
		statsSql = fmt.Sprintf("SELECT COUNT(1) FROM (SELECT DISTINCT %s FROM %s%s) TMP;", fieldStr, tableName, filterSql)
	} else {
		sql = fmt.Sprintf("SELECT %s FROM %s%s%s;", fieldStr, tableName, filterSql, limitSql)
		statsSql = fmt.Sprintf("SELECT COUNT(1) FROM (SELECT %s FROM %s%s) TMP;", fieldStr, tableName, filterSql)
	}
	timeSql = fmt.Sprintf("SELECT MAX(StatsTime) FROM %s;", tableName)

	db := global.GDb
	// 获取总行数
	db.Raw(statsSql).Scan(&count)
	// 获取统计时间
	db.Raw(timeSql).Scan(&statsTime)

	return
}

// GetComputerUpdatesStatusData 生成计算机更新安装状态数据
func GetComputerUpdatesStatusData(ruleID int, targetGroupID string, actionID int, screen map[string]map[string][]string, show []string, pageNum, pageSize int, isExport bool) (data []map[string]interface{}, fieldsList []string, count int64) {
	db := global.GDb
	// 查询关联表 获取所有需要的RevisionID
	var ruleRevisionIDs []int
	db.Model(&model.ComputerUpdateRelation{}).Where("rule_id = ? and is_valid is true", ruleID).Distinct("revision_id").Pluck("revision_id", &ruleRevisionIDs)
	if len(ruleRevisionIDs) == 0 { // 如果该报表中所有相关的更新均被清理，则报表返回为空
		return
	}

	// 获取某个组下面所有的所有计算机
	var targetIDs []int
	db.Model(&model.ComputerInGroup{}).Where("TargetGroupID = ?", targetGroupID).Pluck("TargetID", &targetIDs)
	if len(targetIDs) == 0 { // 如果没有满足条件的计算机，则报表为空
		return
	}

	fieldsList = utils.DelStringItem(show, "Status")        // Status特殊在另外的表里，且最终返回中文，需要特殊处理
	fieldsList = utils.DelStringItem(show, "ActionID")      // ActionID为必传且单选
	fieldsList = utils.DelStringItem(show, "TargetGroupID") // SQL中不包含组相关联查，所以此处移除避免SQL语句出错

	fieldsStr := strings.Join(MustSelectFields, ",") // 计算机和更新的字段，有些不需要展示，但是这里要查询以便于做筛选
	var statusFilterSql = ""
	// 之前已经判断如果没有相关更新，则返回空列表，所以此处不需要判断
	var revisionCondition = GenSqlStrInt(ruleRevisionIDs)
	revisionFilterSql := fmt.Sprintf("SELECT * FROM revision WHERE UpdateType NOT IN ('Category', 'Detectoid') AND ProductRevisionID IS NOT NULL AND ClassificationRevisionID IS NOT NULL AND RevisionID IN (%s)", revisionCondition)

	if len(screen) > 0 {
		if inSearch, ok := screen["in"]; ok {
			for key, val := range inSearch {
				if key == "Status" {
					statusFilterSql += fmt.Sprintf(" AND %s IN (%s)", key, GenSqlStrString(val))
				} else {
					revisionFilterSql += fmt.Sprintf(" AND %s IN (%s)", key, GenSqlStrString(val))
				}
			}
		}
		if rangeSearch, ok := screen["range"]; ok {
			for key, val := range rangeSearch {
				if key == "CreationDate" {
					// 开始时间减8小时
					tmp, _ := time.Parse("2006-01-02 15:04:05", val[0])
					val[0] = tmp.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
					// 结束时间减8小时
					tmp1, _ := time.Parse("2006-01-02 15:04:05", val[1])
					val[1] = tmp1.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
				}
				revisionFilterSql += fmt.Sprintf(" AND %s >= '%s' AND %s <= '%s'", key, val[0], key, val[1])
			}
		}
	}
	// 判断是否为报表导出，如果为报表，则不需要分页
	var limitSql = ""
	if !isExport {
		limitSql = fmt.Sprintf(" LIMIT %s,%s", strconv.Itoa((pageNum-1)*pageSize), strconv.Itoa(pageSize)) // 分页SQL
	}
	// 将Status转成中文返回给前端
	var statusAliasSql = "CASE Status "
	for key, val := range ComputerUpdateStatusMap {
		statusAliasSql += fmt.Sprintf("WHEN '%s' THEN '%s' ", strconv.Itoa(key), val)
	}
	statusAliasSql += "ELSE '状态未知或不适用' END Status"
	// 严重程度也需要返回中文
	msrcSeveritySql := "CASE MsrcSeverity "
	for key, val := range MSRCSeverityEnMap {
		msrcSeveritySql += fmt.Sprintf("WHEN '%s' THEN '%s' ", key, val)
	}
	msrcSeveritySql += "ELSE '未指定' END MsrcSeverity"
	// 筛选计算机SQL 之前已经判断如果没有相关更新，则返回空列表，所以此处不需要判断
	computerFilterSql := fmt.Sprintf("SELECT * FROM computer_target WHERE TargetID IN (%s)", GenSqlStrInt(targetIDs))
	// 获取数据SQL
	needShowFieds := strings.Join(fieldsList, ",")
	dataSql := fmt.Sprintf(`SELECT %s,%s FROM (
		SELECT IFNULL(Status,'0') Status,%s FROM
		(SELECT %s,%s,
		(SELECT Title FROM property WHERE RevisionID = r.RevisionID AND Language IN ('zh-cn', 'en') ORDER BY Language DESC LIMIT 1) AS Title,
		(SELECT Title FROM property WHERE RevisionID = r.ProductRevisionID AND Language IN ('zh-cn', 'en') ORDER BY Language DESC LIMIT 1) AS ProductTitle,
		(SELECT Title FROM property WHERE RevisionID = r.ClassificationRevisionID AND Language IN ('zh-cn', 'en') ORDER BY Language DESC LIMIT 1) AS ClassificationTitle
		FROM (%s) t,(%s) r) tmp LEFT JOIN update_status_per_computer st
		ON tmp.TargetID = st.TargetID AND tmp.RevisionID = st.RevisionID %s%s) b;`, needShowFieds, statusAliasSql, needShowFieds, fieldsStr, msrcSeveritySql, computerFilterSql, revisionFilterSql, statusFilterSql, limitSql)

	db.Raw(dataSql).Scan(&data)

	if isExport {
		fieldsList = append(fieldsList, "Status")
		fieldsList = append(fieldsList, "ApproveStatus")
		return
	}

	var group model.ComputerTargetGroup
	db.Select("TargetGroupName").Where("TargetGroupID = ?", targetGroupID).First(&group)

	for _, item := range data {
		if group.TargetGroupID == utils.UUIDGroupUnassigned {
			name := "待分配组"
			item["TargetGroupID"] = &name
		} else {
			item["TargetGroupID"] = &group.TargetGroupName
		}
		val, _ := ApproveStatusFilterMap[strconv.Itoa(actionID)]
		item["ActionID"] = &val
	}

	// 查询获取到的数据总数
	countSql := fmt.Sprintf(`
		SELECT COUNT(1) FROM (
		SELECT IFNULL(Status,'0') Status FROM
		(SELECT TargetID,ProductRevisionID,ClassificationRevisionID,RevisionID,MsrcSeverity
		FROM (%s) t,(%s) r) tmp LEFT JOIN update_status_per_computer st
		ON tmp.TargetID = st.TargetID AND tmp.RevisionID = st.RevisionID %s) b;`, computerFilterSql, revisionFilterSql, statusFilterSql)
	db.Raw(countSql).Scan(&count)

	return
}

// GetComputers 获取所有计算机
func GetComputers() (computers []ReportComputerData) {
	db := global.GDb

	if err := db.Raw("select * from computer_target left outer join computer_summary_for_microsoft_updates on computer_target.TargetID = computer_summary_for_microsoft_updates.TargetID").Scan(&computers).Error; err != nil {
		global.GLog.Error("GetComputers", zap.Any("err", err))
	}
	return
}

// GetDeploymentRevisions
func GetDeploymentRevisions() (updates []ReportUpdatesData, revisionsAllCompouterMap map[int]map[string]interface{}, revisionIDs []int) {
	db := global.GDb
	revisionIDs = service.GetAllRevisionIDs()
	revisionsAllCompouterMap = make(map[int]map[string]interface{})
	// 获取每个RevisionID和所有计算机组的审批关系
	var revisionsAllCompouter []struct {
		RevisionID, ActionID int
		AdminName            string
		LastChangeTime       *time.Time
	}
	db.Model(&model.Deployment{}).Select("RevisionID", "ActionID", "AdminName", "LastChangeTime").Where("RevisionID in (?) and TargetGroupID = ?", revisionIDs, utils.UUIDAllComputer).Scan(&revisionsAllCompouter)
	for i := 0; i < len(revisionsAllCompouter); i++ {
		item := revisionsAllCompouter[i]
		revisionsAllCompouterMap[item.RevisionID] = map[string]interface{}{
			"ActionID":       item.ActionID,
			"AdminName":      item.AdminName,
			"LastChangeTime": item.LastChangeTime,
		}
	}

	if err := db.Raw(`select tmp.*, d.AdminName, d.ActionID, d.LastChangeTime
			from (select r.RevisionID, r.KBArticleID, r.SecurityBulletinID, r.ProductRevisionID, r.ClassificationRevisionID, r.CreationDate,
			r.ImportedTime, r.UpdateID, r.RevisionNumber, r.LastChangedAnchor, r.MsrcSeverity, c.TargetGroupID, c.TargetGroupName
      		from revision r, computer_target_group c where r.UpdateType not in ('Category', 'Detectoid')
			and r.ProductRevisionID is not null and r.ClassificationRevisionID is not null and c.TargetGroupID not in
            ('A0A08746-4DBE-4A37-9ADF-9E7652C0B421', 'D374F42A-9BE2-4163-A0FA-3C86A401B7A7')) as tmp
         	left outer join deployment d on tmp.RevisionID = d.RevisionID and d.TargetGroupID = tmp.TargetGroupID;
	`).Scan(&updates).Error; err != nil {
		global.GLog.Error("GetDeploymentRevisions", zap.Any("err", err))
	}
	return
}

// GetRevisions 获取所有更新和对应安装状态统计
func GetRevisions() (allRevisions []ReportUpdateInstallStats, revisionIDs []int) {
	revisionIDs = service.GetAllRevisionIDs()

	db := global.GDb
	if err := db.Raw(`select r.RevisionID as RevisionID, s.* from revision r left outer join statistics_for_per_update s on r.RevisionID = s.RevisionID
		where r.UpdateType NOT IN ('Category', 'Detectoid') AND r.ProductRevisionID IS NOT NULL AND r.ClassificationRevisionID IS NOT NULL;
	`).Scan(&allRevisions).Error; err != nil {
		global.GLog.Error("GetRevisions", zap.Any("err", err))
	}
	return
}

// ReportExportCommon 报表导出
func ReportExportCommon(c *gin.Context, rule *model.CustomReportRules, tableName string, titleList []map[string]interface{}) {
	db := global.GDb
	// 字段过滤
	var screen map[string]map[string][]string
	_ = json.Unmarshal([]byte(rule.Screen), &screen)
	filterSql := GenFilterSql(screen)
	var show []string
	_ = json.Unmarshal([]byte(rule.Show), &show)
	// 筛选出需要展示的字段，去掉不需要展示的字段
	var fieldList []string
	for _, item := range show {
		if !utils.ContainStr(item, NotShowFields) {
			fieldList = append(fieldList, item)
		}
	}
	fieldsStr := strings.Join(show, ",")
	showFieldsStr := strings.Join(fieldList, ",")
	sql := ""
	if rule.ServiceType == ServiceType.Updates {
		sql = fmt.Sprintf("SELECT %s from (SELECT DISTINCT %s from %s%s) temp;", showFieldsStr, fieldsStr, tableName, filterSql)
	} else {
		sql = fmt.Sprintf("SELECT %s from (SELECT %s from %s%s) temp;", showFieldsStr, fieldsStr, tableName, filterSql)
	}
	// 获取数据
	var results []map[string]interface{}
	if err := db.Raw(sql).Scan(&results).Error; err != nil {
		global.GLog.Error("ReportExport", zap.Any("err", err))
	}
	// 根据数据库字段获取中文字段名
	var titleDict = make(map[string]string, 0)
	for _, item := range titleList {
		titleDict[fmt.Sprintf("%v", item["key"])] = fmt.Sprintf("%v", item["val"])
	}
	// 新建xlsx文件
	file := xlsx.NewFile()
	sheet, _ := file.AddSheet("sheet")
	// 设置表格头headers
	row := sheet.AddRow()
	row.AddCell().Value = "CMOS服务器名称"
	for _, item := range fieldList {
		row.AddCell().Value = titleDict[item]
	}
	// 写入数据
	cmosName := GetFullDomainName() // CMOS服务器名称
	for _, result := range results {
		row = sheet.AddRow()
		row.AddCell().Value = cmosName
		for _, item := range fieldList {
			row.AddCell().Value = Interface2String(result[item])
		}
	}
	// 设置Response Content信息
	contentDisposition := fmt.Sprintf(`attachment; filename="%s-report.xlsx"`, ServiceTypeMap[rule.ServiceType]) // 文件名
	c.Writer.Header().Add("Content-Disposition", contentDisposition)
	c.Writer.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// 回写到 web 流媒体 形成下载
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		global.GLog.Error("ReportExport", zap.Any("err", err))
	}
	r := bytes.NewReader(buffer.Bytes())
	http.ServeContent(c.Writer, c.Request, "", time.Now(), r)
}

// ExportComputerUpdatesStatusData 第五个报表导出
func ExportComputerUpdatesStatusData(c *gin.Context, rule *model.CustomReportRules, userID int, userEmail string) {
	db := global.GDb

	// 根据数据库字段获取中文字段名
	var titleDict = make(map[string]string)
	for _, item := range ComputerRevisionInstallStatsChoices {
		titleDict[fmt.Sprintf("%v", item["key"])] = fmt.Sprintf("%v", item["val"])
	}
	// 该报表为实时获取 通过列表页接口获取所有数据
	// fields_list为表格要展示字段
	var screen map[string]map[string][]string
	_ = json.Unmarshal([]byte(rule.Screen), &screen)
	var show []string
	_ = json.Unmarshal([]byte(rule.Show), &show)
	data, fieldsList, _ := GetComputerUpdatesStatusData(int(rule.ID), rule.TargetGroupID, rule.ActionID, screen, show, 0, 0, true)
	// 获取计算机组名称（创建报表时单选且必选）
	var targetGroup model.ComputerTargetGroup
	if err := db.Select("TargetGroupName").Where("TargetGroupID = ?", rule.TargetGroupID).First(&targetGroup).Error; err != nil {
		global.GLog.Error("ExportComputerUpdatesStatusData", zap.Any("err", err))
	}
	targetGroupName := targetGroup.TargetGroupName
	if rule.TargetGroupID == utils.UUIDGroupUnassigned {
		targetGroupName = "待分配组"
	}
	// 更新审批状态(创建报表时为单选且必选)
	approveStatus := ApproveStatusFilterMap[strconv.Itoa(rule.ActionID)]
	// sheet表字段
	var title []string
	if len(fieldsList) > 0 {
		title = []string{"CMOS服务器名称", "计算机组名称", "更新审批状态"}
		for _, item := range fieldsList {
			title = append(title, titleDict[item])
		}
	}

	// 新建xlsx文件
	file := xlsx.NewFile()
	sheet, _ := file.AddSheet("sheet")
	// 设置表格头headers
	row := sheet.AddRow()
	for _, item := range title {
		row.AddCell().Value = item
	}
	// 写入数据
	cmosName := GetFullDomainName() // CMOS服务器名称
	for _, result := range data {
		row = sheet.AddRow()
		row.AddCell().Value = cmosName
		row.AddCell().Value = targetGroupName
		row.AddCell().Value = approveStatus
		for _, item := range fieldsList {
			row.AddCell().Value = Interface2String(result[item])
		}
	}
	// 设置Response Content信息
	contentDisposition := fmt.Sprintf(`attachment; filename="%s-report.xlsx"`, ServiceTypeMap[rule.ServiceType]) // 文件名
	c.Writer.Header().Add("Content-Disposition", contentDisposition)
	c.Writer.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// 回写到 web 流媒体 形成下载
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		global.GLog.Error("ExportComputerUpdatesStatusData", zap.Any("err", err))
	}
	r := bytes.NewReader(buffer.Bytes())
	// 记录操作日志
	operateDesc := fmt.Sprintf("导出报表【%s】", rule.Name)
	GenOperateRecord(OperateAction.ExportReport, operateDesc, ResultStatus.SUCCESS, userID, userEmail)
	http.ServeContent(c.Writer, c.Request, "", time.Now(), r)
}

// GetReportScreenUpdates 第五个报表根据条件筛选对应更新
func GetReportScreenUpdates(c *gin.Context) (int, string, []map[string]interface{}, int) {
	db := global.GDb

	var params struct {
		TargetGroupID string                         `json:"TargetGroupID"`
		Screen        map[string]map[string][]string `json:"screen"`
		ActionID      string                         `json:"ActionID"`
		Page          map[string]int                 `json:"page"`
	}
	err := c.BindJSON(&params)
	if err != nil {
		global.GLog.Error("GetReportScreenUpdates", zap.Any("err", err))
		return response.FAILED, "参数获取失败", nil, 0
	}

	pageNum := params.Page["number"]
	pageSize := params.Page["size"]

	filterGroups := []string{params.TargetGroupID, utils.UUIDAllComputer}
	action, _ := strconv.Atoi(params.ActionID)
	actionID := action
	filterActions := []int{actionID}
	// 未审批的状态特殊，为2，3，前端目前只给后端3，2需要自己添加
	if actionID == utils.BlockAction {
		filterActions = append(filterActions, utils.PreDeploymentCheckAction)
	}
	// 获取某个组下面满足审批状态的所有RevisionID
	var deploymentRevisions []int
	db.Model(&model.Deployment{}).Where("TargetGroupID in (?) and ActionID in (?)", filterGroups, filterActions).Pluck("RevisionID", &deploymentRevisions).Distinct("RevisionID")

	if actionID == utils.BlockAction {
		// 当筛选未审批的时候，因为筛选了所有计算机组，所以可能会导致审批到安装/卸载的更新被筛选出来，这里要先查出来，之后做差集
		var needRemoveRevisions []int
		db.Model(&model.Deployment{}).Where("TargetGroupID = ? and ActionID in (?)", params.TargetGroupID, []int{utils.InstallAction, utils.UninstallAction}).Pluck("RevisionID", &needRemoveRevisions).Distinct("RevisionID")

		deploymentRevisions = goset.IntMinus(deploymentRevisions, needRemoveRevisions).List()
	} else if actionID == utils.DeclineAction { // 拒绝的时候可能多筛选出bundle的revision,需要剔除
		tmp, _ := service.GenBundleRevisionIds(utils.IntMapStr(deploymentRevisions))
		bundleRevisions := utils.StrMapInt(tmp)
		intersection := goset.IntIntersect(deploymentRevisions, bundleRevisions).List()
		deploymentRevisions = goset.IntMinus(deploymentRevisions, intersection).List()
	}
	if len(deploymentRevisions) == 0 {
		return response.SUCCESS, response.ErrorMsg[response.SUCCESS], nil, 0
	}

	statusFilterSql := ""
	revisionFilterSql := fmt.Sprintf("SELECT RevisionID FROM revision WHERE UpdateType NOT IN ('Category', 'Detectoid') AND ProductRevisionID IS NOT NULL AND ClassificationRevisionID IS NOT NULL AND RevisionID IN (%s)", GenSqlStrInt(deploymentRevisions))
	limitSql := fmt.Sprintf(" LIMIT %s,%s", strconv.Itoa((pageNum-1)*pageSize), strconv.Itoa(pageSize))

	search := params.Screen
	if len(search) > 0 {
		if inSearch, ok := search["in"]; ok {
			for key, val := range inSearch {
				if key == "Status" {
					statusFilterSql += fmt.Sprintf(" AND %s IN (%s)", key, GenSqlStrString(val))
				} else {
					revisionFilterSql += fmt.Sprintf(" AND %s IN (%s)", key, GenSqlStrString(val))
				}
			}
		}
		if rangeSearch, ok := search["range"]; ok {
			for key, val := range rangeSearch {
				if key == "CreationDate" {
					// 开始时间减8小时
					tmp, _ := time.Parse("2006-01-02 15:04:05", val[0])
					val[0] = tmp.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
					// 结束时间减8小时
					tmp1, _ := time.Parse("2006-01-02 15:04:05", val[1])
					val[1] = tmp1.Add(time.Hour * -8).Format("2006-01-02 15:04:05")
				}
				revisionFilterSql += fmt.Sprintf(" AND %s >= '%s' AND %s <= '%s'", key, val[0], key, val[1])
			}
		}
	}
	// 获取更新sql
	var data []map[string]interface{}
	dataSql := fmt.Sprintf("SELECT RevisionID,(SELECT Title FROM property WHERE RevisionID = r.RevisionID AND Language IN ('zh-cn', 'en') ORDER BY Language DESC LIMIT 1) AS Title FROM (%s) r %s;", revisionFilterSql, limitSql)
	db.Raw(dataSql).Scan(&data)
	// 计算总数
	var count int
	countSql := fmt.Sprintf("SELECT COUNT(1) FROM (%s) r;", revisionFilterSql)
	db.Raw(countSql).Scan(&count)
	return response.SUCCESS, response.ErrorMsg[response.SUCCESS], data, count
}

func DataListToResultDict(data []model.ComputerLiveStats, period int, useMondayInLiveStats bool) map[string]map[string]int {
	// 将数据库查询返回的data list转换为前端需要的result字典
	// return: result字典var result map[string]map[string]int
	result := make(map[string]map[string]int)
	var yearMonth string
	var yearWeek string
	for _, value := range data {
		if _, ok := result[value.OSversion]; !ok {
			result[value.OSversion] = map[string]int{}
		}
		if period == ComputerLiveStatsCycle.MONTH {
			if value.Month%10 == value.Month {
				yearMonth = strconv.Itoa(value.Year) + "-0" + strconv.Itoa(value.Month)
			} else {
				yearMonth = strconv.Itoa(value.Year) + "-" + strconv.Itoa(value.Month)
			}
			result[value.OSversion][yearMonth] = value.ComputerCount
		} else if period == ComputerLiveStatsCycle.WEEK {
			if value.Week%10 == value.Week {
				yearWeek = strconv.Itoa(value.Year) + "-0" + strconv.Itoa(value.Week)
			} else {
				yearWeek = strconv.Itoa(value.Year) + "-" + strconv.Itoa(value.Week)
			}
			if !useMondayInLiveStats {
				result[value.OSversion][yearWeek] = value.ComputerCount
			} else {
				mondayDate := GetMondayDate(yearWeek)
				result[value.OSversion][mondayDate.Format("2006-01-02")] = value.ComputerCount
			}
		} else if period == ComputerLiveStatsCycle.DAY {
			result[value.OSversion][value.Date.Format("2006-01-02")] = value.ComputerCount
		}
	}
	return result
}

func AddDailyZeroData(start string, end string, result map[string]map[string]int) map[string]map[string]int {
	// 将零数据补全至日活的查找数据result中
	// return: result字典
	startParse, _ := time.Parse("2006-01-02", start)
	endParse, _ := time.Parse("2006-01-02", end)
	for _, v := range result {
		dateStart := startParse
		for dateStart.Before(endParse) || dateStart.Equal(endParse) {
			dateStartStr := dateStart.Format("2006-01-02")
			if _, ok := v[dateStartStr]; !ok {
				v[dateStartStr] = 0
			}
			dateStart = dateStart.Add(time.Hour * 24)
		}
	}
	return result
}

func AddWeeklyZeroData(start string, end string, weekStart int, weekEnd int, result map[string]map[string]int, useMondayInLiveStats bool) map[string]map[string]int {
	// 将零数据补全至周活的查找数据result中
	// return: result字典
	yearStart := start[0:4]
	yearEnd := end[0:4]
	startYearLastDay, _ := time.Parse("2006-01-02", yearStart+"-12-31")
	_, startYearLastWeekNum := startYearLastDay.ISOWeek()
	var yearWeek string
	if yearStart != yearEnd {
		for _, v := range result {
			dateStart := weekStart
			for dateStart <= startYearLastWeekNum {
				if dateStart%10 == dateStart {
					yearWeek = yearStart + "-0" + strconv.Itoa(dateStart)
				} else {
					yearWeek = yearStart + "-" + strconv.Itoa(dateStart)
				}
				if useMondayInLiveStats {
					mondayDate := GetMondayDate(yearWeek).Format("2006-01-02")
					if _, ok := v[mondayDate]; !ok {
						if _, ok2 := v[yearWeek]; !ok2 {
							v[mondayDate] = 0
						} else {
							v[mondayDate] = v[yearWeek]
							delete(v, yearWeek)
						}
					}
				} else {
					if _, ok := v[yearWeek]; !ok {
						v[yearWeek] = 0
					}
				}
				dateStart += 1
			}
			dateStart = 1
			for dateStart <= weekEnd {
				if dateStart%10 == dateStart {
					yearWeek = yearEnd + "-0" + strconv.Itoa(dateStart)
				} else {
					yearWeek = yearEnd + "-" + strconv.Itoa(dateStart)
				}
				if useMondayInLiveStats {
					mondayDate := GetMondayDate(yearWeek).Format("2006-01-02")
					if _, ok := v[mondayDate]; !ok {
						v[mondayDate] = 0
					}
				} else {
					if _, ok := v[yearWeek]; !ok {
						v[yearWeek] = 0
					}
				}
				dateStart += 1
			}
		}
	} else {
		for _, v := range result {
			dateStart := weekStart
			for dateStart <= weekEnd {
				if dateStart%10 == dateStart {
					yearWeek = yearEnd + "-0" + strconv.Itoa(dateStart)
				} else {
					yearWeek = yearEnd + "-" + strconv.Itoa(dateStart)
				}
				if useMondayInLiveStats {
					mondayDate := GetMondayDate(yearWeek).Format("2006-01-02")
					if _, ok := v[mondayDate]; !ok {
						v[mondayDate] = 0
					}
				} else {
					if _, ok := v[yearWeek]; !ok {
						v[yearWeek] = 0
					}
				}
				dateStart += 1
			}
		}
	}
	return result
}

func AddMonthlyZeroData(start string, end string, result map[string]map[string]int) map[string]map[string]int {
	// 将零数据补全至月活的查找数据result中
	// return: result字典
	yearStart := start[0:4]
	yearEnd := end[0:4]
	monthStart, _ := strconv.Atoi(start[5:7])
	monthEnd, _ := strconv.Atoi(end[5:7])
	december := 12
	if yearStart != yearEnd {
		for _, v := range result {
			dateStart := monthStart
			for dateStart <= december {
				var yearMonth string
				if dateStart%10 == dateStart {
					yearMonth = yearStart + "-0" + strconv.Itoa(dateStart)
				} else {
					yearMonth = yearStart + "-" + strconv.Itoa(dateStart)
				}
				if _, ok := v[yearMonth]; !ok {
					v[yearMonth] = 0
				}
				dateStart += 1
			}
			dateStart = 1
			for dateStart <= monthEnd {
				var yearMonth string
				if dateStart%10 == dateStart {
					yearMonth = yearEnd + "-0" + strconv.Itoa(dateStart)
				} else {
					yearMonth = yearEnd + "-" + strconv.Itoa(dateStart)
				}
				if _, ok := v[yearMonth]; !ok {
					v[yearMonth] = 0
				}
				dateStart += 1
			}
		}
	} else {
		for _, v := range result {
			dateStart := monthStart
			for dateStart <= monthEnd {
				var yearMonth string
				if dateStart%10 == dateStart {
					yearMonth = yearStart + "-0" + strconv.Itoa(dateStart)
				} else {
					yearMonth = yearStart + "-" + strconv.Itoa(dateStart)
				}
				if _, ok := v[yearMonth]; !ok {
					v[yearMonth] = 0
				}
				dateStart += 1
			}
		}
	}
	return result
}

func GetMondayDate(yearWeek string) time.Time {
	//计算并返回年份加周次(eg.‘2020-37’)该周的周一的日期
	//return: date type 2020-09-07
	yearNum := yearWeek[0:4]         // 取到年份
	weekNum := yearWeek[5:7]         // 取到周
	strYearStart := yearNum + "0101" // 当年第一天
	yearStartDate, _ := time.Parse("20060102", strYearStart)
	yearStartWeekday := yearStartDate.Weekday().String()
	yearStartYear, _ := yearStartDate.ISOWeek()
	var dayDelta int
	weekdayMap := map[string]int{
		"Monday": 1, "Tuesday": 2, "Wednesday": 3, "Thursday": 4, "Friday": 5, "Saturday": 6, "Sunday": 7,
	}
	yearNumInt, _ := strconv.Atoi(yearNum)
	weekNumInt, _ := strconv.Atoi(weekNum)
	if yearStartYear < yearNumInt {
		dayDelta = (8 - weekdayMap[yearStartWeekday]) + (weekNumInt-1)*7
	} else {
		dayDelta = (8 - weekdayMap[yearStartWeekday]) + (weekNumInt-2)*7
	}
	d, _ := time.ParseDuration("24h")
	totalDelta := time.Duration(dayDelta) * d
	mondayDate := yearStartDate.Add(totalDelta)
	return mondayDate
}

func SortResultDictKeys(value map[string]int) []string {
	// 将每个OS的value_dict中的日期进行排序并返回已排序的value_dict字典
	// return: 排序过的value_dict字典
	var sortedKeys []string
	for key := range value {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}
