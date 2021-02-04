package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goblog/core/global"
	"goblog/core/response"
	"goblog/modules/entity"
	"goblog/modules/model"
	"goblog/service/admin"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// report mapping
func GetReportMapping(c *gin.Context) {
	productTitle, classificationTitle := admin.GetProductClassification()
	groupDict := admin.GetAllGroupsDict()

	// 更新筛选字段
	admin.UpdateTitleChoices[5]["filter"] = productTitle
	admin.UpdateTitleChoices[6]["filter"] = classificationTitle
	admin.UpdateTitleChoices[7]["filter"] = groupDict
	// 计算机更新安装状态筛选字段
	admin.ComputerRevisionInstallStatsChoices[0]["filter"] = groupDict
	admin.ComputerRevisionInstallStatsChoices[7]["filter"] = productTitle
	admin.ComputerRevisionInstallStatsChoices[8]["filter"] = classificationTitle

	var data = map[string]interface{}{
		"ComputerUpdatesStatus":        admin.ComputerRevisionInstallStatsChoices,
		"Updates":                      admin.UpdateTitleChoices,
		"Computer":                     admin.ComputerTitleChoices,
		"UpdateInstallationStatistics": admin.UpdateInstallStatsTitleChoices,
		"DSS":                          admin.DssTitleChoices,
	}
	resp := entity.NewResponse().SetData(data)
	response.Ok(resp, c)
	return
}

func ReportStaticsCi(c *gin.Context) {
	resp := entity.NewResponse()
	admin.StatusStatsOfPerComputer()
	admin.StatusStatsOfPerUpdate()
	admin.ComputerReportStats()
	admin.UpdateInstallReportStats()
	response.Ok(resp, c)
	return
}

func ReportStaticsLiveCi(c *gin.Context) {
	resp := entity.NewResponse()
	param := c.DefaultQuery("param", "daily")

	if param == "daily" {
		admin.ComputerDailyLiveStats()
	} else if param == "weekly" {
		admin.ComputerWeeklyLiveStats()
	} else if param == "monthly" {
		admin.ComputerMonthlyLiveStats()
	} else {
		resp.SetCode(response.ParamError)
		response.Fail(resp, c)
		return
	}
	response.Ok(resp, c)
	return
}

// GetReportRules 获取自定义报表规则列表
func GetReportRules(c *gin.Context) {
	resp := entity.NewResponse()
	pageNum, pageSize := admin.GetPageParams(c)
	userID, _ := admin.GetUserInfo(c)

	newRuleList, count := admin.ReportRuleList(userID, pageNum, pageSize)

	resp.SetData(newRuleList).SetMeta(map[string]int64{"count": count})
	response.Ok(resp, c)
	return
}

// AddReportRules 新建自定义报表规则
func AddReportRules(c *gin.Context) {
	resp := entity.NewResponse()

	code, msg, ruleID := admin.ReportRuleCreate(c)
	if code != response.SUCCESS {
		resp.SetMsg(msg)
		response.Fail(resp, c)
		return
	} else {
		resp.SetData(map[string]int{"rule_id": ruleID})
		response.Ok(resp, c)
		return
	}
}

// EditReportRules 编辑自定义报表规则
func EditReportRules(c *gin.Context) {
	resp := entity.NewResponse()

	code, msg, ruleID := admin.ReportRuleEdit(c)
	if code != response.SUCCESS {
		resp.SetMsg(msg)
		response.Fail(resp, c)
		return
	} else {
		resp.SetData(map[string]int{"rule_id": ruleID})
		response.Ok(resp, c)
		return
	}
}

// DeleteReportRules 删除自定义报表规则
func DeleteReportRules(c *gin.Context) {
	resp := entity.NewResponse()

	code, msg := admin.ReportRuleDel(c)
	if code != response.SUCCESS {
		resp.SetMsg(msg)
		response.Fail(resp, c)
		return
	} else {
		response.Ok(resp, c)
		return
	}
}

