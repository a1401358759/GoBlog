package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"goblog/core/global"
	"goblog/core/response"
	"goblog/modules/model"
	"goblog/service"
	"goblog/service/cusss/dss"
	"goblog/service/download"
	"goblog/utils"
	"goblog/utils/component"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 获取上游服务器配置
func GetParentUssConfigs() []global.UssSimple {
	globalUssConfigs := service.GetGlobalUssConfig()
	var ussList []global.UssSimple
	for _, uss := range globalUssConfigs {
		ussList = append(ussList, uss)
	}
	return ussList
}

// 修改上游服务器配置
func EditParentUssConfig(ussList []map[string]interface{}) error {
	for _, uss := range ussList {
		if ussId, ok := uss["id"]; ok {
			// check proxy
			client, simpleUss, refreshAnchor := service.RequestCheckClient(ussId.(int), uss)
			protocol := utils.If(*simpleUss.ServerUseTls, "https://", "http://").(string)
			checkUrl := protocol + *simpleUss.ServerIP + ":" + *simpleUss.ServerPort
			if ussId.(int) == dss.UssServerKindCUS {
				checkUrl += "/api/v2/config/verify_cus"
				request, _ := http.NewRequest("GET", checkUrl, nil)
				resp, err := client.Do(request)
				if err != nil || resp.StatusCode != 200 {
					return errors.New("pre check proxy status error")
				} else {
					respXml, _ := ioutil.ReadAll(resp.Body)
					if !strings.Contains(string(respXml), "http://www.microsoft.com") {
						return errors.New("pre check proxy status error")
					}
				}
			} else {
				checkUrl += "/ServerSyncWebService/ServerSyncWebService.asmx"
				request, _ := http.NewRequest("POST", checkUrl, bytes.NewBuffer([]byte(`
						<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
                            <soap:Body>
                                <GetAuthConfig xmlns="http://www.microsoft.com/SoftwareDistribution"></GetAuthConfig>
                            </soap:Body>
                        </soap:Envelope>`)))
				request.Header.Add("Content-Type", "text/xml; charset=utf-8")
				request.Header.Add("SOAPAction", "http://www.microsoft.com/SoftwareDistribution/GetAuthConfig")
				resp, err := client.Do(request)
				if err != nil || resp.StatusCode != 200 {
					return errors.New("pre check proxy status error")
				} else {
					respXml, _ := ioutil.ReadAll(resp.Body)
					if !strings.Contains(string(respXml), "http://www.microsoft.com") {
						return errors.New("pre check proxy status error")
					}
				}
			}
			// refresh anchor
			if refreshAnchor {
				uss["LastConfigAnchor"] = nil
				uss["LastConfigSyncAnchor"] = nil
				uss["LastSyncAnchor"] = nil
				uss["LastDeploymentAnchor"] = nil
			}
			err := global.GDb.Table("parent_uss_config").Where("id=?", ussId).Updates(&uss).Error
			if err != nil {
				return err
			}
		} else {
			return errors.New("param not exists uss pk")
		}
	}
	return nil
}

// 获取同步定时任务信息
func GetSyncSchedule() (map[string]interface{}, error) {
	sc := service.GetGlobalServerConfig()
	syncSchedule := map[string]interface{}{
		"ID":              sc.ID,
		"SyncFirstTime":   sc.SyncFirstTime,
		"SyncMode":        sc.SyncMode,
		"SyncTimesPerDay": sc.SyncTimesPerDay,
	}
	return syncSchedule, nil
}

// 设置同步定时任务信息
func SetSyncSchedule(schedule map[string]interface{}) error {
	err := global.GDb.Table("server_configuration").Updates(schedule).Error
	if err != nil {
		return err
	}
	service.RefreshServerConfig(false)
	DssHandleTimerScheduleRestart()
	return nil
}

// 停止定时同步
func DssHandleTimerScheduleStop() int {
	allLock := global.GRedis.HGetAll(utils.GetFuncName(dss.Handler)).Val()
	for pcName := range allLock {
		global.GRedis.HDel(utils.GetFuncName(dss.Handler), pcName)
	}
	return 1
}

// 开启定时同步
func AutoStartSchedule() {
	sc := service.GetServerConfig()
	if sc.SyncMode == service.SyncTypeAuto {
		allLock := global.GRedis.HGetAll(utils.GetFuncName(dss.Handler)).Val()
		pcStatus := global.GRedis.HGet(utils.GetFuncName(dss.Handler), utils.GetPCName()).Val()
		if len(allLock) == 0 || pcStatus == "alive" {
			beginTime := sc.SyncFirstTime
			t, _ := time.ParseInLocation("15:04:05", beginTime, time.Local)
			component.TimerScheduleServer(component.NewScheduleTask(dss.Handler, t, sc.SyncTimesPerDay, false, false))
		}
	}
}

