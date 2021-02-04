package admin

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"goblog/core/global"
	"goblog/modules/entity/cusss/deserializer"
	"goblog/modules/entity/cusss/serializer"
	"goblog/modules/entity/cusss/uss/response"
	"goblog/modules/model"
	"goblog/service"
	"goblog/service/cusss/dss"
	"goblog/service/download"
	"goblog/utils"
	"gorm.io/gorm"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BaseRevisionSql = `
	select software.RevisionID                                                                                 as RevisionID,
		   software.UpdateID                                                                                   as UpdateID,
		   software.RevisionNumber                                                                             as RevisionNumber,
		   if(prop.Title is not null, prop.Title,
			  (select Title from property where Language = 'en' and RevisionID = software.RevisionID limit 1)) as Title,
		   software.ProductRevisionID                                                                          as ProductRevisionID,
		   software.ClassificationRevisionID                                                                   as ClassificationRevisionID,
		   software.MsrcSeverity                                                                               as MsrcSeverity,
		   DATE_FORMAT(software.ImportedTime, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00')                                 as ImportedTime,
		   DATE_FORMAT(software.CreationDate, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00')                                 as CreationDate,
		   software.KBArticleID                                                                                as KBArticleID,
		   dep.InstalledCount                                                                                  as InstalledCount%s
	from (select RevisionID, UpdateID, RevisionNumber, ProductRevisionID, ClassificationRevisionID, MsrcSeverity, ImportedTime, CreationDate, KBArticleID
		from revision where ProductRevisionID is not null and ClassificationRevisionID is not null and UpdateType not in ('Category', 'Detectoid')) software
		left join (select * from property where Language = 'zh-cn') prop
	on prop.RevisionID = software.RevisionID
		join (select RevisionID, count(RevisionID) as InstalledCount from deployment where ActionID=0 group by
		RevisionID union
		select DISTINCT RevisionID, 0 as InstalledCount from deployment where ActionID !=0 and RevisionID
		not in (select RevisionID from deployment where ActionID=0) group by RevisionID) dep on software.RevisionID = dep.RevisionID
	`
	RevisionCountSql = `
	select count(1) as Count
	from revision
	where ProductRevisionID is not null
	  and ClassificationRevisionID is not null
	  and UpdateType not in ('Category', 'Detectoid')
	`
	GroupCountSql    = `select count(1) as groupCount from computer_target_group`
	SoftWareTitleSql = `SELECT a.RevisionID as RevisionID,
							   (SELECT Title
								FROM property
								WHERE RevisionID = a.RevisionID AND Language IN ('zh-cn', 'en')
								ORDER BY Language DESC
								LIMIT 1)    AS Title
						FROM (SELECT r.RevisionID, r.ProductRevisionID, r.ClassificationRevisionID
							  FROM revision r
							  WHERE r.UpdateType NOT IN ('Category', 'Detectoid')
								AND r.ProductRevisionID IS NOT NULL
								AND r.ClassificationRevisionID IS NOT NULL % s) a
						order by RevisionID`
	RevisionTitleSql = `SELECT a.RevisionID        as RevisionID,
							   a.CheckedInFrontend as CheckedInFrontend,
							   (SELECT Title
								FROM property
								WHERE RevisionID = a.RevisionID AND Language IN ('zh-cn', 'en')
								ORDER BY Language DESC
								LIMIT 1)           AS Title,
							   (SELECT Description
								FROM property
								WHERE RevisionID = a.RevisionID
								  AND Language IN ('zh-cn', 'en')
								ORDER BY Language DESC
								LIMIT 1)           AS Description,
								a.CategoryType
						FROM (SELECT r.RevisionID, r.ProductRevisionID, r.ClassificationRevisionID, r.CheckedInFrontend, r.CategoryType
							  FROM revision r
							  WHERE %s) a
						order by RevisionID`
	RevisionComputerSql = `
						select *
						from (
								 select c.TargetID,
										c.FullDomainName,
										c.IPAddress,
										c.ComputerMake,
										DATE_FORMAT(c.LastReportedStatusTime, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00') as LastReportedStatusTime,
										if(u.RevisionID = % d, u.Status, '0')                                  as Status
								 from computer_target c
										  left join (select TargetID, Status, RevisionID
													 from update_status_per_computer
													 where RevisionID = % d) u on c.TargetID = u.TargetID
							 ) as result`
	RevisionComputerCountSql = `
						select count(1) as count
						from (
								 select c.TargetID,
										c.FullDomainName,
										c.IPAddress,
										c.ComputerMake,
										DATE_FORMAT(c.LastReportedStatusTime, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00') as LastReportedStatusTime,
										if(u.RevisionID = % d, u.Status, '0')                                  as Status
								 from computer_target c
										  left join (select TargetID, Status, RevisionID
													 from update_status_per_computer
													 where RevisionID = % d) u on c.TargetID = u.TargetID
							 ) as result`
	RevisionGroupSql = `
						select *
						from (
								 select g.TargetGroupID,
										g.TargetGroupName,
										g.Description,
										if(d.ActionID is null, %d, d.ActionID)              as ApproveStatus,
										d.RevisionID,
										d.AdminName,
										if(l.computers_count is null, 0, l.computers_count) as ComputerCount
								 from computer_target_group g
										  left join
									  (select de.*
									   from (select * from deployment where RevisionID = % d) de
												right join
											(select RevisionID, Max(DeploymentID) as DeploymentID, TargetGroupID
											 from deployment
											 group by RevisionID, TargetGroupID) dep
											on de.DeploymentID = dep.DeploymentID) d on d.TargetGroupID = g.TargetGroupID
										  left join (select i.TargetGroupID, count(c.TargetID) as computers_count
													 from computer_target c
															  left outer JOIN computer_in_group i on c.TargetID = i.TargetID
													 group by i.TargetGroupID) l on l.TargetGroupID = g.TargetGroupID)
								 as result
						where (result.RevisionID = % d or result.RevisionID is null)
						  and (result.ApproveStatus in (0, 1, 2, 3, 8) or result.ApproveStatus is null)
						  and result.TargetGroupID not in ('A0A08746-4DBE-4A37-9ADF-9E7652C0B421', 'D374F42A-9BE2-4163-A0FA-3C86A401B7A7')
					   `
	RevisionGroupCountSql = `
						select count(1)
						from (
								 select g.TargetGroupID,
										g.TargetGroupName,
										g.Description,
										if(d.ActionID is null, %d, d.ActionID)              as ApproveStatus,
										d.RevisionID,
										d.AdminName,
										if(l.computers_count is null, 0, l.computers_count) as ComputerCount
								 from computer_target_group g
										  left join
									  (select de.*
									   from (select * from deployment where RevisionID = % d) de
												right join
											(select RevisionID, Max(DeploymentID) as DeploymentID, TargetGroupID
											 from deployment
											 group by RevisionID, TargetGroupID) dep
											on de.DeploymentID = dep.DeploymentID) d on d.TargetGroupID = g.TargetGroupID
										  left join (select i.TargetGroupID, count(c.TargetID) as computers_count
													 from computer_target c
															  left outer JOIN computer_in_group i on c.TargetID = i.TargetID
													 group by i.TargetGroupID) l on l.TargetGroupID = g.TargetGroupID)
								 as result
						where (result.RevisionID = % d or result.RevisionID is null)
						  and (result.ApproveStatus in (0, 1, 2, 3, 8) or result.ApproveStatus is null)
						  and result.TargetGroupID not in ('A0A08746-4DBE-4A37-9ADF-9E7652C0B421', 'D374F42A-9BE2-4163-A0FA-3C86A401B7A7')
					   `
)

func SortCondition(condition string) string {
	sortCondition := strings.TrimPrefix(condition, "-")
	if strings.HasPrefix(condition, "-") {
		sortCondition += " DESC"
	}
	return sortCondition
}