// ReportDetail 报表详情
func ReportDetail(c *gin.Context) {
	resp := entity.NewResponse()
	pageNum, pageSize := admin.GetPageParams(c)
	userID, _ := admin.GetUserInfo(c)
	ruleID := c.Param("ruleID")
	var err error

	db := global.GDb
	var rule model.CustomReportRules
	if err = db.Where("id = ? or (user_id = ? and built_in is true)", ruleID, userID).First(&rule).Error; err == gorm.ErrRecordNotFound {
		resp.SetCode(response.IDNotFound)
		response.Fail(resp, c)
		return
	}

	var show []string
	_ = json.Unmarshal([]byte(rule.Show), &show)
	var screen map[string]map[string][]string
	_ = json.Unmarshal([]byte(rule.Screen), &screen)
	meta := make(map[string]interface{})
	meta["count"] = 0
	meta["last_stats_time"] = ""
	ruleInfo := make(map[string]interface{})
	meta["rule"] = ruleInfo
	ruleInfo["name"] = rule.Name
	ruleInfo["service_type"] = rule.ServiceType
	ruleInfo["show"] = show
	ruleInfo["screens"] = screen
	ruleInfo["desc"] = rule.Desc
	ruleInfo["built_in"] = rule.BuiltIn
	ruleInfo["ActionID"] = strconv.Itoa(rule.ActionID)
	ruleInfo["TargetGroupID"] = rule.TargetGroupID
	ruleInfo["is_valid"] = rule.IsValid
	ruleInfo["revision_ids"] = []int{}
	// 第五个报表获取相关联数据
	if rule.ServiceType == admin.ServiceType.ComputerUpdatesStatus {
		var revisionIDs []int
		db.Model(&model.ComputerUpdateRelation{}).Where("rule_id = ? and is_valid is true", rule.ID).Pluck("revision_id", &revisionIDs)
		ruleInfo["revision_ids"] = revisionIDs
	}
	// 失效情况下直接返回空列表，但前端需要meta数据用于编辑
	if !rule.IsValid {
		meta["last_stats_time"] = time.Now().Format("2006-01-02 15:04:05")
		resp.SetMeta(meta)
		response.Ok(resp, c)
		return
	}

	var results []map[string]interface{}
	var tableName string

	if rule.ServiceType == admin.ServiceType.Updates { // 更新
		tableName = model.RevisionStatement{}.TableName()
	} else if rule.ServiceType == admin.ServiceType.Computer { // 计算机
		tableName = model.ComputerStatement{}.TableName()
	} else if rule.ServiceType == admin.ServiceType.UpdateInstallStats { // 更新安装统计
		tableName = model.RevisionInstallStatistics{}.TableName()
	} else if rule.ServiceType == admin.ServiceType.DSS { // 下游服务器
		tableName = model.DSSStatement{}.TableName()
	} else if rule.ServiceType == admin.ServiceType.ComputerUpdatesStatus { // 计算机更新安装状态
		data, _, totalCount := admin.GetComputerUpdatesStatusData(int(rule.ID), rule.TargetGroupID, rule.ActionID, screen, show, pageNum, pageSize, false)
		meta["count"] = totalCount
		meta["last_stats_time"] = time.Now().Format("2006-01-02 15:04:05")
		resp.SetMeta(meta).SetData(data)
		response.Ok(resp, c)
		return
	} else {
		resp.SetCode(response.IDNotFound)
		response.Fail(resp, c)
		return
	}
	// 获取sql，count，时间
	sql, count, statsTime := admin.ReportCommonSql(tableName, screen, show, pageNum, pageSize)
	// 获取数据
	if err = db.Raw(sql).Scan(&results).Error; err != nil {
		global.GLog.Error("ReportDetail", zap.Any("err", err))
	}
	meta["count"] = count
	meta["last_stats_time"] = statsTime
	resp.SetMeta(meta).SetData(results)
	response.Ok(resp, c)
	return
}

// ReportExport 报表导出
func ReportExport(c *gin.Context) {
	resp := entity.NewResponse()
	userID, userEmail := admin.GetUserInfo(c)
	ruleID := c.Param("ruleID")
	var err error

	db := global.GDb
	var rule model.CustomReportRules
	if err = db.Where("id = ? or (user_id = ? and built_in is true)", ruleID, userID).First(&rule).Error; err == gorm.ErrRecordNotFound {
		resp.SetCode(response.IDNotFound)
		response.Fail(resp, c)
		return
	}

	var tableName = ""
	var titleList = make([]map[string]interface{}, 0)

	if rule.ServiceType == admin.ServiceType.Updates { // 更新
		tableName = model.RevisionStatement{}.TableName()
		titleList = admin.UpdateTitleChoices
	} else if rule.ServiceType == admin.ServiceType.Computer { // 计算机
		tableName = model.ComputerStatement{}.TableName()
		titleList = admin.ComputerTitleChoices
	} else if rule.ServiceType == admin.ServiceType.UpdateInstallStats { // 更新安装统计
		tableName = model.RevisionInstallStatistics{}.TableName()
		titleList = admin.UpdateInstallStatsTitleChoices
	} else if rule.ServiceType == admin.ServiceType.DSS { // 下游服务器
		tableName = model.DSSStatement{}.TableName()
		titleList = admin.DssTitleChoices
	} else if rule.ServiceType == admin.ServiceType.ComputerUpdatesStatus { // 计算机更新安装状态
		admin.ExportComputerUpdatesStatusData(c, &rule, userID, userEmail)
		return
	} else {
		resp.SetCode(response.IDNotFound)
		response.Fail(resp, c)
		return
	}
	admin.ReportExportCommon(c, &rule, tableName, titleList)
	return
}