// 重启定时同步任务
func DssHandleTimerScheduleRestart() {
	DssHandleTimerScheduleStop()
	go AutoStartSchedule()
}

// 设置是否开启Express
func SetExpressEnable(expressEnable map[string]interface{}) error {
	if err := global.GDb.Table("server_configuration").Where("id=?", expressEnable["id"]).Update("EnableExpress", expressEnable["EnableExpress"]).Error; err != nil {
		return err
	}
	service.RefreshServerConfig(false)
	if expressEnable["EnableExpress"] != nil && expressEnable["EnableExpress"].(bool) {
		var preTasks []struct {
			FileDigest, SyncSource, PatchingType string
			RevisionID                           int
		}
		var uriMap = make(map[string]string)
		var expressTask []string
		var commonTask []string
		uss := service.GetGlobalUssConfig()
		for _, u := range uss {
			schema := utils.If(*u.ServerUseTls, "https://", "http://").(string)
			if *u.ID == dss.UssServerKindWSUS {
				uriMap["WSUS"] = schema + *u.ServerIP + ":" + *u.ServerPort + "/Content/"
			} else {
				uriMap["CUS"] = schema + *u.ServerIP + ":" + *u.ServerPort + "/Content/"
			}
		}
		global.GDb.Table("files").Joins("left join revision on files.RevisionID = revision.RevisionID").
			Select("files.FileDigest, revision.SyncSource, files.PatchingType, revision.RevisionID").
			Where("files.IsOnServer = false").Order("files.PatchingType desc").Scan(&preTasks)
		for _, preTask := range preTasks {
			isExpress := utils.If(preTask.PatchingType == "2", 1, 0).(int)
			source := utils.If(preTask.SyncSource == "WSUS", 2, 1).(int)
			task, err := download.NewDownloadTask(preTask.FileDigest, uriMap[preTask.SyncSource], source, isExpress, preTask.RevisionID)
			if err != nil {
				return err
			}
			if isExpress > 0 {
				expressTask = append(expressTask, task)
			} else {
				commonTask = append(commonTask, task)
			}
		}
		commonTask = append(commonTask, expressTask...)
		q := download.NewDownloadQueue()
		runningDigest := q.GetRunning()
		q.Clear()
		if len(runningDigest) > 0 {
			var newSlice = make([]string, len(commonTask)+1)
			var valSlice = make([]string, 1)
			runningTask, err := download.ForceInitTask(runningDigest)
			if err != nil {
				return err
			}
			valSlice = append(valSlice, runningTask)
			copy(newSlice[0:1], valSlice)
			copy(newSlice[1:], commonTask[:])
			commonTask = newSlice
		}
		q.Push(commonTask...)
	}
	return nil
}

func GetOsVersion(isLiveStats string) map[string]string {
	var os = make(map[string]string)
	var query []string
	if isLiveStats == "true" { // gorm
		global.GDb.Model(&model.ComputerLiveStats{}).Distinct("os_version").Pluck("os_version", &query)
	} else {
		global.GDb.Model(&model.ComputerTarget{}).Distinct("OSVersion").Pluck("os_version", &query)
	}
	for i := 0; i < len(query); i++ {
		os[query[i]] = query[i]
	}
	return os
}

// GetHomeCountInfo 首页统计信息
func GetHomeCountInfo(db *gorm.DB) (computerCount, groupsCount, dssCount, updatesCount int64) {
	db.Model(&model.ComputerTarget{}).Count(&computerCount)
	db.Model(&model.ComputerTargetGroup{}).Count(&groupsCount)
	db.Model(&model.Dss{}).Count(&dssCount)
	updatesCount = int64(service.GetRevisionsCount())
	return
}

// GetMSRCSeverityPie 首页 MSRCSeverity 严重程度饼图
func GetMSRCSeverityPie(db *gorm.DB) map[string]int {
	var MSRCSeverity = make(map[string]int)
	var MSRCSeverityList []struct {
		MSRCCount    int
		MSRCSeverity string
	}
	db.Raw("SELECT COUNT(RevisionID) AS MSRCCount, MsrcSeverity FROM revision WHERE UpdateType NOT IN ('Category', 'Detectoid') AND ProductRevisionID IS NOT NULL AND ClassificationRevisionID IS NOT NULL GROUP BY MsrcSeverity;").Scan(&MSRCSeverityList)
	for _, item := range MSRCSeverityList {
		MSRCSeverity[item.MSRCSeverity] = item.MSRCCount
	}
	return MSRCSeverity
}

