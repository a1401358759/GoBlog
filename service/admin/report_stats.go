package admin

import (
	"fmt"
	"goblog/core/global"
	"goblog/modules/model"
	"goblog/service"
	"goblog/utils"
	"goblog/utils/component"
	"strconv"
	"time"

	"github.com/jinzhu/now"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// StatusStatsOfPerComputer 单个计算机安装更新状态的统计
func StatusStatsOfPerComputer() {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除现有数据
		if err := tx.Exec("TRUNCATE TABLE computer_summary_for_microsoft_updates;").Error; err != nil {
			return err
		}
		// 聚合查询数据
		var ret []struct{ TargetID, Status, StatusCount int }
		if err := tx.Raw("select TargetID, Status, count(id) as StatusCount from update_status_per_computer group by TargetID, Status;").Scan(&ret).Error; err != nil {
			return err
		}
		// 处理数据格式，避免在循环中查询数据库
		var existsTargetIds []int
		var retDict = make(map[string]int, 0) // {'1_0': 1, '1_2': 1}
		for i := 0; i < len(ret); i++ {
			existsTargetIds = append(existsTargetIds, ret[i].TargetID)
			retDict[strconv.Itoa(ret[i].TargetID)+"_"+strconv.Itoa(ret[i].Status)] = ret[i].StatusCount
		}
		// 获取所有的revision的数量
		revisionsCount := service.GetRevisionsCount()
		var haveUsedTargetDict = make(map[int]bool, 0) // 已经统计过的计算机ID
		var statsList []map[string]interface{}

		for i := 0; i < len(ret); i++ {
			item := ret[i]
			// 已经统计过的计算机不需要再次统计
			if _, ok := haveUsedTargetDict[item.TargetID]; !ok {
				// 各个status对应的数量
				notInstalled := retDict[strconv.Itoa(item.TargetID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.NeedInstall)]                   // 需要安装
				downloaded := retDict[strconv.Itoa(item.TargetID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.DownloadedNotInstalled)]          // 已下载未安装
				installed := retDict[strconv.Itoa(item.TargetID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.Installed)]                        // 已完成安装
				failed := retDict[strconv.Itoa(item.TargetID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.InstallFailed)]                       // 安装失败
				installedPendingReboot := retDict[strconv.Itoa(item.TargetID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.InstalledNeedReboot)] // 安装待重启
				haveUsedTargetDict[item.TargetID] = true

				statsList = append(statsList, map[string]interface{}{
					"version_id":             1,
					"TargetID":               item.TargetID,
					"Unknown":                revisionsCount - notInstalled - downloaded - installed - failed - installedPendingReboot,
					"NotInstalled":           notInstalled,
					"Downloaded":             downloaded,
					"Installed":              installed,
					"Failed":                 failed,
					"InstalledPendingReboot": installedPendingReboot,
					"LastChangeTime":         time.Now().UTC(),
				})
			}
		}
		// 获取所有没有上报状态的计算机
		var notExistsTargetIds []int
		if err := tx.Model(&model.ComputerTarget{}).Where("TargetID not in (?)", existsTargetIds).Select("TargetID").Pluck("TargetID", &notExistsTargetIds).Error; err != nil {
			return err
		}
		for i := 0; i < len(notExistsTargetIds); i++ {
			statsList = append(statsList, map[string]interface{}{
				"version_id":             1,
				"TargetID":               notExistsTargetIds[i],
				"Unknown":                revisionsCount,
				"NotInstalled":           0,
				"Downloaded":             0,
				"Installed":              0,
				"Failed":                 0,
				"InstalledPendingReboot": 0,
				"LastChangeTime":         time.Now().UTC(),
			})
		}
		if err := tx.Model(&model.ComputerSummaryForMicrosoftUpdates{}).Create(&statsList).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		global.GLog.Error("StatusStatsOfPerComputer", zap.Any("err", err))
	}
}