func RevisionList(filter []utils.QueryFilter, sort string, pageSize int, pageNumber int) ([]map[string]interface{}, int, int, error) {
	if pageSize == 0 {
		pageSize = 15
	}
	if pageNumber == 0 {
		pageNumber = 1
	}
	querySql := fmt.Sprintf(BaseRevisionSql, " ")
	countSql := RevisionCountSql
	// where 查询条件
	if len(filter) > 0 {
		for _, f := range filter {
			querySql += " And " + f.Handle()
			countSql += " And " + f.Handle()
		}
	}
	// order by 排序条件
	var sortSql string
	if len(sort) > 0 {
		sortSql = " order by "
		sortSql += SortCondition(sort)
	} else {
		sortSql = " order by RevisionID DESC "
	}
	querySql += sortSql
	// limit 分页条件
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	var resultList []map[string]interface{}
	var count int
	var groupCount int
	if err := global.GDb.Raw(querySql).Scan(&resultList).Error; err != nil {
		return resultList, count, groupCount - 2, err
	}
	if err := global.GDb.Raw(countSql).Scan(&count).Error; err != nil {
		return resultList, count, groupCount - 2, err
	}
	if err := global.GDb.Raw(GroupCountSql).Scan(&groupCount).Error; err != nil {
		return resultList, count, groupCount - 2, err
	}
	return resultList, count, groupCount - 2, nil
}

func GetFileUrl(fileDigest, fileName string) string {
	b64Digest, _ := base64.StdEncoding.DecodeString(fileDigest)
	digest := hex.EncodeToString(b64Digest)
	parentDir := digest[len(digest)-2:]
	suffix := strings.Split(fileName, ".")[len(strings.Split(fileName, "."))-1]
	return "/" + parentDir + "/" + digest + "." + suffix
}

// 获取指定目录下的所有文件
func GetLocalFileList(dirPth string, files *[]string) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		global.GLog.Error("GetLocalFileList", zap.Any("err", err))
	}

	PthSep := string(os.PathSeparator)

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			GetLocalFileList(filepath.Join(dirPth, fi.Name()), files)
		} else {
			*files = append(*files, PthSep+filepath.Join(filepath.Base(filepath.Dir(dirPth+PthSep+fi.Name())), fi.Name()))
		}
	}
}

func GetRevisionDetail(revisionId int) map[string]interface{} {
	// 此接口尝试利用go sync wait异步处理，
	// 性能提升从4.0ms至2.5ms，较为可观
	var wg sync.WaitGroup
	var bundles []model.Bundle
	var updateId string
	global.GDb.Table("revision").Select("UpdateID").Where("RevisionID=?", revisionId).Scan(&updateId)
	global.GDb.Where("RevisionID=?", revisionId).Find(&bundles)
	allRevisionIds := make([]string, 0)
	allRevisionIds = append(allRevisionIds, strconv.Itoa(revisionId))
	for _, bundle := range bundles {
		allRevisionIds = append(allRevisionIds, strconv.Itoa(bundle.BundleRevisionID))
	}
	sc := service.GetGlobalServerConfig()
	filesQuery := global.GDb.Where("RevisionID in (" + strings.Join(allRevisionIds, ",") + ")")
	if !sc.EnableExpress {
		filesQuery = filesQuery.Where("PatchingType != '2'")
	}
	var files []model.DownloadFiles
	filesQuery.Find(&files)
	// 异步遍历本地文件，验证文件完整性
	canBeApprovedChan := make(chan bool)
	go func(channel chan bool, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		var canBeApproved = true
		for _, file := range files {
			filePath := global.GConfig.CUS.DirPath + "/" + GetFileUrl(file.FileDigest, file.FileName)
			_, err := os.Lstat(filePath)
			if !os.IsNotExist(err) {
				fi, _ := os.Stat(filePath)
				if fi.Size() == file.Size {
					file.IsOnServer = true
				} else {
					file.IsOnServer = false
					file.BytesDownloaded = fi.Size()
				}
			} else {
				file.IsOnServer = false
				file.BytesDownloaded = 0
			}
			if !file.IsOnServer && file.PatchingType != "2" {
				canBeApproved = false
			}
			global.GDb.Save(&file)
		}
		channel <- canBeApproved
	}(canBeApprovedChan, &wg)
	// GetCanUnInstall
	canUnInstallChan := make(chan bool)
	go func(channel chan bool, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		// 可公用的查询结构，最好分开做，以免异步操作时候出现异步读写错误
		revisionQuery := global.GDb.Table("revision").Where("RevisionID in (" + strings.Join(allRevisionIds, ",") + ")")
		var canUnInstall = false
		var canUnInstallCount int64
		revisionQuery.Where("CanUninstall = true").Count(&canUnInstallCount)
		if canUnInstallCount > 0 {
			canUnInstall = true
		}
		channel <- canUnInstall
	}(canUnInstallChan, &wg)
	// GetCanRequireUserInput
	canRequireUserInputChan := make(chan bool)
	go func(channel chan bool) {
		defer wg.Done()
		wg.Add(1)
		revisionQuery := global.GDb.Table("revision").Where("RevisionID in (" + strings.Join(allRevisionIds, ",") + ")")
		var canRequireUserInput = false
		var canRequireUserInputCount int64
		revisionQuery.Where("InstallRequiresUserInput = true").Count(&canRequireUserInputCount)
		if canRequireUserInputCount > 0 {
			canRequireUserInput = true
		}
		channel <- canRequireUserInput
	}(canRequireUserInputChan)
	// GetInstallBehavior
	installBehaviorChan := make(chan string)
	go func(channel chan string, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		var maxBehaviors string
		var installBehavior string
		global.GDb.Raw("select max(InstallRebootBehavior) from revision where RevisionID in (" + strings.Join(allRevisionIds, ",") + ")").Scan(&maxBehaviors)
		if len(maxBehaviors) > 0 {
			installBehavior = RebootBehaviorMsg[maxBehaviors]
		} else {
			installBehavior = RebootBehaviorMsg[RebootBehavior.NeverReboots]
		}
		channel <- installBehavior
	}(installBehaviorChan, &wg)
	// GetRevision
	revisionChan := make(chan map[string]interface{})
	go func(channel chan map[string]interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		revision := make(map[string]interface{})
		global.GDb.Table("revision").Select("CreationDate, ImportedTime, "+
			"IsApproved, KBArticleID, LanguageMask, LastChangedAnchor, LocalUpdateID, MsrcSeverity, "+
			"ProductRevisionID, PublicationState, RevisionID, RevisionNumber, SecurityBulletinID, UpdateID, "+
			"UpdateType, ClassificationRevisionID, EulaID, EulaExplicitlyAccepted").Where("RevisionID=?", revisionId).Scan(&revision)
		channel <- revision
	}(revisionChan, &wg)
	// GetProperty
	propertyChan := make(chan map[string]interface{})
	go func(channel chan map[string]interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		property := make(map[string]interface{})
		global.GDb.Table("property").Select("Title, Description, more_info_url").Where("RevisionID=?", revisionId).
			Where("Language in (" + "'en'," + "'zh-cn'" + ")").
			Order("Language desc").Limit(1).Scan(&property)
		channel <- property
	}(propertyChan, &wg)
	// superseded
	supersededChan := make(chan map[string][]map[string]interface{})
	go func(channel chan map[string][]map[string]interface{}, wg *sync.WaitGroup) {
		var supersedingUpdateIds []string
		var supersededRevisionIds []int
		err := global.GDb.Table("superseded").Select("UpdateID").Where("RevisionID=?", revisionId).Scan(&supersedingUpdateIds).Error
		if err != nil {
			fmt.Println(err.Error())
		}
		err = global.GDb.Table("superseded").Select("RevisionID").Where("UpdateID=?", updateId).Scan(&supersededRevisionIds).Error
		if err != nil {
			fmt.Println(err.Error())
		}
		querySql := SoftWareTitleSql
		var supersedingData []map[string]interface{}
		var supersededData []map[string]interface{}
		if len(supersedingUpdateIds) > 0 {
			var sqlSupersedingUpdateIds []string
			for _, supersedingUpdateId := range supersedingUpdateIds {
				sqlSupersedingUpdateIds = append(sqlSupersedingUpdateIds, "'"+supersedingUpdateId+"'")
			}
			supersedingSql := fmt.Sprintf(querySql, " AND r.UpdateID in ("+strings.Join(sqlSupersedingUpdateIds, ",")+")")
			global.GDb.Raw(supersedingSql).Scan(&supersedingData)
		}
		if len(supersededRevisionIds) > 0 {
			var sqlSupersededRevisionIds []string
			for _, supersededRevisionId := range supersededRevisionIds {
				sqlSupersededRevisionIds = append(sqlSupersededRevisionIds, strconv.Itoa(supersededRevisionId))
			}
			supersededSql := fmt.Sprintf(querySql, " AND r.RevisionID IN ("+strings.Join(sqlSupersededRevisionIds, ",")+")")
			global.GDb.Raw(supersededSql).Scan(&supersededData)
		}
		supersededChan <- map[string][]map[string]interface{}{
			"superseded_list":  supersededData,
			"superseding_list": supersedingData,
		}
	}(supersededChan, &wg)
	// wait for all finish
	wg.Wait()
	// get result from channel
	revision := <-revisionChan
	property := <-propertyChan
	canBeApproved := <-canBeApprovedChan
	canUnInstall := <-canUnInstallChan
	canRequireUserInput := <-canRequireUserInputChan
	installBehavior := <-installBehaviorChan
	suData := <-supersededChan
	// 异步刷新File对应关系Cache
	go service.GenFilesRelationship(utils.StrMapInt(allRevisionIds))
	// GetEula
	var eula model.DownloadFiles
	eulaUrl := ""
	eulaStatus := EulaStatus.NotEula
	if len(revision["EulaID"].(string)) > 0 {
		global.GDb.Select("FileDigest, FileName").Where("RevisionID=?", revisionId).Where("isEula=true").Where("isOnServer=true").Scan(&eula)
		if len(eula.FileDigest) > 0 && len(eula.FileName) > 0 {
			eulaUrl = GetFileUrl(eula.FileDigest, eula.FileName)
			var acceptNum int64
			global.GDb.Model(&model.EulaAcceptance{}).Where("eula_id=?", revision["EulaID"].(string)).Count(&acceptNum)
			if revision["EulaExplicitlyAccepted"].(bool) || int(acceptNum) > 0 {
				eulaStatus = EulaStatus.EulaAccept
			} else if len(eulaUrl) > 0 {
				eulaStatus = EulaStatus.EulaFile
			} else {
				eulaStatus = EulaStatus.EulaNoFile
			}
		}
	}
	// Makeup Structs or Map to map
	res := utils.StructMakeUp(revision, property)
	// Makeup other fields
	res["EulaUrl"] = eulaUrl
	res["EulaStatus"] = eulaStatus
	res["CanBeApproved"] = canBeApproved
	res["CanUnInstall"] = canUnInstall
	res["InstallRequiresUserInput"] = canRequireUserInput
	res["InstallRebootBehavior"] = installBehavior
	res["SupersededList"] = suData["superseded_list"]
	res["SupersedingList"] = suData["superseding_list"]
	global.GRedis.HSet("revision_"+strconv.Itoa(revisionId), "title", property["Title"])
	return res
}