// GetUpdateClassifyPie 首页更新分类饼图
func GetUpdateClassifyPie(db *gorm.DB) map[int]int {
	var updateClassify = make(map[int]int)
	var updateClassifyList []struct {
		Count                    int
		ClassificationRevisionID int
	}
	db.Raw("SELECT COUNT(RevisionID) AS Count, ClassificationRevisionID FROM revision WHERE UpdateType NOT IN ('Category', 'Detectoid') AND ProductRevisionID IS NOT NULL AND ClassificationRevisionID IS NOT NULL GROUP BY ClassificationRevisionID;").Scan(&updateClassifyList)
	for _, item := range updateClassifyList {
		updateClassify[item.ClassificationRevisionID] = item.Count
	}
	return updateClassify
}

// ComputerUpdateStatsPie 首页计算机更新安装状态饼图
func ComputerUpdateStatsPie(db *gorm.DB, computerCount, updatesCount int64) map[int]int {
	var statsMap = make(map[int]int)
	var statsList []struct {
		Status int
		Count  int
	}
	db.Raw("SELECT Status, COUNT(id) as Count from update_status_per_computer GROUP BY Status;").Scan(&statsList)
	// 初始化statsMap数据
	for key := range service.ComputerUpdateStatusConfigMap {
		statsMap[key] = 0
	}

	knownCount := 0
	for _, item := range statsList {
		knownCount += item.Count
		statsMap[item.Status] = item.Count
	}
	// 状态未知或不适用不存在数据库中
	statsMap[service.ComputerUpdateStatusConfig.Unknown] = int(computerCount*updatesCount - int64(knownCount))

	return statsMap
}

func GetServerInfo() map[string]string {
	sc := service.GetGlobalServerConfig()
	info := map[string]string{
		"ServerID":      sc.ServerID,
		"ServerName":    sc.FullDomainName,
		"Platform":      global.GConfig.CUS.Platform,
		"ServerVersion": global.GConfig.CUS.Version,
		"ServerType":    global.GConfig.CUS.ServerType,
	}
	return info
}

func GetPending() map[string]interface{} {
	sh := service.GetGlobalLastSyncHistory()
	info := map[string]interface{}{
		"PendingId":    sh.ID,
		"Pending":      sh.Pending,
		"SyncCategory": sh.SyncCategory,
		"SyncType":     sh.SyncType,
	}
	return info
}