// StatusStatsOfPerUpdate 单个更新被所有计算机安装的状态的统计
func StatusStatsOfPerUpdate() {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除现有数据
		if err := tx.Exec("TRUNCATE TABLE statistics_for_per_update;").Error; err != nil {
			return err
		}
		// 聚合统计数据
		var ret []struct{ RevisionID, Status, StatusCount int }
		if err := tx.Raw("select RevisionID, Status, count(id) as StatusCount from update_status_per_computer group by RevisionID, Status;").Scan(&ret).Error; err != nil {
			return err
		}
		// 处理数据格式，避免在循环中查询数据库
		var existsRevisionIds []int
		var retDict = make(map[string]int, 0) // {'1_0': 1, '1_2': 1}
		for i := 0; i < len(ret); i++ {
			existsRevisionIds = append(existsRevisionIds, ret[i].RevisionID)
			retDict[strconv.Itoa(ret[i].RevisionID)+"_"+strconv.Itoa(ret[i].Status)] = ret[i].StatusCount
		}
		// 获取所有计算机的数量
		var computerCount int64
		if err := tx.Model(&model.ComputerTarget{}).Count(&computerCount).Error; err != nil {
			return err
		}
		var haveUsedRevisionDict = make(map[int]bool, 0) // 已经统计过的RevisionID
		var statsList []map[string]interface{}

		for i := 0; i < len(ret); i++ {
			item := ret[i]
			// 已经统计过的计算机不需要再次统计
			if _, ok := haveUsedRevisionDict[item.RevisionID]; !ok {
				// 各个status对应的数量
				notInstalled := retDict[strconv.Itoa(item.RevisionID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.NeedInstall)]                   // 需要安装
				downloaded := retDict[strconv.Itoa(item.RevisionID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.DownloadedNotInstalled)]          // 已下载未安装
				installed := retDict[strconv.Itoa(item.RevisionID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.Installed)]                        // 已完成安装
				failed := retDict[strconv.Itoa(item.RevisionID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.InstallFailed)]                       // 安装失败
				installedPendingReboot := retDict[strconv.Itoa(item.RevisionID)+"_"+strconv.Itoa(service.ComputerUpdateStatusConfig.InstalledNeedReboot)] // 安装待重启
				haveUsedRevisionDict[item.RevisionID] = true

				statsList = append(statsList, map[string]interface{}{
					"version_id":             1,
					"RevisionID":             item.RevisionID,
					"Unknown":                int(computerCount) - notInstalled - downloaded - installed - failed - installedPendingReboot,
					"NotInstalled":           notInstalled,
					"Downloaded":             downloaded,
					"Installed":              installed,
					"Failed":                 failed,
					"InstalledPendingReboot": installedPendingReboot,
					"LastChangeTime":         time.Now().UTC(),
				})
			}
		}
		// 获取所有没有上报状态的更新
		var notExistsRevisionIDs []int
		if err := tx.Model(&model.Revision{}).Where("UpdateType NOT IN ('Category', 'Detectoid') and ProductRevisionID IS NOT NULL AND ClassificationRevisionID IS NOT NULL and RevisionID not in (?)", existsRevisionIds).Pluck("RevisionID", &notExistsRevisionIDs).Error; err != nil {
			return err
		}
		for i := 0; i < len(notExistsRevisionIDs); i++ {
			statsList = append(statsList, map[string]interface{}{
				"version_id":             1,
				"RevisionID":             notExistsRevisionIDs[i],
				"Unknown":                computerCount,
				"NotInstalled":           0,
				"Downloaded":             0,
				"Installed":              0,
				"Failed":                 0,
				"InstalledPendingReboot": 0,
				"LastChangeTime":         time.Now().UTC(),
			})
		}
		if err := tx.Model(&model.StatisticsForPerUpdate{}).Create(&statsList).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		global.GLog.Error("StatusStatsOfPerUpdate", zap.Any("err", err))
	}
}

// DelReportUselessData 删除reporting发过来的无用的数据，包括BasicData,ExtendedData和MiscData
func DelReportUselessData() {
	db := global.GDb
	if err := db.Where("1 = 1").Delete(&model.ExtendedData{}).Error; err != nil {
		global.GLog.Error("DelReportUselessData", zap.Any("err", err))
	}
	if err := db.Where("1 = 1").Delete(&model.MiscData{}).Error; err != nil {
		global.GLog.Error("DelReportUselessData", zap.Any("err", err))
	}
	if err := db.Where("1 = 1").Delete(&model.BasicData{}).Error; err != nil {
		global.GLog.Error("DelReportUselessData", zap.Any("err", err))
	}
}