func GetUpdateRelatedComputer(pageSize, pageNumber int, filter []utils.QueryFilter, sort string, revisionId int) ([]map[string]interface{}, int) {
	querySql := fmt.Sprintf(RevisionComputerSql, revisionId, revisionId)
	countSql := fmt.Sprintf(RevisionComputerCountSql, revisionId, revisionId)
	if len(filter) > 0 {
		if len(filter) > 0 {
			for idx, f := range filter {
				if idx == 0 {
					querySql += " Where " + f.Handle()
					countSql += " Where " + f.Handle()
				} else {
					querySql += " And " + f.Handle()
					countSql += " And " + f.Handle()
				}
			}
		}
	}
	var sortSql string
	if len(sort) > 0 {
		sortSql = " order by "
		sortSql += SortCondition(sort)
	}
	querySql += sortSql
	pageSize = utils.If(pageSize > 0, pageSize, 15).(int)
	pageNumber = utils.If(pageNumber > 0, pageNumber, 1).(int)
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	var computers []map[string]interface{}
	var count int64
	global.GDb.Raw(querySql).Scan(&computers)
	global.GDb.Raw(countSql).Scan(&count)
	return computers, int(count)
}

// GetUpdateComputersPie 某个更新对应的所有计算机的安装状态的饼图统计
func GetUpdateComputersPie(revisionID int) (results []map[string]int) {
	db := global.GDb
	// 该RevisionID下所有计算机的安装状态
	var statsList []struct {
		Status, Count int
	}
	sql := fmt.Sprintf("select Status, count(TargetID) as Count from update_status_per_computer where RevisionID = %d group by Status;", revisionID)
	db.Raw(sql).Scan(&statsList)
	// 所有计算机的数量
	var computerCount int64
	db.Model(&model.ComputerTarget{}).Select("TargetID").Count(&computerCount)
	// 初始化results数据
	for key := range service.ComputerUpdateStatusConfigMap {
		results = append(results, map[string]int{"key": key, "value": 0})
	}

	knownCount := 0
	for _, item := range statsList {
		knownCount += item.Count
		for _, result := range results {
			if result["key"] == item.Status {
				result["value"] = item.Count
			}
		}
	}
	// 状态未知或不适用不存在数据库中
	for _, result := range results {
		if result["key"] == service.ComputerUpdateStatusConfig.Unknown {
			result["value"] = int(computerCount - int64(knownCount))
		}
		break
	}
	return
}

// GetUpdateWithGroupComputersPie 某个更新对应的某个组下的计算机的安装状态的饼图统计
func GetUpdateWithGroupComputersPie(revisionID int, groupID string) (results []map[string]int) {
	db := global.GDb
	// 获取组下所有计算机
	var targetIDs []int
	db.Model(&model.ComputerInGroup{}).Where("TargetGroupID = ?", groupID).Pluck("TargetID", &targetIDs)

	// 该RevisionID下所有计算机的安装状态
	var statsList []struct {
		Status, Count int
	}
	sql := fmt.Sprintf("select Status, count(TargetID) as Count from update_status_per_computer where RevisionID = %d and TargetID in (%s) group by Status;", revisionID, GenSqlStrInt(targetIDs))
	db.Raw(sql).Scan(&statsList)
	// 初始化results数据
	for key := range service.ComputerUpdateStatusConfigMap {
		results = append(results, map[string]int{"key": key, "value": 0})
	}

	knownCount := 0
	for _, item := range statsList {
		knownCount += item.Count
		for _, result := range results {
			if result["key"] == item.Status {
				result["value"] = item.Count
			}
			break
		}
	}
	// 状态未知或不适用不存在数据库中
	for _, result := range results {
		if result["key"] == service.ComputerUpdateStatusConfig.Unknown {
			result["value"] = len(targetIDs) - knownCount
		}
	}
	return
}

func GetUpdateRelatedGroup(pageSize, pageNumber int, filter []utils.QueryFilter, sort string, revisionId int) ([]map[string]interface{}, int) {
	var defaultActionId int
	global.GDb.Table("deployment").Select("ActionID").Where("RevisionID=?", revisionId).Where("TargetGroupID=?", utils.UUIDAllComputer).Scan(&defaultActionId)
	defaultActionId = utils.If(defaultActionId > 0, defaultActionId, 2).(int)
	querySql := fmt.Sprintf(RevisionGroupSql, defaultActionId, revisionId, revisionId)
	countSql := fmt.Sprintf(RevisionGroupCountSql, defaultActionId, revisionId, revisionId)
	if len(filter) > 0 {
		if len(filter) > 0 {
			for _, f := range filter {
				querySql += " And " + f.Handle()
				countSql += " And " + f.Handle()
			}
		}
	}
	var sortSql string
	if len(sort) > 0 {
		sortSql = " order by "
		sortSql += SortCondition(sort)
	} else {
		sortSql = "order by FIELD(TargetGroupName,'Unassigned Computers') desc "
	}
	querySql += sortSql
	pageSize = utils.If(pageSize > 0, pageSize, 15).(int)
	pageNumber = utils.If(pageNumber > 0, pageNumber, 1).(int)
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	var groups []map[string]interface{}
	var count int64
	global.GDb.Raw(querySql).Scan(&groups)
	global.GDb.Raw(countSql).Scan(&count)
	return groups, int(count)
}