func GetSyncType() (mapping []map[string]interface{}) {
	for k, v := range service.SyncTypeMap {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return
}

func GetSyncStatus() (mapping []map[string]interface{}) {
	for k, v := range service.SyncStatusMap {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return
}

func GetSyncMode() (mapping []map[string]interface{}) {
	for k, v := range service.SyncModeMap {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return
}

func GetSyncScope() (mapping []map[string]interface{}) {
	for k, v := range service.SyncScope {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return
}

//revision 个数
func GetRevisionCount() map[string]int64 {
	var RevisionCount, RevisionInCategoryCount int64
	global.GDb.Table("revision").Count(&RevisionCount)
	global.GDb.Table("revision_in_category").Count(&RevisionInCategoryCount)
	data := map[string]int64{
		"RevisionCount":           RevisionCount,
		"RevisionInCategoryCount": RevisionInCategoryCount,
	}
	return data
}

//清理规则
func GetCleanRules() (data []map[string]interface{}) {
	C := CleanRules
	CleanRuleList := []int{C.UnusedUpdates, C.NoConnectPC, C.NoNeedUpdates, C.ExpireUpdates, C.SupersededUpdates}
	for _, i := range CleanRuleList {
		data = append(data, map[string]interface{}{
			"id":     i,
			"label":  CleanRulesLabel[i],
			"desc":   CleanRulesDesc[i],
			"module": CleanRulesModule[i],
		})
	}
	return
}

//MSRCSeverity严重程度
func GetMSRCSeverity() (mapping []map[string]interface{}) {
	for k, v := range MSRCSeverityEnMap {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return
}

//审批状态报表过滤
func GetApproveStatus() (mapping []map[string]interface{}) {
	for k, v := range ApproveStatusFilterMap {
		intK, _ := strconv.Atoi(k)
		mapping = append(mapping, map[string]interface{}{
			"key":   intK,
			"label": v,
		})
	}
	return
}

//计算机更新状态
func GetComputerUpdateStatusMap() (mapping []map[string]interface{}) {
	for k, v := range ComputerUpdateStatusMap {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return

}

// Patching Type
func GetPatchingTypeMap() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":   "Unspecified",
			"label": utils.Unspecified,
		},
		{
			"key":   "SelfContained",
			"label": utils.SelfContained,
		},
		{
			"key":   "Express",
			"label": utils.Express,
		},
		{
			"key":   "MSPBinaryDelta",
			"label": utils.MSPBinaryDelta,
		},
		{
			"key":   "Setup360Installer",
			"label": utils.Setup360Installer,
		},
		{
			"key":   "Setup360WIM",
			"label": utils.Setup360WIM,
		},
		{
			"key":   "Setup360Servicing",
			"label": utils.Setup360Servicing,
		},
	}
}

// GetAllErrCode 获取所有错误码mapping
func GetAllErrCode() (mapping []map[string]interface{}) {
	for k, v := range response.ErrorMsg {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return
}

func GetSyncUpdatesType() (mapping []map[string]interface{}) {
	for k, v := range SyncDetailUpdatesMap {
		mapping = append(mapping, map[string]interface{}{
			"key":   k,
			"label": v,
		})
	}
	return
}

//获取前端所有所需的map
func GetAllMap() map[string][]map[string]interface{} {
	mapping := make(map[string][]map[string]interface{})
	mapping["InstallStatus"] = GetComputerUpdateStatusMap()
	mapping["SyncCategory"] = GetSyncScope()
	mapping["SyncStatus"] = GetSyncStatus()
	mapping["SyncMode"] = GetSyncMode()
	mapping["SyncType"] = GetSyncType()
	mapping["MSRC"] = GetMSRCSeverity()
	mapping["CleanRule"] = GetCleanRules()
	mapping["InstallStatus"] = GetComputerUpdateStatusMap()
	mapping["ApproveStatus"] = GetApproveStatus()
	mapping["ErrorCode"] = GetAllErrCode()
	mapping["SyncUpdate"] = GetSyncUpdatesType()
	mapping["PatchingType"] = GetPatchingTypeMap()
	mapping["OsVersion"] = nil
	for i, v := range GetOsVersion("false") {
		mapping["OsVersion"] = append(mapping["OsVersion"], map[string]interface{}{
			"key":   i,
			"label": v,
		})
	}
	mapping["SyncUpdateType"] = []map[string]interface{}{
		{
			"key":   "NewUpdates",
			"label": "新更新",
		},
		{
			"key":   "RevisedUpdates",
			"label": "修订更新",
		},
		{
			"key":   "MSExpiredUpdates",
			"label": "过期更新",
		},
	}
	return mapping
}

// 版本检查 <CUS-C专用>
func VersionCheck(platform string) (interface{}, error) {
	var versionMap = make(map[string]string)
	versionInfoByte, err := ioutil.ReadFile(global.GConfig.CUS.DirPath + "/" + platform + "_latest.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(versionInfoByte, &versionMap); err != nil {
		return nil, err
	}
	return map[string]string{"version": versionMap["version"], "filePath": global.GConfig.CUS.DirPath + versionMap["file_name"]}, nil
}

// 获取版本信息
func Version() (interface{}, error) {
	var checkResult = make(map[string]interface{})
	checkResult["cur"] = global.GConfig.CUS.Version
	if global.GConfig.CUS.ServerType == "CUS-C" {
		checkResult["latest"] = global.GConfig.CUS.Version
		checkResult["url"] = ""
		return checkResult, nil
	} else {
		var checkUrl string
		checkUrl = "http://"
		if global.GConfig.CUS.UseTls {
			checkUrl = "https://"
		}
		remoteUrl := global.GConfig.CUS.DefaultUssIP + ":" + global.GConfig.CUS.DefaultUssPort
		checkUrl += remoteUrl + "/api/v2/config/version/check?platform=" + strings.ToLower(global.GConfig.CUS.Platform)
		resp, err := http.Get(checkUrl)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 200 {
			var respData = make(map[string]interface{})
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(respBody, &respData); err != nil {
				return nil, err
			}
			checkResult["latest"] = respData["data"].(map[string]interface{})["version"]
			checkResult["url"] = remoteUrl + respData["data"].(map[string]interface{})["filePath"].(string)
			return checkResult, nil
		}
		return nil, err
	}
}