// GetScreenUpdates 获取计算机更新安装状态报表筛选前的更新列表
func GetScreenUpdates(c *gin.Context) {
	resp := entity.NewResponse()

	code, msg, data, count := admin.GetReportScreenUpdates(c)

	if code != response.SUCCESS {
		resp.SetMsg(msg)
		response.Fail(resp, c)
		return
	} else {
		resp.SetData(data).SetMeta(map[string]int{"count": count})
		response.Ok(resp, c)
		return
	}
}

// AddOperate 活跃度统计导出图片时添加操作记录
func LiveExportAddOperate(c *gin.Context) {
	resp := entity.NewResponse()

	userID, userEmail := admin.GetUserInfo(c)
	operateDesc := "导出图片"
	admin.GenOperateRecord(admin.OperateAction.ExportImage, operateDesc, admin.ResultStatus.SUCCESS, userID, userEmail)

	resp.SetMsg("生成导出报表图片操作记录成功")
	response.Ok(resp, c)
}

// 活跃度统计折线图
func Live(c *gin.Context) {
	// 活跃度折线图
	// return: result字典和meta报表统计时间
	// 获取参数
	resp := entity.NewResponse()
	period, _ := strconv.Atoi(c.DefaultQuery("period", strconv.Itoa(admin.ComputerLiveStatsCycle.DAY)))
	timeNewJson := c.Query("time")
	osListJson := c.Query("os")
	var timeNewSlice []string
	if err := json.Unmarshal([]byte(timeNewJson), &timeNewSlice); err != nil {
		response.Fail(resp, c)
		return
	}
	var osListSlice []string
	if err := json.Unmarshal([]byte(osListJson), &osListSlice); err != nil {
		response.Fail(resp, c)
		return
	}
	start := timeNewSlice[0]
	end := timeNewSlice[1]

	var data []model.ComputerLiveStats
	err := global.GDb.Model(model.ComputerLiveStats{}).Select("date", "computer_count", "os_version", "week", "stats_time", "year", "month").
		Where("date >= ? AND date <= ? AND cycle = ? AND os_version IN (?)", start, end, period, osListSlice).Find(&data).Error
	if err != nil {
		global.GLog.Error("LiveDatabase", zap.Any("err", err))
	}

	var result map[string]map[string]int
	if data != nil {
		result = admin.DataListToResultDict(data, period, true)
		if period == admin.ComputerLiveStatsCycle.DAY {
			result = admin.AddDailyZeroData(start, end, result)
		} else if period == admin.ComputerLiveStatsCycle.WEEK {
			weekStartTime, _ := time.Parse("2006-01-02", start)
			_, weekStartNum := weekStartTime.ISOWeek()
			weekEndTime, _ := time.Parse("2006-01-02", end)
			yearEndNum, weekEndNum := weekEndTime.ISOWeek()
			result = admin.AddWeeklyZeroData(start, strconv.Itoa(yearEndNum), weekStartNum, weekEndNum, result, true)
		} else if period == admin.ComputerLiveStatsCycle.MONTH {
			result = admin.AddMonthlyZeroData(start, end, result)
		}
	}

	meta := make(map[string]string)
	if data != nil {
		statsTime := data[len(data)-1].StatsTime
		statsTimeDelta := statsTime.Add(time.Hour * 8)
		meta["last_stats_time"] = statsTimeDelta.Format("2006-01-02 15:04:05")
	} else {
		meta["last_stats_time"] = ""
	}
	resp.SetData(result).SetMeta(meta)
	response.Ok(resp, c)
	return
}