func GetUpdateRelatedFiles(pageSize, pageNumber, revisionId int) ([]map[string]interface{}, map[string]interface{}) {
	// 此处逻辑做了大量精简，主要省略了每次都读文件的操作。因为该方法在详情页之后调用，而详情页会做一次文件轮询检查的操作，
	// 两处有逻辑先后，即Revision的文件列表页必须由Revision详情页进入，所以此处不必再次进行文件检查。
	var files []map[string]interface{}
	query := global.GDb.Table("files").Select("BytesDownloaded, FileDigest, FileName, PatchingType, IsOnServer, Size, TotalBytesForDownload")
	revisionIds, err := service.GenBundleRevisionIds([]string{strconv.Itoa(revisionId)})
	if err != nil {
		global.GLog.Error("获取bundle revision id列表出错: ", zap.Any("err", err))
	}
	revisionIds = append(revisionIds, strconv.Itoa(revisionId))
	query = query.Where("RevisionID in (" + strings.Join(revisionIds, ",") + ")")
	sc := service.GetGlobalServerConfig()
	if !sc.EnableExpress {
		query = query.Where("PatchingType != '2'")
	}
	query = query.Limit(pageSize).Offset((pageNumber - 1) * pageSize)
	query.Scan(&files)
	q := download.NewDownloadQueue()
	runningTaskDigest := q.GetRunning()
	var meta = make(map[string]interface{})
	meta["IsDownloading"] = false
	TotalBytesDownloaded := 0
	TotalSize := 0
	HaveDownload := 0
	for i := 0; i < len(files); i++ {
		if files[i]["FileDigest"] == runningTaskDigest {
			files[i]["IsDownloading"] = true
			meta["IsDownloading"] = true
		}
		if utils.ToInt(files[i]["IsOnServer"]) > 0 {
			HaveDownload += 1
		}
		files[i]["FileUrl"] = GetFileUrl(files[i]["FileDigest"].(string), files[i]["FileName"].(string))
		TotalBytesDownloaded += utils.ToInt(files[i]["BytesDownloaded"])
		TotalSize += utils.ToInt(files[i]["Size"])
	}
	if TotalSize > 0 {
		meta["DownloadedPercentage"] = float64(TotalBytesDownloaded) / float64(TotalSize)
	} else {
		meta["DownloadedPercentage"] = 0
	}
	go service.GenFilesRelationship(utils.StrMapInt(revisionIds))
	meta["DownloadedSize"] = TotalBytesDownloaded
	meta["TotalSize"] = TotalSize
	meta["HaveDownloadCount"] = HaveDownload
	meta["Count"] = len(files)
	return files, meta
}

func RevisionStartDownloadFiles(revisionId string) {
	var revisionIds []string
	revisionIds = append(revisionIds, revisionId)
	go download.ProcessDownload(revisionIds, true, true)
}

func RevisionStopDownloadFiles(revisionId string) {
	var revisionIds []string
	revisionIds = append(revisionIds, revisionId)
	go download.ProcessDownload(revisionIds, false, true)
}

func approveCheck(revision model.Revision) error {
	if len(revision.EulaID) > 0 {
		var eulaAccepted int64
		global.GDb.Table("eula_acceptance").Where("eula_id=?", revision.EulaID).Count(&eulaAccepted)
		if !(revision.EulaExplicitlyAccepted || eulaAccepted > 0) {
			return errors.New("尚未接受许可协议, 无法安装或卸载该更新。")
		}
	}
	if revision.PublicationState == utils.PublicationState.Expired {
		return errors.New("此更新已过期，无法审批安装该更新，建议拒绝该更新。")
	}
	return nil
}