// ComputerReportStats 计算机报表异步统计
func ComputerReportStats() {
	db := global.GDb
	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除现有数据
		if err := tx.Exec("TRUNCATE TABLE computer_statement;").Error; err != nil {
			return err
		}
		// 这里不使用ComputerStament结构体的原因是创建时零值默认不更新
		var computers []map[string]interface{}
		// 获取最新数据
		allComputers := GetComputers()
		for i := 0; i < len(allComputers); i++ {
			item := allComputers[i]
			computers = append(computers, map[string]interface{}{
				"FullDomainName":         item.FullDomainName,
				"TargetID":               item.TargetID,
				"LastSyncTime":           item.LastSyncTime,
				"LastReportedStatusTime": item.LastReportedStatusTime,
				"OSVersion":              item.OSVersion,
				"NotInstalled":           item.NotInstalled,
				"Downloaded":             item.Downloaded,
				"InstalledPendingReboot": item.InstalledPendingReboot,
				"Failed":                 item.Failed,
				"Installed":              item.Installed,
				"Unknown":                item.Unknown,
				"IPAddress":              item.IPAddress,
				"ComputerMake":           item.ComputerMake,
				"ComputerModel":          item.ComputerModel,
				"FirmwareVersion":        item.FirmwareVersion,
				"BiosName":               item.BiosName,
				"BiosVersion":            item.BiosVersion,
				"BiosReleaseDate":        item.BiosReleaseDate,
				"OSDescription":          item.OSDescription,
				"OSMajorVersion":         item.OSMajorVersion,
				"OSMinorVersion":         item.OSMinorVersion,
				"OSBuildNumber":          item.OSBuildNumber,
				"MobileOperator":         item.MobileOperator,
				"ClientVersion":          item.ClientVersion,
				"StatsTime":              time.Now().UTC(), // autoCreateTime对map类型的创建不生效，需要显式指定
			})
		}
		if err := tx.Model(&model.ComputerStatement{}).Create(&computers).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		global.GLog.Error("ComputerReportStats", zap.Any("err", err))
	}
}

// UpdateReportStats 更新报表异步统计
func UpdateReportStats() {
	db := global.GDb
	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除现有数据
		if err := tx.Exec("TRUNCATE TABLE revision_statement;").Error; err != nil {
			return err
		}

		allRevisions, revisionsAllComputerDict, revisionIDs := GetDeploymentRevisions()

		var otherRevisionID []struct{ ProductRevisionID, ClassificationRevisionID int }
		if err := db.Model(&model.Revision{}).Select("ProductRevisionID", "ClassificationRevisionID").Where(`UpdateType NOT IN ('Category', 'Detectoid') AND ProductRevisionID IS NOT NULL 
			AND ClassificationRevisionID IS NOT NULL;
		`).Distinct("ProductRevisionID", "ClassificationRevisionID").Scan(&otherRevisionID).Error; err != nil {
			return err
		}
		for i := 0; i < len(otherRevisionID); i++ {
			revisionIDs = append(revisionIDs, otherRevisionID[i].ProductRevisionID)
			revisionIDs = append(revisionIDs, otherRevisionID[i].ClassificationRevisionID)
		}

		revisionTitleDict := GetRevisionTitles(revisionIDs)

		// 这里不使用ComputerStament结构体的原因是创建时零值默认不更新
		var updates []map[string]interface{}
		// 获取最新数据
		for i := 0; i < len(allRevisions); i++ {
			item := allRevisions[i]
			allComputerDict := revisionsAllComputerDict[*item.RevisionID]
			actionID := item.ActionID
			if actionID == nil && allComputerDict["ActionID"] != nil {
				action, _ := allComputerDict["ActionID"].(int)
				actionID = &action
			}
			approveStatus := "未审批"
			if actionID != nil {
				if *actionID == utils.PreDeploymentCheckAction {
					tmp := utils.BlockAction
					actionID = &tmp
				}
				approveStatus = ApproveStatusFilterMap[strconv.Itoa(*actionID)]
			}

			targetGroupName := item.TargetGroupName
			if item.TargetGroupID == utils.UUIDGroupUnassigned {
				targetGroupName = "待分配组"
			}
			// 获取审批人，拒绝的时候需要从所有计算机组集成审批信息
			adminName := item.AdminName
			if adminName == "" && allComputerDict["AdminName"] != nil {
				adminName = allComputerDict["AdminName"].(string)
			}
			// 获取审批时间，拒绝的时候需要从所有计算机组集成审批信息
			lastChangeTime := item.LastChangeTime
			if lastChangeTime == nil && allComputerDict["LastChangeTime"] != nil {
				lastChangeTime = allComputerDict["LastChangeTime"].(*time.Time)
			}

			updates = append(updates, map[string]interface{}{
				"Title":                    revisionTitleDict[*item.RevisionID],
				"RevisionID":               item.RevisionID,
				"UpdateID":                 item.UpdateID,
				"RevisionNumber":           item.RevisionNumber,
				"MsrcSeverity":             MSRCSeverityEnMap[item.MsrcSeverity],
				"ProductRevisionID":        item.ProductRevisionID,
				"ClassificationRevisionID": item.ClassificationRevisionID,
				"ProductTitle":             revisionTitleDict[*item.ProductRevisionID],
				"ClassificationTitle":      revisionTitleDict[*item.ClassificationRevisionID],
				"KBArticleID":              item.KBArticleID,
				"SecurityBulletinID":       item.SecurityBulletinID,
				"CreationDate":             item.CreationDate,
				"ImportedTime":             item.ImportedTime,
				"LastChangedAnchor":        item.LastChangedAnchor,
				"ActionID":                 actionID,
				"ApproveStatus":            approveStatus,
				"DeploymentTime":           lastChangeTime,
				"AdminName":                adminName,
				"TargetGroupID":            item.TargetGroupID,
				"TargetGroupName":          targetGroupName,
				"StatsTime":                time.Now().UTC(), // autoCreateTime对map类型的创建不生效，需要显式指定
			})
		}
		if err := tx.Model(&model.RevisionStatement{}).Create(&updates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		global.GLog.Error("UpdateReportStats", zap.Any("err", err))
	}
}