// 活跃度统计Excel导出
func LiveExport(c *gin.Context) {
	// 活跃度excel导出
	// :return: 日活/周活/月活 Excel
	// 获取参数
	resp := entity.NewResponse()
	period, _ := strconv.Atoi(c.DefaultQuery("period", strconv.Itoa(admin.ComputerLiveStatsCycle.DAY)))
	timeNewJson := c.Query("time")
	osListJson := c.Query("os")
	var timeNewSlice []string
	if err := json.Unmarshal([]byte(timeNewJson), &timeNewSlice); err != nil {
		response.Fail(resp, c)
		return
	}
	var osListSlice []string
	if err := json.Unmarshal([]byte(osListJson), &osListSlice); err != nil {
		response.Fail(resp, c)
		return
	}
	start := timeNewSlice[0]
	end := timeNewSlice[1]

	var data []model.ComputerLiveStats
	err := global.GDb.Model(model.ComputerLiveStats{}).Select("date", "computer_count", "os_version", "week", "stats_time", "year", "month").
		Where("date >= ? AND date <= ? AND cycle = ? AND os_version IN (?)", start, end, period, osListSlice).Find(&data).Error
	if err != nil {
		global.GLog.Error("LiveExportDatabase", zap.Any("err", err))
	}

	var result map[string]map[string]int
	var useMondayInLiveStats bool
	if data != nil {
		result = admin.DataListToResultDict(data, period, useMondayInLiveStats)
	}
	// 写入Excel文件
	var filename string
	var title []string
	fullDomainName := admin.GetFullDomainName()
	file := xlsx.NewFile()
	sheet, _ := file.AddSheet("Sheet1")
	if period == admin.ComputerLiveStatsCycle.DAY || period == admin.ComputerLiveStatsCycle.MONTH {
		if period == admin.ComputerLiveStatsCycle.DAY {
			filename = "daily-report" // 文件名
		} else if period == admin.ComputerLiveStatsCycle.MONTH {
			filename = "monthly-report"
		}
		title = []string{"服务器名称", "日期", "活跃量", "操作系统版本"} // sheet表字段
		if data != nil {
			if period == admin.ComputerLiveStatsCycle.DAY {
				result = admin.AddDailyZeroData(start, end, result)
			} else if period == admin.ComputerLiveStatsCycle.MONTH {
				result = admin.AddMonthlyZeroData(start, end, result)
			}
			row := sheet.AddRow()
			for _, item := range title {
				row.AddCell().Value = item
			}
			var sortedDictKeys []string
			for _, v := range result {
				sortedDictKeys = admin.SortResultDictKeys(v)
			}
			for k, v := range result {
				for _, key := range sortedDictKeys {
					row := sheet.AddRow()
					row.AddCell().Value = fullDomainName
					row.AddCell().Value = key
					row.AddCell().Value = strconv.Itoa(v[key])
					row.AddCell().Value = k
				}
			}
		}
	} else if period == admin.ComputerLiveStatsCycle.WEEK {
		filename = "weekly-report"
		title = []string{"服务器名称", "日期", "周次", "活跃量", "操作系统版本"}
		if data != nil {
			weekStartTime, _ := time.Parse("2006-01-02", start)
			_, weekStartNum := weekStartTime.ISOWeek()
			weekEndTime, _ := time.Parse("2006-01-02", end)
			yearEndNum, weekEndNum := weekEndTime.ISOWeek()
			result = admin.AddWeeklyZeroData(start, strconv.Itoa(yearEndNum), weekStartNum, weekEndNum, result, false)
			row := sheet.AddRow()
			for _, item := range title {
				row.AddCell().Value = item
			}
			var sortedDictKeys []string
			for _, v := range result {
				sortedDictKeys = admin.SortResultDictKeys(v)
			}
			for k, v := range result {
				for _, key := range sortedDictKeys {
					row := sheet.AddRow()
					row.AddCell().Value = fullDomainName
					weekMondayDate := admin.GetMondayDate(key)
					row.AddCell().Value = weekMondayDate.Format("2006-01-02")
					row.AddCell().Value = key
					row.AddCell().Value = strconv.Itoa(v[key])
					row.AddCell().Value = k
				}
			}
		}
	}
	// 设置Response Content信息
	contentDisposition := fmt.Sprintf(`attachment; filename="%s.xlsx"`, filename) // 文件名
	c.Writer.Header().Add("Content-Disposition", contentDisposition)
	c.Writer.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	// 回写到 web 流媒体 形成下载
	var buffer bytes.Buffer
	if err := file.Write(&buffer); err != nil {
		global.GLog.Error("LiveExport", zap.Any("err", err))
	}
	r := bytes.NewReader(buffer.Bytes())
	http.ServeContent(c.Writer, c.Request, "", time.Now(), r)
}