func RevisionApprove(revisionId string, actionId int, groupIds []string, groupNames []string, userEmail string) (error, string, int) {
	var revision model.Revision
	intRevisionId, _ := strconv.Atoi(revisionId)
	global.GDb.First(&revision)
	err := approveCheck(revision)
	if err != nil {
		return err, "", 0
	}
	var groupMap = make(map[string]string)
	for idx, groupId := range groupIds {
		groupMap[groupId] = groupNames[idx]
	}
	// 已经满足当前审批条件的deployment
	var yesGroupIds []string
	global.GDb.Table("deployment").Select("TargetGroupID").
		Where("RevisionID="+revisionId).
		Where("TargetGroupID in ("+strings.Join(groupIds, ",")+")").
		Where("ActionID=?", actionId).Scan(&yesGroupIds)
	var realGroupIds []string
	if len(yesGroupIds) > 0 {
		for _, groupId := range groupIds {
			if !utils.ContainStr(groupId, yesGroupIds) {
				realGroupIds = append(realGroupIds, groupId)
			}
		}
	}
	var revisionIds []string
	bundleRevisionIds, err := service.GenBundleRevisionIds([]string{revisionId})
	if err != nil {
		return err, "", 0
	}
	revisionIds = append(revisionIds, bundleRevisionIds...)
	revisionIds = append(revisionIds, revisionId)
	sc := service.GetGlobalServerConfig()
	adminName := sc.FullDomainName + "/" + userEmail
	// 查询需要删除的旧的Deployment
	var delDeploymentsMap = make(map[string]int)
	var delDeployments []model.Deployment
	if actionId == utils.DeclineAction {
		global.GDb.Where("RevisionID in (", strings.Join(revisionIds, ",")+")").
			Where("TargetGroupID != ?", utils.UUIDGroupDss).
			Where("ActionID != ?", utils.DeclineAction).Find(&delDeployments)
		var needDeclineRevisions []int
		var deadDeployment []map[string]interface{}
		for _, delDeployment := range delDeployments {
			needDeclineRevisions = append(needDeclineRevisions, delDeployment.RevisionID)
			deadDeployment = append(deadDeployment, map[string]interface{}{
				"version_id":      1,
				"RevisionID":      delDeployment.RevisionID,
				"TargetGroupID":   delDeployment.TargetGroupID,
				"TargetGroupName": delDeployment.TargetGroupName,
				"ActionID":        delDeployment.ActionID,
				"DeploymentGuid":  uuid.New().String(),
				"AdminName":       adminName,
				"LastChangeTime":  time.Now().UTC(),
			})
		}
		global.GDb.Where("RevisionID in (", strings.Join(revisionIds, ",")+")").
			Where("TargetGroupID != ?", utils.UUIDGroupDss).
			Where("ActionID != ?", utils.DeclineAction).Delete(&model.Deployment{})
		var newDeployments []map[string]interface{}
		for _, revisionId := range needDeclineRevisions {
			newDeployments = append(newDeployments, map[string]interface{}{
				"version_id":      1,
				"RevisionID":      revisionId,
				"TargetGroupID":   utils.UUIDAllComputer,
				"TargetGroupName": utils.NameAllComputer,
				"ActionID":        utils.DeclineAction,
				"DeploymentGuid":  uuid.New().String(),
				"AdminName":       adminName,
				"LastChangeTime":  time.Now().UTC(),
			})
		}
		global.GDb.Model(&model.DeadDeployment{}).CreateInBatches(&deadDeployment, 1000)
		global.GDb.Model(&model.Deployment{}).CreateInBatches(&newDeployments, 1000)
		if len(needDeclineRevisions) > 0 {
			download.ProcessDownload(utils.SliceIntToString(needDeclineRevisions), false, false)
		}
		return nil, revision.UpdateID, revision.RevisionNumber
	} else {
		global.GDb.Where("RevisionID in (", strings.Join(revisionIds, ",")+")").
			Where("TargetGroupID in (" + strings.Join(realGroupIds, ",") + ")").
			Where("ActionID in (0, 3, 1, 2, 5, 8)").Find(&delDeployments)
		for _, delDeployment := range delDeployments {
			delDeploymentsMap[strconv.Itoa(delDeployment.RevisionID)+"_"+delDeployment.TargetGroupID] = delDeployment.ActionID
		}
		global.GDb.Where("RevisionID in (", strings.Join(revisionIds, ",")+")").
			Where("TargetGroupID in (" + strings.Join(realGroupIds, ",") + ")").
			Where("ActionID in (0, 3, 1, 2, 5, 8)").Delete(&model.Deployment{})
		// 如果当前RevisionID和所有计算机组有ActionID=8的deployment关系，说明之前是decline状态，需要修改和所有计算机组的deployment关系(ActionID=2)
		var declineDeployExists int64
		global.GDb.Table("deployment").Where("RevisionID="+revisionId).Where("TargetGroupID=?", utils.UUIDAllComputer).Where("ActionID=?", utils.DeclineAction).Count(&declineDeployExists)
		if declineDeployExists > 0 {
			global.GDb.Table("deployment").Where("RevisionID="+revisionId).Where("ActionID=?", utils.DeclineAction).
				Updates(map[string]interface{}{
					"ActionID":       utils.PreDeploymentCheckAction,
					"LastChangeTime": time.Now().UTC(),
					"AdminName":      adminName})
			global.GDb.Table("deployment").Where("RevisionID in (", strings.Join(bundleRevisionIds, ",")+")").Where("ActionID=?", utils.DeclineAction).
				Updates(map[string]interface{}{
					"ActionID":       utils.BundleAction,
					"LastChangeTime": time.Now().UTC(),
					"AdminName":      adminName})
		}
		var newDeployments []map[string]interface{}
		var deadDeployments []map[string]interface{}
		for _, groupId := range realGroupIds {
			newDeployments = append(newDeployments, map[string]interface{}{
				"version_id":      1,
				"RevisionID":      intRevisionId,
				"TargetGroupID":   groupId,
				"TargetGroupName": groupMap[groupId],
				"ActionID":        actionId,
				"DeploymentGuid":  uuid.New().String(),
				"AdminName":       adminName,
				"LastChangeTime":  time.Now().UTC(),
			})
			for _, bundleRevisionId := range bundleRevisionIds {
				intBundleRevisionId, _ := strconv.Atoi(bundleRevisionId)
				newDeployments = append(newDeployments, map[string]interface{}{
					"version_id":      1,
					"RevisionID":      intBundleRevisionId,
					"TargetGroupID":   groupId,
					"TargetGroupName": groupMap[groupId],
					"ActionID":        utils.BundleAction,
					"DeploymentGuid":  uuid.New().String(),
					"AdminName":       adminName,
					"LastChangeTime":  time.Now().UTC(),
				})
			}
			for _, revisionId := range revisionIds {
				key := revisionId + "_" + groupId
				deadActionId, ok := delDeploymentsMap[key]
				if ok {
					intDeadRevisionId, _ := strconv.Atoi(revisionId)
					deadDeployments = append(deadDeployments, map[string]interface{}{
						"version_id":      1,
						"RevisionID":      intDeadRevisionId,
						"TargetGroupID":   groupId,
						"TargetGroupName": groupMap[groupId],
						"ActionID":        deadActionId,
						"DeploymentGuid":  uuid.New().String(),
						"AdminName":       adminName,
						"LastChangeTime":  time.Now().UTC(),
					})
				}
			}
		}
		global.GDb.Model(&model.Deployment{}).CreateInBatches(&newDeployments, 1000)
		global.GDb.Model(&model.DeadDeployment{}).CreateInBatches(&deadDeployments, 1000)
		if actionId == utils.InstallAction {
			download.ProcessDownload([]string{revisionId}, true, true)
		} else {
			download.ProcessDownload([]string{revisionId}, false, true)
		}
		service.GenDeploymentRelationship(utils.StrMapInt(revisionIds))
		return nil, revision.UpdateID, revision.RevisionNumber
	}
}

func GetClassifications() []map[string]interface{} {
	var classificationData []map[string]interface{}
	classificationSql := `select r.RevisionID from
        (select r1.LocalUpdateID, Max(r1.RevisionNumber) RevisionNumber from revision r1
        group by LocalUpdateID) tmp
    left join revision r on r.LocalUpdateID = tmp.LocalUpdateID and r.RevisionNumber = tmp.RevisionNumber
    where r.CategoryType = 'UpdateClassification';`
	var classificationIds []int
	var sqlClassificationIds []string
	global.GDb.Raw(classificationSql).Scan(&classificationIds)
	for _, classificationId := range classificationIds {
		sqlClassificationIds = append(sqlClassificationIds, strconv.Itoa(classificationId))
	}
	querySql := fmt.Sprintf(RevisionTitleSql, " r.RevisionID IN ("+strings.Join(sqlClassificationIds, ",")+")")
	global.GDb.Raw(querySql).Scan(&classificationData)
	return classificationData
}

func GetProductions() ([]map[string]interface{}, []interface{}, map[int][]int) {
	var productData []map[string]interface{}
	productSql := `
	select parent.RevisionID as Id, parent.CategoryType, child.RevisionID as ChildId from (
    select RevisionID, LocalUpdateID, CategoryType from revision r
    join (select Max(RevisionNumber) as rn, UpdateID from revision where CategoryType in ('Company', 'ProductFamily')
    GROUP BY UpdateID) latest
    on r.UpdateID = latest.UpdateID and r.RevisionNumber = latest.rn)
    as parent
    left join update_for_prerequisite u on parent.LocalUpdateID = u.LocalUpdateID
    left join revision_prerequisite re on u.PrerequisiteID = re.PrerequisiteID
    left join (select RevisionID, LocalUpdateID, CategoryType from revision r
    join (select Max(RevisionNumber) as rn, UpdateID from revision where CategoryType in ('Company', 'ProductFamily', 'Product')
    GROUP BY UpdateID) latest on r.UpdateID = latest.UpdateID and r.RevisionNumber = latest.rn) child on
    re.RevisionID = child.RevisionID;
	`
	var ProductSelfRelatedList []map[string]interface{}
	global.GDb.Raw(productSql).Scan(&ProductSelfRelatedList)
	var productTreeMap = make(map[int][]int)
	var firstLevelProducts []interface{}
	var productIds []interface{}
	for _, productSelfRelated := range ProductSelfRelatedList {
		productIds = append(productIds, productSelfRelated["Id"])
		if utils.ToInt(productSelfRelated["ChildId"]) > 0 {
			productIds = append(productIds, utils.ToInt(productSelfRelated["ChildId"]))
			productTreeMap[utils.ToInt(productSelfRelated["Id"])] = append(productTreeMap[utils.ToInt(productSelfRelated["Id"])], utils.ToInt(productSelfRelated["ChildId"]))
		}
		if productSelfRelated["CategoryType"] == "Company" {
			firstLevelProducts = append(firstLevelProducts, utils.ToInt(productSelfRelated["Id"]))
		}
	}
	firstLevelProducts, _ = utils.SliceDistinct(firstLevelProducts)
	productIds, _ = utils.SliceDistinct(productIds)
	var sqlProductIds []string
	for _, productId := range productIds {
		sqlProductIds = append(sqlProductIds, strconv.Itoa(utils.ToInt(productId)))
	}
	querySql := fmt.Sprintf(RevisionTitleSql, " r.RevisionID IN ("+strings.Join(sqlProductIds, ",")+")")
	global.GDb.Raw(querySql).Scan(&productData)
	return productData, firstLevelProducts, productTreeMap
}

func productRecursion(productDataMap map[int]map[string]interface{}, productTreeMap map[int][]int, pk int) (res interface{}) {
	data := productDataMap[pk]
	childIds := productTreeMap[pk]
	if len(childIds) > 0 {
		var children []interface{}
		for _, childId := range childIds {
			children = append(children, productRecursion(productDataMap, productTreeMap, childId))
		}
		return map[string]interface{}{
			"RevisionID":        data["RevisionID"],
			"Title":             data["Title"],
			"Description":       data["Description"],
			"CheckedInFrontend": data["CheckedInFrontend"],
			"Children":          children,
		}
	}
	return data
}