// DssReportStats Dss下游服务器报表异步统计
func DssReportStats() {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除现有数据
		if err := tx.Exec("TRUNCATE TABLE dss_statement;").Error; err != nil {
			return err
		}
		// 获取所有下游服务器
		var dsses []ReportDssData
		db.Model(&model.Dss{}).Scan(&dsses)
		// 获取最新数据
		var dssStaments []map[string]interface{}
		for i := 0; i < len(dsses); i++ {
			item := dsses[i]
			isReplica := "自治"
			if item.IsReplica {
				isReplica = "副本"
			}
			dssStaments = append(dssStaments, map[string]interface{}{
				"DssID":                  item.ID,
				"LastSyncTime":           item.LastSyncTime,
				"ServerId":               item.ServerId,
				"FullDomainName":         item.FullDomainName,
				"IsReplica":              isReplica,
				"LastRollupTime":         item.LastRollupTime,
				"Version":                item.Version,
				"UpdateCount":            item.UpdateCount,
				"ApprovedUpdateCount":    item.ApprovedUpdateCount,
				"NotApprovedUpdateCount": item.NotApprovedUpdateCount,
				"CriticalOrSecurityUpdatesNotApprovedForInstallCount": item.CriticalOrSecurityUpdatesNotApprovedForInstallCount,
				"ExpiredUpdateCount":                 item.ExpiredUpdateCount,
				"DeclinedUpdateCount":                item.DeclinedUpdateCount,
				"UpdatesUpToDateCount":               item.UpdatesUpToDateCount,
				"UpdatesNeedingFilesCount":           item.UpdatesNeedingFilesCount,
				"CustomComputerTargetGroupCount":     item.CustomComputerTargetGroupCount,
				"ComputerTargetCount":                item.ComputerTargetCount,
				"ComputerTargetsNeedingUpdatesCount": item.ComputerTargetsNeedingUpdatesCount,
				"StatsTime":                          time.Now().UTC(), // autoCreateTime对map类型的创建不生效，需要显式指定
			})
		}
		if err := tx.Model(&model.DSSStatement{}).Create(&dssStaments).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		global.GLog.Error("DssReportStats", zap.Any("err", err))
	}
}