func GetProductTree() []interface{} {
	productData, firstLevelProducts, productTreeMap := GetProductions()
	var productDataMap = make(map[int]map[string]interface{})
	for _, product := range productData {
		productDataMap[utils.ToInt(product["RevisionID"])] = product
	}
	var res []interface{}
	for _, firstLevelProduct := range firstLevelProducts {
		res = append(res, productRecursion(productDataMap, productTreeMap, utils.ToInt(firstLevelProduct)))
	}
	return res
}

// 获取下游服务器列表
func GetDss(pageSize, pageNumber int, filter []utils.QueryFilter) ([]map[string]interface{}, int) {
	var dss []map[string]interface{}
	var count int64
	pageSize = utils.If(pageSize == 0, 15, pageSize).(int)
	pageNumber = utils.If(pageNumber == 0, 1, pageNumber).(int)
	query := global.GDb.Table("dss").
		Select("id, ServerId, FullDomainName, LastRollupTime, LastSyncTime, Version")
	countQuery := global.GDb.Table("dss")
	//如果进行筛选
	if len(filter) > 0 {
		for _, f := range filter {
			query = query.Where(f.Handle())
			countQuery = countQuery.Where(f.Handle())
		}
	}
	countQuery.Count(&count)
	query = query.Limit(pageSize).Offset((pageNumber - 1) * pageSize)
	query.Scan(&dss)
	return dss, int(count)
}

// 获取下游服务器详情
func DssDetail(dssID string) map[string]interface{} {
	var dss = make(map[string]interface{})
	global.GDb.Model(&model.Dss{}).
		Select("id, ServerId, FullDomainName, LastRollupTime, LastSyncTime, Version, UpdateCount,"+
			" NotApprovedUpdateCount, UpdatesNeedingFilesCount, UpdatesUpToDateCount, CustomComputerTargetGroupCount,"+
			" ComputerTargetCount, ComputerTargetsNeedingUpdatesCount").
		Where("id = ?", dssID).First(&dss)
	return dss
}

// 获取同步记录列表
func GetSyncHistory(pageSize, pageNumber int, filter []utils.QueryFilter) ([]map[string]interface{}, int) {
	var syncHistory []map[string]interface{}
	var count int64
	pageSize = utils.If(pageSize == 0, 15, pageSize).(int)
	pageNumber = utils.If(pageNumber == 0, 1, pageNumber).(int)
	query := global.GDb.Table("sync_history").
		Select("id, ParentServerIP, StartTime, FinishTime, NewUpdates, RevisedUpdates, ExpiredUpdates, MSExpiredUpdates, ReplicationMode")
	countQuery := global.GDb.Table("sync_history")
	//如果进行筛选
	if len(filter) > 0 {
		for _, f := range filter {
			query = query.Where(f.Handle())
			countQuery = countQuery.Where(f.Handle())
		}
	}
	countQuery.Count(&count)
	query = query.Limit(pageSize).Offset((pageNumber - 1) * pageSize)
	query.Scan(&syncHistory)
	return syncHistory, int(count)
}

// 获取文件下载和未下载列表
func GetFileList(isDownloading string, pageSize, pageNumber int) ([]map[string]interface{}, int) {
	var Files []map[string]interface{}
	var PageFiles []map[string]interface{}
	var DigestIdMap = make(map[interface{}][]interface{})
	var IdIndexMap = make(map[interface{}][]int)
	var Ids []interface{}
	var Index []int
	var IfDuplicate bool
	var InIndex int
	var data []map[string]interface{}
	var total int
	global.GDb.Table("files").Select("id, FileDigest").Find(&Files)
	// 构建digest：id 的map，一个digest可以对应多个id
	for _, file := range Files {
		DigestIdMap[file["FileDigest"]] = append(DigestIdMap[file["FileDigest"]], file["id"])
	}
	// 获取redis下载队列中file的全部digest
	q := download.NewDownloadQueue()
	runningTaskDigest := q.GetRunning()
	downloadTaskDigest := append(q.GetAll(), runningTaskDigest)
	// 获取digest对应的所有的id
	for d := range downloadTaskDigest {
		if id, ok := DigestIdMap[d]; ok {
			Ids = append(Ids, id)
		}
	}
	if isDownloading == "file_download" && Ids != nil {
		page := Ids[(pageNumber-1)*pageSize : pageNumber*pageSize]
		// 获取id：index的map，一个id可以对应多个index，用于补上去重后的slice
		for in, id := range Ids {
			IdIndexMap[id] = append(IdIndexMap[id], in)
		}
		global.GDb.Table("files").
			Select("id, BytesDownloaded, FileDigest, FileName, PatchingType, IsOnServer, Size").
			Where("id (in) ?", page).Find(&PageFiles)
		AllFiles := make([]map[string]interface{}, len(page))
		// 如果有重复id的文件，因为要全部展示，补上去重后的slice
		if len(page) > len(PageFiles) {
			IfDuplicate = true
			for in, file := range PageFiles {
				flag := 1
				for index := range IdIndexMap[file["id"]] {
					AllFiles[index] = PageFiles[in]
					if flag == 1 {
						Index = append(Index, index)
					}
					flag += 1
				}
			}
		}
	} else if isDownloading == "file_un_download" {
		q := global.GDb.Table("files").
			Select("id, BytesDownloaded, FileDigest, FileName, PatchingType, IsOnServer, Size, RevisionID").
			Where("IsEula = ? AND IsOnServer = ?", 0, 0).
			Not("PatchingType = ?", "2").
			Limit(pageSize).Offset((pageNumber - 1) * pageSize)
		if Ids == nil {
			q.Find(&PageFiles)
		} else {
			q.Not("id in (?)", Ids).Find(PageFiles)
		}
	} else {
		return nil, 0
	}
	total = len(PageFiles)
	BundledMap, _ := BundledMap()
	var RevisionMap = make(map[int]interface{})
	var RevisionIds []int
	for _, pageFile := range PageFiles {
		RevisionId := utils.ToInt(pageFile["RevisionID"])
		rId, ok := BundledMap[RevisionId]
		if ok {
			RevisionMap[RevisionId], _ = strconv.Atoi(rId)
		} else {
			RevisionMap[RevisionId] = pageFile["RevisionID"]
		}
		RevisionIds = append(RevisionIds, utils.ToInt(RevisionMap[RevisionId]))
	}

	for i := 0; i < len(PageFiles); i = i + 1 {
		//如果有两个文件都处于下载中的时候，只显示第一个文件在下载中
		for _, index := range Index {
			InIndex = utils.If(i == index, 1, 0).(int)
		}

		data = append(data, map[string]interface{}{
			"Id":              PageFiles[i]["id"],
			"FileName":        PageFiles[i]["FileName"],
			"FileDigest":      PageFiles[i]["FileDigest"],
			"BytesDownloaded": PageFiles[i]["BytesDownloaded"],
			"PatchingType":    PageFiles[i]["PatchingType"],
			"Size":            PageFiles[i]["Size"],
			"FileUrl":         GetFileUrl(PageFiles[i]["FileDigest"].(string), PageFiles[i]["FileName"].(string)),
			"RevisionID":      RevisionMap[utils.ToInt(PageFiles[i]["RevisionID"])],
			"RevisionName":    GetRevisionTitles(RevisionIds)[utils.ToInt(RevisionMap[utils.ToInt(PageFiles[i]["RevisionID"])])],
			"IsDownloading":   pageNumber == 1 && PageFiles[i]["FileDigest"] == runningTaskDigest && ((InIndex == 1 && IfDuplicate) || !IfDuplicate),
		})
	}

	return data, total
}

// 操作文件下载队列
func EditFiles(action bool, TargetDigests []string) error {
	q := download.NewDownloadQueue()
	relTasks := make([]string, 0)
	var TaskDigestMap = make(map[string]string)
	for _, digest := range TargetDigests {
		if task, err := download.ForceInitTask(digest); err == nil {
			relTasks = append(relTasks, task)
			TaskDigestMap[task] = digest
		} else {
			return err
		}
	}
	if action {
		q.Push(relTasks...)
	} else {
		for _, task := range relTasks {
			DownloadingDigest := q.GetRunning()
			// 当前文件正在下载则需要暂停下载，非正在下载文件才需要移除
			q.Move(task)
			if TaskDigestMap[task] == DownloadingDigest {
				q.Pause()
			}
		}

	}
	return nil
}

func SelectProduct(revisionIds []int) error {
	if err := global.GDb.Table("revision").Where("UpdateType in ("+strings.Join([]string{
		"'Category'", "'Company'", "'ProductFamily'", "'Product'",
	}, ",")).Update("CheckedInFrontend", false).Error; err != nil {
		return err
	}
	if err := global.GDb.Table("revision").Where("RevisionID in (?)", revisionIds).Update("CheckedInFrontend", true).Error; err != nil {
		return err
	}
	return nil
}

func SelectClassification(revisionIds []int) error {
	if err := global.GDb.Table("revision").Where("UpdateType = 'Category'").Where("CategoryType = 'UpdateClassification'").Update("CheckedInFrontend", false).Error; err != nil {
		return err
	}
	if err := global.GDb.Table("revision").Where("RevisionID in (?)", revisionIds).Update("CheckedInFrontend", true).Error; err != nil {
		return err
	}
	return nil
}

// 文件上传,传入文件路径，返回各个数据增加数量
func UploadFile(filePath string) (newUpdateCount, newRevisionCount, newLanguageCount, newFileCount int64, err error) {
	var fileXml []byte
	var fileTxt [][]byte
	var oldUpdateCount, oldRevisionCount, oldLanguageCount, oldFileCount int64
	db := global.GDb
	db.Table("update").Count(&oldUpdateCount)
	db.Table("revision").Count(&oldRevisionCount)
	db.Table("update_language").Count(&oldLanguageCount)
	db.Table("files").Count(&oldFileCount)
	fileXml, fileTxt = utils.DecompressFiles(filePath)
	var exportPackage serializer.ExportPackage
	if err := xml.Unmarshal(fileXml, &exportPackage); err != nil {
		global.GLog.Error("exportPackage Unmarshal", zap.Any("err", err))
		return 0, 0, 0, 0, err
	}
	//找到语言节点 导入更新语言
	var Languages = exportPackage.Languages.Language
	var newLanguages []model.UpdateLanguage
	var allLanguages []int
	db.Table("update_language").Select("LanguageID").Pluck("LanguageID", &allLanguages)
	for _, value := range Languages {
		if !utils.ContainInt(value.LanguageId, allLanguages) {
			newLanguages = append(newLanguages, model.UpdateLanguage{
				LanguageID: value.LanguageId,
				Enabled:    value.Enabled,
				LongName:   value.LongName,
				ShortName:  value.ShortName,
			})
		}
	}
	if err := db.CreateInBatches(&newLanguages, 3000).Error; err != nil {
		global.GLog.Error("Add Language error", zap.Any("err", err))
		return 0, 0, 0, 0, err
	}
	// 找到更新节点 导入入更新
	// 每个更新一个事务
	var Updates = exportPackage.Updates.Update
	var i int = 0
	for _, value := range Updates {
		if err := db.Transaction(func(tx *gorm.DB) error {
			var updateID = value.UpdateId
			var XmlStr []byte
			// TODO: 目前有可能出现第一个字符为空的情况
			if string(fileTxt[i][36]) == "," && string(fileTxt[i][0:36]) == strings.ToLower(updateID) {
				XmlStr = fileTxt[i][55:]
			} else if string(fileTxt[i][1:37]) == strings.ToLower(updateID) {
				XmlStr = fileTxt[i][56:]
			}
			//	调用存储过程
			if len(string(XmlStr)) > 0 {
				// 压缩xml
				var XmlCompressed string
				XmlCompressedStr, _ := utils.XmlCompress(string(XmlStr))
				XmlCompressed = XmlCompressedStr
				var Meta deserializer.Meta
				if err := xml.Unmarshal(XmlStr, &Meta); err != nil {
					global.GLog.Error("Unmarshal Meta error", zap.Any("err", err))
					return err
				}
				// 解析xml
				CoreXml, OtherPropertiesXml, ExtendedPropertiesXml, LocalizeList, EulaList := dss.ParseXml(Meta)
				// 调用存储过程import_data
				err, _ := utils.CallProc("import_data", XmlCompressed, "FileImport", "False", CoreXml, EulaList, ExtendedPropertiesXml, LocalizeList, OtherPropertiesXml, 0, "@result", "@eula_digests", "@eula_uri")
				if err != nil {
					global.GLog.Error("Run Procedure error", zap.Any("err", err))
					return err
				}
			}
			i += 1
			return nil
		}); err != nil {
			newUpdateCount, newRevisionCount, newLanguageCount, newFileCount = GetImportNumber(oldUpdateCount, oldRevisionCount, oldLanguageCount, oldFileCount)
			return newUpdateCount, newRevisionCount, newLanguageCount, newFileCount, err
		}

	}
	// 然后通过文件的上半部分，获取文件的文件名
	var AllFile = exportPackage.Files.File
	var newFileDigest []string
	var MuUrlExist = false
	var MuUrlSql = "UPDATE files SET MUURL=CASE FileDigest"
	for _, value := range AllFile {
		var FileDigest = value.FileDigest
		var FileMuUrl = value.MUURL
		newFileDigest = append(newFileDigest, FileDigest)
		if FileMuUrl == "" || len(FileMuUrl) > 0 {
			MuUrlExist = true
			MuUrlSql += fmt.Sprintf(" WHEN '%s' THEN '%s'", FileDigest, FileMuUrl)
		}
	}
	if len(newFileDigest) == 1 {
		MuUrlSql += fmt.Sprintf(" END WHERE FileDigest = '%s'", newFileDigest)
	} else {
		MuUrlSql += fmt.Sprintf(" END WHERE FileDigest IN (%s)", GenSqlStrString(newFileDigest))
	}
	if MuUrlExist {
		db.Exec(MuUrlSql)
	}
	newUpdateCount, newRevisionCount, newLanguageCount, newFileCount = GetImportNumber(oldUpdateCount, oldRevisionCount, oldLanguageCount, oldFileCount)

	// 处理叶子节点
	err = ImportIsLeaf()
	if err != nil {
		return
	}
	return
}