// UpdateInstallReportStats 更新安装统计报表异步统计
func UpdateInstallReportStats() {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		// 删除现有数据
		if err := tx.Exec("TRUNCATE TABLE revision_install_statistics;").Error; err != nil {
			return err
		}
		// 获取更新统计数据
		allRevisions, revisionIDs := GetRevisions()
		revisionTitleDict := GetRevisionTitles(revisionIDs)
		// 获取所有计算机数量
		var computerCount int64
		db.Model(&model.ComputerTarget{}).Count(&computerCount)
		// 获取最新数据
		var staments []map[string]interface{}
		for i := 0; i < len(allRevisions); i++ {
			item := allRevisions[i]
			staments = append(staments, map[string]interface{}{
				"RevisionID":             item.RevisionID,
				"Title":                  revisionTitleDict[item.RevisionID],
				"ComputerCount":          computerCount,
				"Downloaded":             item.Downloaded,
				"InstalledPendingReboot": item.InstalledPendingReboot,
				"Failed":                 item.Failed,
				"Installed":              item.Installed,
				"NotInstalled":           item.NotInstalled,
				"Unknown":                item.Unknown,
				"StatsTime":              time.Now().UTC(), // autoCreateTime对map类型的创建不生效，需要显式指定
			})
		}
		if err := tx.Model(&model.RevisionInstallStatistics{}).Create(&staments).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		global.GLog.Error("UpdateInstallReportStats", zap.Any("err", err))
	}
}