func ImportIsLeaf() (err error) {
	db := global.GDb
	var isLeaf []string
	var isNotLeaf []string
	if err := db.Transaction(func(tx *gorm.DB) error {
		var baseSql, isLeafSql, isNotLeafSql string
		baseSql = "select revisionID from revision left outer join update_for_prerequisite ufp on revision.LocalUpdateID = ufp.LocalUpdateID "
		isLeafSql = baseSql + "where ufp.LocalUpdateID is null "
		isNotLeafSql = baseSql + "where ufp.LocalUpdateID is not null "
		//isLeaf.Update("IsLeaf", true)
		db.Raw(isLeafSql).Scan(&isLeaf)
		db.Raw(isNotLeafSql).Scan(&isNotLeaf)
		var updateIsLeafSql string = fmt.Sprintf("UPDATE revision set IsLeaf = True WHERE RevisionID in (%s)", GenSqlStrString(isLeaf))
		var updateIsNotLeafSql string = fmt.Sprintf("UPDATE revision set IsLeaf = False WHERE RevisionID in (%s)", GenSqlStrString(isNotLeaf))
		db.Exec(updateIsLeafSql)
		db.Exec(updateIsNotLeafSql)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func ExportFile(filePath string, logName string) (err error) {
	var rootDir, fileName string
	if len(filePath) == 0 && runtime.GOOS != "windows" {
		rootDir = "/tmp/Export_Metadata/"
		fileName = "Export_Metadata.tar.gz"

	} else if len(filePath) == 0 && runtime.GOOS == "windows" {
		rootDir = "C:\\tmp\\Export_Metadata\\"
		fileName = "Export_Metadata.tar.gz"
	} else {
		if runtime.GOOS != "windows" {
			var arr = strings.Split(filePath, "/")
			var fileName = arr[len(arr)-2]
			rootDir = filePath[:len(filePath)-len(fileName)]
		} else {
			var arr = strings.Split(filePath, "\\")
			var fileName = arr[len(arr)-2]
			rootDir = filePath[:len(filePath)-len(fileName)]
		}
	}
	// 判断路径存在
	// 如果返回的错误为nil,说明文件或文件夹存在
	fileInfo, err := os.Stat(rootDir)
	if err != nil {
		os.Mkdir(rootDir, 0777)
	}
	// 检查路径是否有权限(检查文件mode是否有--x--x--x)
	fileMoad := fileInfo.Mode()
	perm := fileMoad.Perm()
	// --x--x--x 表示成十进制数就是73
	// 73: 00 001 001 001
	flag := perm & os.FileMode(73)
	if uint32(flag) == uint32(73) {
		// 允许执行
		ExportMetadata(rootDir)
		ExportPackage(rootDir, logName)
		utils.CompressFiles(rootDir, fileName)
	}
	os.Remove(rootDir)
	return
}
func ExportMetadata(rootDir string) {
	db := global.GDb
	var filePath = rootDir + "export/metadata.txt"

	_, err := os.Stat(rootDir + "export")
	//如果文件夹存在 则删除文件夹重新创建
	if err == nil {
		os.Remove(rootDir + "export")
	}
	os.Mkdir(rootDir+"export", 0777)
	var exportFile, _ = os.Create(filePath)
	var rules []response.Rules
	var revisionId []int
	var revisions []model.Revision
	var revisionSql = "select RevisionID, RootElementXml, RootElementXmlCompressed from rules where RootElementType = 0 order by RevisionID;"
	db.Raw(revisionSql).Scan(&rules)
	db.Table("revision")
	for i := 0; i < len(rules); i++ {
		revisionId = append(revisionId, rules[i].RevisionID)
	}
	db.Table("revision").Where("RevisionID in (?)", revisionId).Scan(&revisions)

	revisionRulesMap := make(map[int]map[string]interface{}, 0)
	for i := 0; i < len(revisions); i++ {
		revisionRulesMap[revisions[i].RevisionID] = map[string]interface{}{
			"update_id":       strings.ToLower(revisions[i].UpdateID),
			"revision_number": revisions[i].RevisionNumber,
		}
	}

	for i := 0; i < len(rules); i++ {
		var revision = revisionRulesMap[rules[i].RevisionID]
		var updateId = revision["update_id"]
		var revisionNumber = revision["revision_number"]
		var ruleXml string
		if len(rules[i].RootElementXml) > 0 {
			ruleXml = rules[i].RootElementXml
		} else {
			ruleXml, _ = utils.XmlUnCompress(rules[i].RootElementXmlCompressed)
		}
		var xmlLen = len(ruleXml) - 1

		var hexRevisionNumber = fmt.Sprintf("%x", revisionNumber)
		var revisionNumberBuffer bytes.Buffer
		for j := 0; j < 8-len(hexRevisionNumber); j++ {
			revisionNumberBuffer.WriteString("0")
		}
		hexRevisionNumber = revisionNumberBuffer.String() + hexRevisionNumber

		var hexXmlLen = fmt.Sprintf("%x", xmlLen)
		var xmlLenBuffer bytes.Buffer
		for j := 0; j < 8-len(hexXmlLen); j++ {
			xmlLenBuffer.WriteString("0")
		}
		hexXmlLen = xmlLenBuffer.String() + hexXmlLen

		var prefixStr = fmt.Sprintf("%s", updateId) + "," + hexRevisionNumber + "," + hexXmlLen + ","
		var updateStr = prefixStr + ruleXml
		if updateStr[len(updateStr)-1:] == ">" {
			updateStr = updateStr + "\n"
		}
		_, err := exportFile.WriteString(updateStr) //写入文件(字符串)
		if err != nil {
			return
		}
	}
	return

}

func ExportPackage(rootDir, logName string) {
	//生成metadata.txt时 已创建export文件夹
	var filePath = rootDir + "export/package.xml"
	exportFile, _ := os.Create(filePath)
	db := global.GDb
	var exportPackage serializer.ExportPackage
	exportPackage.ServerID = service.GetGlobalServerConfig().ServerID
	exportPackage.CreationTime = time.Now().UTC().String()
	exportPackage.FormatVersion = "1.0"
	exportPackage.ProtocolVersion = "1.20"

	//处理languages模块
	var languages serializer.Languages
	db.Table("update_language").Scan(&languages.Language)
	exportPackage.Languages = languages

	//处理files模块
	var files serializer.Files
	db.Table("files").Scan(&files.File)
	exportPackage.Files = files
	//处理updates模块
	//TODO: 暂时没有考虑metadata压缩的问题
	var updates serializer.Updates
	var exportLogUpdates serializer.ExportLogUpdates
	var exportLogUpdateFile serializer.ExportLogFiles
	var updateSql = "select r.RevisionID, r.UpdateID as UpdateId, r.RevisionNumber, f.FileDigest, ric.Type, ric.Value from revision r left join files f on r.RevisionID = f.RevisionID left join revision_in_category ric on r.RevisionID = ric.RevisionID order by r.RevisionID"
	db.Raw(updateSql).Scan(&updates.Update)
	for i := 0; i < len(updates.Update); i++ {
		//updates.Update[i].Categories.Category = "Category"
		var updateID = updates.Update[i].UpdateId
		var revisionNumber = updates.Update[i].RevisionNumber
		var filesSql = fmt.Sprintf("select f.FileDigest from files f join revision r on f.RevisionID=r.RevisionID  where r.UpdateID ='%s' and r.RevisionNumber =%s", updateID, revisionNumber)
		var categoriesSql = fmt.Sprintf("select ric.Value from revision_in_category ric join revision r on ric.RevisionID=r.RevisionID  where ric.Type='Category' and r.UpdateID ='%s' and r.RevisionNumber =%s", updateID, revisionNumber)
		var classificationsSql = fmt.Sprintf("select ric.Value from revision_in_category ric join revision r on ric.RevisionID=r.RevisionID  where ric.Type='Classification' and r.UpdateID ='%s' and r.RevisionNumber =%s", updateID, revisionNumber)
		db.Raw(filesSql).Scan(&updates.Update[i].Files.File)
		db.Raw(categoriesSql).Scan(&updates.Update[i].Categories.Category)
		db.Raw(classificationsSql).Scan(&updates.Update[i].Classifications.Classification)
		//	log
		if len(logName) > 0 {
			var exportLogSql = fmt.Sprintf("SELECT p.Title, r.UpdateID, r.RevisionNumber from property p join revision r on p.RevisionID =r.RevisionID  where r.UpdateID='%s' and r.RevisionNumber =%s and p.Language in ('zh-cn', 'en') ORDER BY p.`Language` desc limit 1", updateID, revisionNumber)
			var exportLogFileSql = fmt.Sprintf("SELECT f.FileDigest, f.FileName from test2827.files f where f.UpdateID='%s' and f.RevisionNumber =%s", updateID, revisionNumber)
			db.Raw(exportLogSql).Scan(&exportLogUpdates.Update[i].Update)
			db.Raw(exportLogFileSql).Scan(&exportLogUpdateFile)
		}
	}
	exportPackage.Updates = updates
	if xmlIndentByteData, err2 := xml.MarshalIndent(exportPackage, "", "  "); err2 == nil {
		strData := string(xmlIndentByteData)
		fmt.Println(strData)
		fmt.Println(strings.Replace(strData, "RedPacketQueryRequest", "xml", -1))
		_, err := exportFile.WriteString(strData) //写入文件(字符串)
		if err != nil {
			return
		}
	}

}