// ComputerDailyLiveStats 每天计算机活跃数量统计
func ComputerDailyLiveStats() {
	db := global.GDb

	today := time.Now().UTC()
	yestoday := today.Add(time.Hour * -24)
	startDate := time.Date(yestoday.Year(), yestoday.Month(), yestoday.Day(), 16, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
	endDate := time.Date(today.Year(), today.Month(), today.Day(), 16, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")

	var count int64
	db.Model(&model.ComputerLiveStats{}).Where("date = ? and cycle = ?", today.Format("2006-01-02"), ComputerLiveStatsCycle.DAY).Count(&count)

	if count == 0 {
		var dailyInfo []LiveStats
		sql := fmt.Sprintf("select OSVersion, COUNT(TargetID) as Count from computer_target where LastSyncTime >= '%s' and LastSyncTime < '%s' group by OSVersion;", startDate, endDate)
		if err := db.Raw(sql).Scan(&dailyInfo).Error; err != nil {
			global.GLog.Error("ComputerDailyLiveStats", zap.Any("err", err))
		}
		var dailyList []map[string]interface{}
		for i := 0; i < len(dailyInfo); i++ {
			dailyList = append(dailyList, map[string]interface{}{
				"date":           today,
				"computer_count": dailyInfo[i].Count,
				"cycle":          ComputerLiveStatsCycle.DAY,
				"year":           today.Year(),
				"os_version":     dailyInfo[i].OSVersion,
				"stats_time":     today,
				"created_time":   today,
				"last_update":    today,
			})
		}
		if err := db.Model(&model.ComputerLiveStats{}).Create(&dailyList).Error; err != nil {
			global.GLog.Error("ComputerDailyLiveStats", zap.Any("err", err))
		}
	}
}

func ComputerWeeklyLiveStats() {
	/*
		每周计算机活跃数量统计
		如果这周还未到周日，则只更新这周的周活
		如果已经进入新的一周了，则创建新一周的周活
	*/
	db := global.GDb

	today := time.Now().UTC()
	nowYear, nowWeek := today.ISOWeek()
	sunday := now.BeginningOfWeek()
	nextSunday := sunday.AddDate(0, 0, 7)

	startDate := time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 16, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
	endDate := time.Date(nextSunday.Year(), nextSunday.Month(), nextSunday.Day(), 16, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")

	var WeeklyInfo []LiveStats
	sql := fmt.Sprintf("select OSVersion, COUNT(TargetID) as Count from computer_target where LastSyncTime >= '%s' and LastSyncTime < '%s' group by OSVersion;", startDate, endDate)
	if err := db.Raw(sql).Scan(&WeeklyInfo).Error; err != nil {
		global.GLog.Error("ComputerWeeklyLiveStats", zap.Any("err", err))
	}
	for i := 0; i < len(WeeklyInfo); i++ {
		info := map[string]interface{}{
			"date":           today,
			"computer_count": WeeklyInfo[i].Count,
			"stats_time":     today,
			"created_time":   today,
			"last_update":    today,
		}
		var stats model.ComputerLiveStats
		if err := db.Where(model.ComputerLiveStats{Year: nowYear, Week: nowWeek, Cycle: ComputerLiveStatsCycle.WEEK, OSversion: WeeklyInfo[i].OSVersion}).Assign(info).FirstOrCreate(&stats).Error; err != nil {
			global.GLog.Error("ComputerWeeklyLiveStats", zap.Any("err", err))
		}
	}
}

// ComputerMonthlyLiveStats 每月计算机活跃数量统计
func ComputerMonthlyLiveStats() {
	db := global.GDb

	today := time.Now().UTC()
	monthBegin := now.BeginningOfMonth().Add(time.Hour * -8)
	monthEnd := now.EndOfMonth()

	startDate := time.Date(monthBegin.Year(), monthBegin.Month(), monthBegin.Day(), 16, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
	endDate := time.Date(monthEnd.Year(), monthEnd.Month(), monthEnd.Day(), 16, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")

	var monthlyInfo []LiveStats
	sql := fmt.Sprintf("select OSVersion, COUNT(TargetID) as Count from computer_target where LastSyncTime >= '%s' and LastSyncTime < '%s' group by OSVersion;", startDate, endDate)
	if err := db.Raw(sql).Scan(&monthlyInfo).Error; err != nil {
		global.GLog.Error("ComputerMonthlyLiveStats", zap.Any("err", err))
	}
	for i := 0; i < len(monthlyInfo); i++ {
		info := map[string]interface{}{
			"date":           today,
			"computer_count": monthlyInfo[i].Count,
			"stats_time":     today,
			"created_time":   today,
			"last_update":    today,
		}
		var stats model.ComputerLiveStats
		if err := db.Where(model.ComputerLiveStats{Year: today.Year(), Month: int(today.Month()), Cycle: ComputerLiveStatsCycle.MONTH, OSversion: monthlyInfo[i].OSVersion}).Assign(info).FirstOrCreate(&stats).Error; err != nil {
			global.GLog.Error("ComputerMonthlyLiveStats", zap.Any("err", err))
		}
	}
}

// reporting异步任务
func ReportMainLoop() {
	beginTime := time.Now().Add(time.Second * 10)
	// 每2小时执行一次
	component.TimerScheduleServer(component.NewScheduleTask(StatusStatsOfPerComputer, beginTime.Add(time.Second*5), 12, false))
	component.TimerScheduleServer(component.NewScheduleTask(StatusStatsOfPerUpdate, beginTime.Add(time.Second*10), 12, false))
	// 自定义报表数据每2小时异步统计
	component.TimerScheduleServer(component.NewScheduleTask(ComputerReportStats, beginTime.Add(time.Second*15), 12, false))
	component.TimerScheduleServer(component.NewScheduleTask(UpdateReportStats, beginTime.Add(time.Second*20), 12, false))
	component.TimerScheduleServer(component.NewScheduleTask(DssReportStats, beginTime.Add(time.Second*25), 12, false))
	component.TimerScheduleServer(component.NewScheduleTask(UpdateInstallReportStats, beginTime.Add(time.Second*30), 12, false))

	// 每天凌晨1:00执行一次
	t0, _ := time.ParseInLocation(utils.OnlyTimeFormat, "01:00:00", time.Local)
	component.TimerScheduleServer(component.NewScheduleTask(DelReportUselessData, t0, 1, false))
	// 每天23:55:00执行一次日活统计
	t1, _ := time.ParseInLocation(utils.OnlyTimeFormat, "23:55:00", time.Local)
	component.TimerScheduleServer(component.NewScheduleTask(ComputerDailyLiveStats, t1, 1, false))
	// 每天23:56:00执行一次周活统计
	t2, _ := time.ParseInLocation(utils.OnlyTimeFormat, "23:56:00", time.Local)
	component.TimerScheduleServer(component.NewScheduleTask(ComputerWeeklyLiveStats, t2, 1, false))
	// 每天23:57:00执行一次月活统计
	t3, _ := time.ParseInLocation(utils.OnlyTimeFormat, "23:57:00", time.Local)
	component.TimerScheduleServer(component.NewScheduleTask(ComputerMonthlyLiveStats, t3, 1, false))
}
