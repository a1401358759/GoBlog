package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"goblog/core/global"
	"goblog/core/response"
	"goblog/modules/model"
	"goblog/service"
	"goblog/service/download"
	"goblog/utils"
	"goblog/utils/goset"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	GroupWithOutUSSSql = `select ctg.TargetGroupID,
								 ctg.ParentGroupID,
								 ctg.TargetGroupName,
								 ctg.Description,
								 ctg.OrderValue,
								 IfNULL(i.count, 0) as ComputerCount
						  from computer_target_group ctg
						  left join (select TargetGroupID, count(TargetGroupID) as count from computer_in_group group by TargetGroupID) i
                          on ctg.TargetGroupID = i.TargetGroupID
						  where ctg.TargetGroupID != 'D374F42A-9BE2-4163-A0FA-3C86A401B7A7'
						  order by ctg.OrderValue`
	ComputerGroupsSql = `
						select *
						from (
								 select g.TargetGroupID,
										g.TargetGroupName,
										g.Description,
										if(ii.TargetID is null, 'false', 'true')            as Status,
										ii.TargetID,
										if(l.computers_count is null, 0, l.computers_count) as ComputerCount
								 from computer_target_group g
										  left join computer_in_group ii on g.TargetGroupID = ii.TargetGroupID
									 and ii.TargetID = % d
										  left join (select i.TargetGroupID, count(c.TargetID) as computers_count
													 from computer_target c
															  left outer JOIN computer_in_group i on c.TargetID = i.TargetID
													 group by i.TargetGroupID) l on l.TargetGroupID = g.TargetGroupID)
								 as result
						where result.TargetID = % d
						`
	ComputerUpdateSql = `
						select *
						from (
								 select software.RevisionID                                                                                 as RevisionID,
										software.UpdateID                                                                                   as UpdateID,
										software.RevisionNumber                                                                             as RevisionNumber,
										if(prop.Title is not null, prop.Title, (select Title
																				from property
																				where Language = 'en' and RevisionID = software.RevisionID
																				limit 1))                                                   as Title,
										software.ProductRevisionID                                                                          as ProductRevisionID,
										software.ClassificationRevisionID                                                                   as ClassificationRevisionID,
										software.MsrcSeverity,
										DATE_FORMAT(software.ImportedTime, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00')                                 as ImportedTime,
										DATE_FORMAT(software.CreationDate, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00')                                 as CreationDate,
										software.KBArticleID,
										dep.InstalledCount,
										if(us.TargetID = % d, us.Status, 0)                                                                 as Status
								 from (select RevisionID,
											  UpdateID,
											  RevisionNumber,
											  ProductRevisionID,
											  ClassificationRevisionID,
											  MsrcSeverity,
											  ImportedTime,
											  CreationDate,
											  KBArticleID
									   from revision
									   where ProductRevisionID is not null
										 and ClassificationRevisionID is not null
										 and UpdateType not in ('Category', 'Detectoid')) software
										  left join (select * from property where Language = 'zh-cn') prop
													on prop.RevisionID = software.RevisionID
										  join (select RevisionID, count(RevisionID) as InstalledCount
												from deployment
												where ActionID = 0
												group by RevisionID
												union
												select DISTINCT RevisionID, 0 as InstalledCount
												from deployment
												where ActionID != 0
												  and RevisionID
													not in (select RevisionID from deployment where ActionID = 0)
												group by RevisionID) dep on software.RevisionID = dep.RevisionID
										  left join (select Status, RevisionID, TargetID
													 from update_status_per_computer
													 where TargetID = % d) us on software.RevisionID = us.RevisionID) result
                        `
	ComputerUpdateCountSql = `
								select count(1)
								from (
										 select software.RevisionID                                                                                 as RevisionID,
												software.UpdateID                                                                                   as UpdateID,
												software.RevisionNumber                                                                             as RevisionNumber,
												if(prop.Title is not null, prop.Title, (select Title
																						from property
																						where Language = 'en' and RevisionID = software.RevisionID
																						limit 1))                                                   as Title,
												software.ProductRevisionID                                                                          as ProductRevisionID,
												software.ClassificationRevisionID                                                                   as ClassificationRevisionID,
												software.MsrcSeverity,
												DATE_FORMAT(software.ImportedTime, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00')                                 as ImportedTime,
												DATE_FORMAT(software.CreationDate, '%%Y-%%m-%%dT%%H:%%i:%%s+00:00')                                 as CreationDate,
												software.KBArticleID,
												dep.InstalledCount,
												if(us.TargetID = % d, us.Status, 0)                                                                 as InstallStatus
										 from (select RevisionID,
													  UpdateID,
													  RevisionNumber,
													  ProductRevisionID,
													  ClassificationRevisionID,
													  MsrcSeverity,
													  ImportedTime,
													  CreationDate,
													  KBArticleID
											   from revision
											   where ProductRevisionID is not null
												 and ClassificationRevisionID is not null
												 and UpdateType not in ('Category', 'Detectoid')) software
												  left join (select * from property where Language = 'zh-cn') prop
															on prop.RevisionID = software.RevisionID
												  join (select RevisionID, count(RevisionID) as InstalledCount
														from deployment
														where ActionID = 0
														group by RevisionID
														union
														select DISTINCT RevisionID, 0 as InstalledCount
														from deployment
														where ActionID != 0
														  and RevisionID
															not in (select RevisionID from deployment where ActionID = 0)
														group by RevisionID) dep on software.RevisionID = dep.RevisionID
												  left join (select Status, RevisionID, TargetID
															 from update_status_per_computer
															 where TargetID = % d) us on software.RevisionID = us.RevisionID) result
                        `
	GroupDetailSql = `
                    select ctg.TargetGroupName, ctg.Description, ifNull(i.count, 0) as computers_count
					from computer_target_group ctg
							 left join (select TargetGroupID, count(TargetGroupID) as count from computer_in_group group by TargetGroupID) i
									   on ctg.TargetGroupID = i.TargetGroupID
					where ctg.TargetGroupID = '%s'
                     `
	GroupRelatedComputerSql = `
					select FullDomainName, TargetID, IPAddress, OSVersion, ComputerMake, LastReportedStatusTime
					from computer_target
					where TargetID in (select TargetID from computer_in_group where TargetGroupID %s '%s')
					`
	GroupRelatedComputerCountSql = `
					select count(1)
					from computer_target
					where TargetID in (select TargetID from computer_in_group where TargetGroupID %s '%s')
					`
)

func delRulesGroup(groupIds []string) error {
	var rules []map[string]interface{}
	if err := global.GDb.Table("custom_report_rules").Select("id, screen").Where("built_in is false").Where("service_type = ?", 1).Scan(&rules).Error; err != nil {
		global.GLog.Error("query custom_report_rules error,", zap.Any("err", err))
		return err
	}
	for _, rule := range rules {
		var screen map[string]map[string][]string
		_ = json.Unmarshal([]byte(rule["screen"].(string)), &screen)
		if inFilter, ok := screen["in"]; ok {
			if targetGroupId, ok := inFilter["TargetGroupID"]; ok {
				// 将被删除的计算机组和报表本身存储的计算机组做交集
				newGroupIds := goset.StrIntersect(targetGroupId, groupIds).List()
				if len(newGroupIds) == 0 {
					delete(screen["in"], "TargetGroupID")
					// 如果删除TargetGroupID后，整个in中不包含别的筛选条件，则将整个in删除
					if len(screen["in"]) == 0 {
						delete(screen, "in")
					}
				} else {
					screen["in"]["TargetGroupID"] = newGroupIds
				}
				screenStr, err := json.Marshal(screen)
				if err != nil {
					return err
				}
				if err := global.GDb.Table("custom_report_rules").Where("id=?", rule["id"]).Update("screen", string(screenStr)).Error; err != nil {
					global.GLog.Error("update custom_report_rules error,", zap.Any("err", err))
					return err
				}
			}
		}
	}
	return nil
}

func ComputerTargetList(pageSize, pageNumber int, filter []utils.QueryFilter, sort string) ([]map[string]interface{}, int) {
	var computers []map[string]interface{}
	var count int64
	pageSize = utils.If(pageSize == 0, 15, pageSize).(int)
	pageNumber = utils.If(pageNumber == 0, 1, pageNumber).(int)
	query := global.GDb.Table("computer_target").
		Select("TargetID, FullDomainName, ComputerMake, IPAddress, OSVersion, LastSyncTime, LastReportedStatusTime")
	countQuery := global.GDb.Table("computer_target")
	if len(filter) > 0 {
		for _, f := range filter {
			query = query.Where(f.Handle())
			countQuery = countQuery.Where(f.Handle())
		}
	}
	if len(sort) > 0 {
		query = query.Order(SortCondition(sort))
	}
	countQuery.Count(&count)
	query = query.Limit(pageSize).Offset((pageNumber - 1) * pageSize)
	query.Scan(&computers)
	return computers, int(count)
}

func ComputerDetail(targetId int) (map[string]interface{}, error) {
	computer := make(map[string]interface{})
	err := global.GDb.Table("computer_target").Select("TargetID, ComputerID, IPAddress, OSVersion, OSLocale, "+
		"ProcessorArchitecture, ClientVersion, ComputerMake, ComputerModel, FirmwareVersion, BiosName, BiosVersion, "+
		"MobileOperator, BiosReleaseDate, LastSyncTime, LastReportedStatusTime").Where("TargetID=?", targetId).
		Scan(&computer).Error
	if err != nil {
		return computer, err
	}
	return computer, nil
}

func GroupRecursion(groupDataMap map[string]map[string]interface{}, groupTreeMap map[string][]string, root string) interface{} {
	data := groupDataMap[root]
	childIds := groupTreeMap[root]
	if len(childIds) > 0 {
		var children []interface{}
		for _, childId := range childIds {
			children = append(children, GroupRecursion(groupDataMap, groupTreeMap, childId))
		}
		return map[string]interface{}{
			"TargetGroupID":   data["TargetGroupID"],
			"ParentGroupID":   data["ParentGroupID"],
			"TargetGroupName": data["TargetGroupName"],
			"Description":     data["Description"],
			"OrderValue":      data["OrderValue"],
			"ComputerCount":   data["ComputerCount"],
			"Children":        children,
		}
	}
	return data
}

func GroupTree() interface{} {
	var groups []map[string]interface{}
	var treeMap = make(map[string][]string)
	var keyMap = make(map[string]map[string]interface{})
	global.GDb.Raw(GroupWithOutUSSSql).Scan(&groups)
	for _, group := range groups {
		if group["ParentGroupID"] != nil && len(group["ParentGroupID"].(string)) > 0 {
			treeMap[group["ParentGroupID"].(string)] = append(treeMap[group["ParentGroupID"].(string)], group["TargetGroupID"].(string))
		}
		keyMap[group["TargetGroupID"].(string)] = group
	}
	rootGroupId := utils.UUIDAllComputer
	return GroupRecursion(keyMap, treeMap, rootGroupId)
}

func GetGroupSimpleList(pageSize, pageNumber int) ([]map[string]interface{}, error) {
	querySql := `select TargetGroupID,                                                               -- 组ID
					    TargetGroupName,                                                             -- 组名称
					    Description,                                                                 -- 组描述
					    (select count(1)
					     from computer_in_group
					     where computer_target_group.TargetGroupID = TargetGroupID) as ComputerCount -- 组内计算机数
			     from computer_target_group
			     where TargetGroupID != 'D374F42A-9BE2-4163-A0FA-3C86A401B7A7'`
	pageSize = utils.If(pageSize > 0, pageSize, 15).(int)
	pageNumber = utils.If(pageNumber > 0, pageNumber, 1).(int)
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	var groups []map[string]interface{}
	if err := global.GDb.Raw(querySql).Scan(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func GetGroupDetail(groupId string) map[string]interface{} {
	var group = make(map[string]interface{})
	querySql := fmt.Sprintf(GroupDetailSql, groupId)
	global.GDb.Raw(querySql).Scan(&group)
	return group
}

func GetGroupRelatedComputer(pageSize, pageNumber int, sort string, filter []utils.QueryFilter, groupId string) ([]map[string]interface{}, int) {
	var computers []map[string]interface{}
	status := true
	if len(filter) > 0 {
		for _, f := range filter {
			if f.Name == "Status" {
				status = f.Val.(bool)
			}
		}
	}
	op := "="
	if !status {
		op = "!="
	}
	querySql := fmt.Sprintf(GroupRelatedComputerSql, op, groupId)
	countSql := fmt.Sprintf(GroupRelatedComputerCountSql, op, groupId)
	querySql += "software." + SortCondition(sort)
	pageSize = utils.If(pageSize > 0, pageSize, 15).(int)
	pageNumber = utils.If(pageNumber > 0, pageNumber, 1).(int)
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	global.GDb.Raw(querySql).Scan(&computers)
	var count int
	global.GDb.Raw(countSql).Scan(&count)
	return computers, count
}

func GetGroupRelatedUpdate(pageSize, pageNumber int, filter []utils.QueryFilter, sort string, groupId string) ([]map[string]interface{}, int, int) {
	var wg sync.WaitGroup
	// 查询安装的revision
	InstallRevisionIdsChan := make(chan []map[string]interface{})
	go func(channel chan []map[string]interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		var InstallRevisionIds []map[string]interface{}
		if err := global.GDb.Table("revision").Joins("left join deployment on revision.RevisionID = deployment.RevisionID").Select("revision.RevisionID, deployment.ActionID").
			Where("revision.UpdateType not in ('Category', 'Detectoid')").
			Where("revision.ClassificationRevisionID is not null").
			Where("revision.ProductRevisionID is not null").
			Where("deployment.TargetGroupID = ?", groupId).
			Where("deployment.ActionID = ?", utils.InstallAction).Scan(&InstallRevisionIds).Error; err != nil {
		}
		channel <- InstallRevisionIds
	}(InstallRevisionIdsChan, &wg)
	// 查询卸载的revision
	UnInstallRevisionIdsChan := make(chan []map[string]interface{})
	go func(channel chan []map[string]interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		var UnInstallRevisionIds []map[string]interface{}
		global.GDb.Table("revision").Joins("left join deployment on revision.RevisionID = deployment.RevisionID").Select("revision.RevisionID, deployment.ActionID").
			Where("revision.UpdateType not in ('Category', 'Detectoid')").
			Where("revision.ClassificationRevisionID is not null").
			Where("revision.ProductRevisionID is not null").
			Where("deployment.TargetGroupID = ?", groupId).
			Where("deployment.ActionID = ?", utils.UninstallAction).Scan(&UnInstallRevisionIds)
		channel <- UnInstallRevisionIds
	}(UnInstallRevisionIdsChan, &wg)
	// 查询拒绝的revision
	DeclineRevisionIdsChan := make(chan []map[string]interface{})
	go func(channel chan []map[string]interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		var DeclineRevisionIds []map[string]interface{}
		global.GDb.Table("revision").Joins("left join deployment on revision.RevisionID = deployment.RevisionID").Select("revision.RevisionID, deployment.ActionID").
			Where("revision.UpdateType not in ('Category', 'Detectoid')").
			Where("revision.ClassificationRevisionID is not null").
			Where("revision.ProductRevisionID is not null").
			Where("deployment.TargetGroupID = ?", utils.UUIDAllComputer).
			Where("deployment.ActionID = ?", utils.DeclineAction).Scan(&DeclineRevisionIds)
		channel <- DeclineRevisionIds
	}(DeclineRevisionIdsChan, &wg)
	// 查询未审批的有明确关联的revision
	UnApprovedRevisionIdsChan := make(chan []map[string]interface{})
	go func(channel chan []map[string]interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		wg.Add(1)
		var UnApprovedRevisionIds []map[string]interface{}
		global.GDb.Table("revision").Joins("left join deployment on revision.RevisionID = deployment.RevisionID").Select("revision.RevisionID, deployment.ActionID").
			Where("revision.UpdateType not in ('Category', 'Detectoid')").
			Where("revision.ClassificationRevisionID is not null").
			Where("revision.ProductRevisionID is not null").
			Where("deployment.TargetGroupID = ?", groupId).
			Where("deployment.ActionID in (" + strconv.Itoa(utils.PreDeploymentCheckAction) + "," + strconv.Itoa(utils.BlockAction) + ")").Scan(&UnApprovedRevisionIds)
		channel <- UnApprovedRevisionIds
	}(UnApprovedRevisionIdsChan, &wg)

	wg.Wait()

	InstalledRevisionIds := <-InstallRevisionIdsChan
	UnInstalledRevisionIds := <-UnInstallRevisionIdsChan
	DeclineRevisionIds := <-DeclineRevisionIdsChan
	UnApprovedRevisionIds := <-UnApprovedRevisionIdsChan
	var DelRevisionIds []map[string]interface{}
	DelRevisionIds = append(DelRevisionIds, InstalledRevisionIds...)
	DelRevisionIds = append(DelRevisionIds, UnInstalledRevisionIds...)
	DelRevisionIds = append(DelRevisionIds, DeclineRevisionIds...)
	DelRevisionIds = append(DelRevisionIds, UnApprovedRevisionIds...)
	var DelRevisionIdConditions []string
	for _, revisionAction := range DelRevisionIds {
		DelRevisionIdConditions = append(DelRevisionIdConditions, strconv.Itoa(utils.ToInt(revisionAction["RevisionID"])))
	}
	var OtherUnApprovedRevisionIds []map[string]interface{}
	global.GDb.Table("revision").Joins("left join deployment on revision.RevisionID = deployment.RevisionID").Select("revision.RevisionID, deployment.ActionID").
		Where("revision.UpdateType not in ('Category', 'Detectoid')").
		Where("revision.ClassificationRevisionID is not null").
		Where("revision.ProductRevisionID is not null").
		Where("deployment.TargetGroupID = ?", utils.UUIDAllComputer).
		Where("deployment.ActionID in (" + strconv.Itoa(utils.PreDeploymentCheckAction) + "," + strconv.Itoa(utils.BlockAction) + ")").
		Where("revision.RevisionID not in (" + strings.Join(DelRevisionIdConditions, ",") + ")").Scan(&OtherUnApprovedRevisionIds)
	UnApprovedRevisionIds = append(UnApprovedRevisionIds, OtherUnApprovedRevisionIds...)
	// 处理筛选条件
	var preQuery []map[string]interface{}
	var queryScope []int
	if len(filter) > 0 {
		for _, f := range filter {
			if f.Name == "ApprovalStatus" {
				if reflect.TypeOf(f.Val).Kind() == reflect.Slice {
					for i := 0; i < reflect.ValueOf(f.Val).Len(); i++ {
						v := reflect.ValueOf(f.Val).Index(i)
						queryScope = append(queryScope, utils.ToInt(v.Interface()))
					}
				}
			}
		}
	} else {
		queryScope = []int{utils.InstallAction, utils.UninstallAction, utils.PreDeploymentCheckAction, utils.BlockAction, utils.DeclineAction}
	}
	if utils.ContainInt(utils.InstallAction, queryScope) {
		preQuery = append(preQuery, InstalledRevisionIds...)
	}
	if utils.ContainInt(utils.UninstallAction, queryScope) {
		preQuery = append(preQuery, UnInstalledRevisionIds...)
	}
	if utils.ContainInt(utils.DeclineAction, queryScope) {
		preQuery = append(preQuery, DeclineRevisionIds...)
	}
	if utils.ContainInt(utils.PreDeploymentCheckAction, queryScope) ||
		utils.ContainInt(utils.BlockAction, queryScope) ||
		utils.ContainInt(utils.EvaluateAction, queryScope) ||
		utils.ContainInt(utils.BundleAction, queryScope) ||
		utils.ContainInt(utils.DssAction, queryScope) {
		preQuery = append(preQuery, UnApprovedRevisionIds...)
	}
	var revisionConditions []string
	caseSql := ""
	if len(preQuery) > 0 {
		caseSql += " ,case "
		for _, q := range preQuery {
			caseSql += "when software.RevisionID = " + strconv.Itoa(utils.ToInt(q["RevisionID"])) + " then " + strconv.Itoa(utils.ToInt(q["ActionID"])) + " "
			revisionConditions = append(revisionConditions, strconv.Itoa(utils.ToInt(q["RevisionID"])))
		}
		caseSql += " else 2 end ApprovalStatus "
	} else {
		caseSql += ", 2 as ApprovalStatus"
	}
	querySql := fmt.Sprintf(BaseRevisionSql, caseSql)
	countSql := RevisionCountSql
	if len(revisionConditions) > 0 {
		querySql += " and software.RevisionID in (" + strings.Join(revisionConditions, ",") + ") "
		countSql += " and RevisionID in (" + strings.Join(revisionConditions, ",") + ") "
	} else {
		querySql += " and software.RevisionID = 0"
		countSql += " and RevisionID = 0 "
	}
	var sortSql string
	if len(sort) > 0 {
		sortSql = " order by "
		sortSql += "software." + SortCondition(sort)
	} else {
		sortSql = " order by software.RevisionID DESC "
	}
	querySql += sortSql
	pageSize = utils.If(pageSize > 0, pageSize, 15).(int)
	pageNumber = utils.If(pageNumber > 0, pageNumber, 1).(int)
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	var revisions []map[string]interface{}
	global.GDb.Raw(querySql).Scan(&revisions)
	var count int64
	global.GDb.Raw(countSql).Scan(&count)
	var groupCount int64
	global.GDb.Raw(GroupCountSql).Scan(&groupCount)
	return revisions, int(count), int(groupCount)
}

func GroupDistribute(action bool, groupId string, computerIds []int) string {
	var group model.ComputerTargetGroup
	global.GDb.Where("TargetGroupID=?", groupId).First(&group)
	if action {
		var unassignedComputers []int
		global.GDb.Table("computer_in_group").Select("TargetID").Where("TargetGroupID=?", utils.UUIDGroupUnassigned).Scan(&unassignedComputers)
		intersectComputers := goset.StrIntersect(utils.SliceIntToString(unassignedComputers), utils.SliceIntToString(computerIds)).List()
		global.GDb.Table("computer_in_group").Where("TargetID in ("+strings.Join(intersectComputers, ",")+")").
			Where("TargetGroupID=?", utils.UUIDGroupUnassigned).Update("TargetGroupID", groupId)
		differentComputerIds := goset.StrMinus(utils.SliceIntToString(computerIds), utils.SliceIntToString(unassignedComputers)).List()
		var differentComputers []model.ComputerTarget
		global.GDb.Where("TargetID in (" + strings.Join(differentComputerIds, ",") + ")").Find(&differentComputers)
		var newComputerInGroup []model.ComputerInGroup
		for _, differentComputer := range differentComputers {
			newComputerInGroup = append(newComputerInGroup, model.ComputerInGroup{
				TargetGroupID: groupId,
				TargetID:      differentComputer.TargetID,
				ComputerID:    differentComputer.ComputerID,
			})
		}
		global.GDb.CreateInBatches(&newComputerInGroup, 1000)
	} else {
		global.GDb.Where("TargetID in ("+strings.Join(utils.SliceIntToString(computerIds), ",")+")").Where("TargetGroupID=?", groupId).Delete(&model.ComputerInGroup{})
		var allComputerIds []int
		global.GDb.Table("computer_in_group").Select("TargetID").Scan(&allComputerIds)
		differentComputerIds := goset.StrMinus(utils.SliceIntToString(computerIds), utils.SliceIntToString(allComputerIds)).List()
		var differentComputers []model.ComputerTarget
		global.GDb.Where("TargetID in (" + strings.Join(differentComputerIds, ",") + ")").Find(&differentComputers)
		var newComputerInGroup []model.ComputerInGroup
		for _, differentComputer := range differentComputers {
			newComputerInGroup = append(newComputerInGroup, model.ComputerInGroup{
				TargetGroupID: utils.UUIDGroupUnassigned,
				TargetID:      differentComputer.TargetID,
				ComputerID:    differentComputer.ComputerID,
			})
		}
		global.GDb.CreateInBatches(&newComputerInGroup, 1000)
	}
	return group.TargetGroupName
}

func CreateGroup(TargetGroupName, ParentGroupID, Description string) (error, string) {
	sc := service.GetGlobalServerConfig()
	if sc.ReplicationMode == service.SyncModeReplica {
		return errors.New("副本模式无法进行计算机组操作。"), ""
	}
	if ParentGroupID == utils.UUIDGroupUnassigned {
		return errors.New("未分配组不可以创建子组。"), ""
	}
	var nameExists int64
	global.GDb.Table("computer_target_group").Where("TargetGroupName=?", TargetGroupName).Count(&nameExists)
	if nameExists > 0 {
		return errors.New("该计算机组名称已经存在。"), ""
	}
	var parentGroup model.ComputerTargetGroup
	global.GDb.Where("TargetGroupID=?", ParentGroupID).First(&parentGroup)
	global.GDb.Create(model.ComputerTargetGroup{
		TargetGroupID:   strings.ToUpper(uuid.New().String()),
		TargetGroupName: TargetGroupName,
		Description:     Description,
		ParentGroupID:   ParentGroupID,
		IsBuiltin:       false,
		OrderValue:      parentGroup.OrderValue + 1,
	})
	return nil, parentGroup.TargetGroupName
}

func EditGroupName(groupId, groupName string) (error, string) {
	var nameExists int64
	global.GDb.Table("computer_target_group").Where("TargetGroupName=?", groupName).Count(&nameExists)
	if nameExists > 0 {
		return errors.New("该计算机组名称已经存在。"), ""
	}
	var group model.ComputerTargetGroup
	global.GDb.Where("TargetGroupID=?", groupId).First(&group)
	oldName := group.TargetGroupName
	group.TargetGroupName = groupName
	global.GDb.Save(&group)
	return nil, oldName
}

func getChildGroupId(parentGroupIds []string) (childStrIds, childIds []string) {
	for {
		var newParentGroupIds []string
		global.GDb.Table("computer_target_group").Select("TargetGroupID").Where("ParentGroupID in (" + strings.Join(parentGroupIds, ",") + ")").Scan(&newParentGroupIds)
		if len(newParentGroupIds) == 0 {
			break
		} else {
			var newParentGroupStrIds []string
			for _, i := range newParentGroupIds {
				newParentGroupStrIds = append(newParentGroupStrIds, "'"+i+"'")
				childStrIds = append(childStrIds, "'"+i+"'")
				childIds = append(childIds, i)
			}
			parentGroupIds = newParentGroupStrIds
		}
	}
	return
}

func DelGroup(groupId string) error {
	var parenGroupIds []string
	parenGroupIds = append(parenGroupIds, "'"+groupId+"'")
	children, childIds := getChildGroupId(parenGroupIds)
	children = append(children, "'"+groupId+"'")
	var relatedComputers []struct {
		TargetID   int
		ComputerID string
	}
	var groupIds []string
	groupIds = append(groupIds, groupId)
	groupIds = append(groupIds, childIds...)
	var err error
	if err = global.GDb.Table("computer_in_group").Where("TargetGroupID in (" + strings.Join(children, ",") + ")").Select("TargetID, ComputerID").Scan(&relatedComputers).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetGroupID in (" + strings.Join(children, ",") + ")").Delete(&model.ComputerInGroup{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetGroupID in (" + strings.Join(children, ",") + ")").Delete(&model.Deployment{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetGroupID in (" + strings.Join(children, ",") + ")").Delete(&model.DeadDeployment{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetGroupID in (" + strings.Join(children, ",") + ")").Delete(&model.RevisionStatement{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetGroupID in (" + strings.Join(children, ",") + ")").Delete(&model.ComputerTargetGroup{}).Error; err != nil {
		return err
	}
	var newRelated []model.ComputerInGroup
	for _, relatedComputer := range relatedComputers {
		var relatedExists int64
		global.GDb.Table("computer_in_group").Where("TargetID = ?", relatedComputer.TargetID).Count(&relatedExists)
		if relatedExists == 0 {
			newRelated = append(newRelated, model.ComputerInGroup{
				TargetGroupID: utils.UUIDGroupUnassigned,
				TargetID:      relatedComputer.TargetID,
				ComputerID:    relatedComputer.ComputerID,
			})
		}
	}
	global.GDb.CreateInBatches(&newRelated, 1000)
	var childrenInterface []interface{}
	for _, c := range children {
		childrenInterface = append(childrenInterface, c)
	}
	if err, _ = utils.CallProc("set_invalid_rule", childrenInterface...); err != nil {
		return err
	}
	if err = global.GDb.Table("custom_report_rules").
		Where("target_group_id in ("+strings.Join(children, ",")+")").
		Where("service_type=5").
		Where("built_in is false").
		Update("is_valid", false).Error; err != nil {
		return err
	}
	if err := delRulesGroup(groupIds); err != nil {
		return err
	}
	return nil
}

func GroupDistributeAll(groupId string) (string, int) {
	var group model.ComputerTargetGroup
	global.GDb.Where("TargetGroupID=?", groupId).First(&group)
	var unassignedComputers []int
	global.GDb.Table("computer_in_group").Select("TargetID").Where("TargetGroupID=?", utils.UUIDGroupUnassigned).Scan(&unassignedComputers)
	db := global.GDb.Table("computer_in_group").Where("TargetID in ("+strings.Join(utils.SliceIntToString(unassignedComputers), ",")+")").
		Where("TargetGroupID=?", utils.UUIDGroupUnassigned).Update("TargetGroupID", groupId)
	updateCount := int(db.RowsAffected)
	return group.TargetGroupName, updateCount
}

func GetComputerRelatedGroups(pageSize, pageNumber int, filter []utils.QueryFilter, sort string, targetId int) []map[string]interface{} {
	querySql := fmt.Sprintf(ComputerGroupsSql, targetId, targetId)
	if len(filter) > 0 {
		if len(filter) > 0 {
			for idx, f := range filter {
				if idx == 0 {
					querySql += " Where " + f.Handle()
				} else {
					querySql += " And " + f.Handle()
				}
			}
		}
	}
	var sortSql string
	if len(sort) > 0 {
		sortSql = " order by "
		sortSql += SortCondition(sort)
	} else {
		sortSql = " order by FIELD(result.TargetGroupName,'Unassigned Computers') desc "
	}
	querySql += sortSql
	pageSize = utils.If(pageSize > 0, pageSize, 15).(int)
	pageNumber = utils.If(pageNumber > 0, pageNumber, 1).(int)
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	var groups []map[string]interface{}
	global.GDb.Raw(querySql).Scan(&groups)
	return groups
}

func GetComputerRelatedUpdates(pageSize, pageNumber int, filter []utils.QueryFilter, sort string, targetId int) ([]map[string]interface{}, int, int) {
	querySql := fmt.Sprintf(ComputerUpdateSql, targetId, targetId)
	countSql := fmt.Sprintf(ComputerUpdateCountSql, targetId, targetId)
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
	} else {
		sortSql = " order by RevisionID DESC "
	}
	querySql += sortSql
	pageSize = utils.If(pageSize > 0, pageSize, 15).(int)
	pageNumber = utils.If(pageNumber > 0, pageNumber, 1).(int)
	querySql += fmt.Sprintf(" limit %d,%d", (pageNumber-1)*pageSize, pageSize)
	var revisions []map[string]interface{}
	var count int64
	var groupCount int64
	global.GDb.Raw(querySql).Scan(&revisions)
	global.GDb.Raw(countSql).Scan(&count)
	global.GDb.Raw(GroupCountSql).Scan(&groupCount)
	return revisions, int(count), int(groupCount) - 2
}

func ComputerDistributed(action bool, computerId string, groupIds []string) {
	var computerInGroups []model.ComputerInGroup
	var computer model.ComputerTarget
	global.GDb.Where("TargetID=" + computerId).First(&computer)
	if action {
		for _, groupId := range groupIds {
			computerInGroups = append(computerInGroups, model.ComputerInGroup{
				TargetGroupID: groupId,
				TargetID:      computer.TargetID,
				ComputerID:    computer.ComputerID,
			})
		}
		global.GDb.Where("TargetID=?", computer.TargetID).
			Where("TargetGroupID=", utils.UUIDGroupUnassigned).Delete(&model.ComputerInGroup{})
	} else {
		global.GDb.Where("TargetID=", computer.TargetID).
			Where("TargetGroupID in (" + strings.Join(groupIds, ",") + ")").Delete(&model.ComputerInGroup{})
		var relationExists int64
		global.GDb.Table("computer_in_group").Where("TargetID=", computer.TargetID).Count(&relationExists)
		if relationExists == 0 {
			global.GDb.Create(model.ComputerInGroup{
				TargetGroupID: utils.UUIDGroupUnassigned,
				TargetID:      computer.TargetID,
				ComputerID:    computer.ComputerID,
				UpdateTime:    time.Now().UTC(),
			})
		}
	}
}

func ComputerDelete(computerId int) error {
	var err error
	if err = global.GDb.Where("TargetID = ?", computerId).Delete(&model.UpdateStatusPerComputer{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID = ?", computerId).Delete(&model.ComputerSummaryForMicrosoftUpdates{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID = ?", computerId).Delete(&model.ComputerStatement{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID = ?", computerId).Delete(&model.ComputerRevisionInstallStats{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID = ?", computerId).Delete(&model.ComputerInGroup{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID = ?", computerId).Delete(&model.ComputerTarget{}).Error; err != nil {
		return err
	}
	return nil
}

func ComputerBulkDelete(computerIds []int) error {
	var err error
	if err = global.GDb.Where("TargetID in (" + strings.Join(utils.SliceIntToString(computerIds), ",") + ")").Delete(&model.UpdateStatusPerComputer{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID in (" + strings.Join(utils.SliceIntToString(computerIds), ",") + ")").Delete(&model.ComputerSummaryForMicrosoftUpdates{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID in (" + strings.Join(utils.SliceIntToString(computerIds), ",") + ")").Delete(&model.ComputerStatement{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID in (" + strings.Join(utils.SliceIntToString(computerIds), ",") + ")").Delete(&model.ComputerRevisionInstallStats{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID in (" + strings.Join(utils.SliceIntToString(computerIds), ",") + ")").Delete(&model.ComputerInGroup{}).Error; err != nil {
		return err
	}
	if err = global.GDb.Where("TargetID in (" + strings.Join(utils.SliceIntToString(computerIds), ",") + ")").Delete(&model.ComputerTarget{}).Error; err != nil {
		return err
	}
	return nil
}

func GetApproveRuleList(userID, pageNum, pageSize int) ([]map[string]interface{}, int64) {
	db := global.GDb
	var count int64
	// 计算总数
	db = db.Model(&model.AutoApproveRules{}).Where("user_id = ?", userID)
	db.Count(&count)
	// 分页
	var ruleList []map[string]interface{}
	db.Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&ruleList)

	return ruleList, count
}

// ApproveRuleCreate 自动审批规则创建
func ApproveRuleCreate(c *gin.Context) (code int, msg string, ruleID int) {
	userID, userEmail := GetUserInfo(c)
	params := ApproveRuleAdd{}
	err := c.BindJSON(&params)
	if err != nil {
		global.GLog.Error("ApproveRuleCreate", zap.Any("err", err))
		return response.FAILED, "参数获取失败", -1
	}

	minutesAfterMidnight := 0
	if params.MinutesAfterMidnight != "" {
		hourStr := strings.Split(params.MinutesAfterMidnight, ":")[0]
		minuteStr := strings.Split(params.MinutesAfterMidnight, ":")[1]
		hour, _ := strconv.Atoi(hourStr)
		minutes, _ := strconv.Atoi(minuteStr)
		minutesAfterMidnight = hour*60 + minutes
	}

	rule := model.AutoApproveRules{
		UserID:               userID,
		Name:                 params.Name,
		TargetGroups:         params.TargetGroups,
		TargetGroupNames:     params.TargetGroupNames,
		ProductIds:           params.ProductIds,
		ClassificationIds:    params.ClassificationIds,
		Action:               &params.Action,
		DateOffset:           params.DateOffset,
		MinutesAfterMidnight: minutesAfterMidnight,
	}

	db := global.GDb
	if err = db.Create(&rule).Error; err != nil {
		global.GLog.Error("ApproveRuleCreate", zap.Any("err", err))
		detail := "创建审批规则失败"
		GenOperateRecord(OperateAction.CreatApproval, detail, ResultStatus.FAILED, userID, userEmail)
		return response.FAILED, response.ErrorMsg[response.FAILED], -1
	} else {
		operateDesc := fmt.Sprintf("创建审批规则【%s】", params.Name)
		GenOperateRecord(OperateAction.CreatApproval, operateDesc, ResultStatus.SUCCESS, userID, userEmail)
		return response.SUCCESS, response.ErrorMsg[response.SUCCESS], rule.ID
	}
}

// ApproveRuleEdit 自动审批规则编辑
func ApproveRuleEdit(c *gin.Context) (code int, msg string, ruleID int) {
	userID, _ := GetUserInfo(c)
	params := ApproveRuleAdd{}
	err := c.BindJSON(&params)
	if err != nil {
		global.GLog.Error("ApproveRuleEdit", zap.Any("err", err))
		return response.FAILED, "参数获取失败", -1
	}

	minutesAfterMidnight := 0
	if params.MinutesAfterMidnight != "" {
		hourStr := strings.Split(params.MinutesAfterMidnight, ":")[0]
		minuteStr := strings.Split(params.MinutesAfterMidnight, ":")[1]
		hour, _ := strconv.Atoi(hourStr)
		minutes, _ := strconv.Atoi(minuteStr)
		minutesAfterMidnight = hour*60 + minutes
	}

	rule := model.AutoApproveRules{
		Name:                 params.Name,
		TargetGroups:         params.TargetGroups,
		TargetGroupNames:     params.TargetGroupNames,
		ProductIds:           params.ProductIds,
		ClassificationIds:    params.ClassificationIds,
		Action:               &params.Action,
		DateOffset:           params.DateOffset,
		MinutesAfterMidnight: minutesAfterMidnight,
	}

	db := global.GDb
	result := db.Where("id = ? and user_id = ?", params.RuleID, userID).Updates(&rule)
	if result.RowsAffected == 0 {
		return response.NotFound, response.ErrorMsg[response.NotFound], -1
	}
	if result.Error != nil {
		global.GLog.Error("ApproveRuleEdit", zap.Any("err", result.Error))
		return response.FAILED, response.ErrorMsg[response.FAILED], -1
	}
	return response.SUCCESS, response.ErrorMsg[response.SUCCESS], params.RuleID
}

// ApproveRuleDel 自动审批规则删除
func ApproveRuleDel(c *gin.Context) (code int, msg string) {
	userID, _ := GetUserInfo(c)
	var err error

	var params struct {
		RuleID []int `json:"ids"`
	}
	err = c.BindJSON(&params)
	if err != nil {
		global.GLog.Error("ApproveRuleDel", zap.Any("err", err))
		return response.FAILED, "参数获取失败"
	}

	db := global.GDb
	if err = db.Where("id in (?) and user_id = ?", params.RuleID, userID).Delete(&model.AutoApproveRules{}).Error; err != nil {
		global.GLog.Error("ApproveRuleDel", zap.Any("err", err))
		return response.FAILED, "删除审批规则失败"
	}
	return response.SUCCESS, response.ErrorMsg[response.SUCCESS]
}

// AutoApproveRuleAsync 执行自动审批
func AutoApproveRuleAsync(db *gorm.DB, ruleList []model.AutoApproveRules, adminName string, userID int, userEmail string) {
	var result = make([]map[string]interface{}, 0)

	// 获取当前存在的组
	var allGroupIDs []string
	if err := db.Model(&model.ComputerTargetGroup{}).Where("TargetGroupID not in (?)", []string{utils.UUIDAllComputer, utils.UUIDGroupDss}).Pluck("TargetGroupID", &allGroupIDs).Error; err != nil {
		global.GLog.Error("RunApproveRule", zap.Any("err", err))
	}

	for _, rule := range ruleList {
		// 如果规则里含有现在不存在的组则该规则失效，拒绝的时候不需要判断，拒绝默认是当前所有组
		targetGroupList := strings.Split(rule.TargetGroups, ",")
		if (*rule.Action != utils.DeclineAction && goset.StrMinus(targetGroupList, allGroupIDs).Count() > 0) || !rule.IsValid {
			result = append(result, map[string]interface{}{"name": rule.Name, "status": utils.ApprovalRuleStatus.INVALID, "success": 0, "failed": 0})
			continue
		}
		// 根据产品和分类获取所有Revision
		productIDs := strings.Split(rule.ProductIds, ",")
		classificationIDs := strings.Split(rule.ClassificationIds, ",")
		revisions := db.Model(&model.Revision{}).Select("RevisionID", "EulaID").Where("UpdateType not in ('Category', 'Detectoid') and ProductRevisionID in (?) and ClassificationRevisionID in (?)", productIDs, classificationIDs)
		if revisions.Error != nil {
			global.GLog.Error("RunApproveRule", zap.Error(revisions.Error))
		}
		// 如果规则未匹配到更新，则跳过本规则，按成功处理
		var revisionCount int64
		revisions.Count(&revisionCount)
		if revisionCount == 0 {
			result = append(result, map[string]interface{}{"name": rule.Name, "status": utils.ApprovalRuleStatus.SUCCESS, "success": 0, "failed": 0})
			continue
		}

		if utils.ContainInt(*rule.Action, []int{utils.InstallAction, utils.UninstallAction, utils.BlockAction}) {
			result = AutoApproveWithRule(rule, revisions, targetGroupList, adminName, result)
		} else if *rule.Action == utils.DeclineAction {
			var revisionIDs []int
			revisions.Pluck("RevisionID", &revisionIDs)
			result = BatchDeclineRevisions(&rule, revisionIDs, adminName, result)
		}
	}
	// 执行完成，更新redis信息
	resultJson, _ := json.Marshal(result)
	global.GRedis.HMSet("auto_approve_status", map[string]interface{}{"pending": 0, "result": string(resultJson), "end": time.Now().Format("2006-01-02 15:04:05")})
	// 记录操作日志
	detail := ""
	for _, item := range result {
		detail += fmt.Sprintf("按【%s】进行审批操作，成功：【%d】，失败：【%d】", item["name"], item["success"], item["failed"])
	}
	GenOperateRecord(OperateAction.RuleApproval, detail, ResultStatus.SUCCESS, userID, userEmail)
}

// AutoApproveWithRule 根据审批规则自动审批(安装，卸载，取消审批)
func AutoApproveWithRule(rule model.AutoApproveRules, revisions *gorm.DB, targetGroupList []string, adminName string, result []map[string]interface{}) []map[string]interface{} {
	db := global.GDb
	var err error

	successCount := 0
	failedCount := 0

	var canApproveRevisions []int       // 可以审批的Revision
	var canApproveBundleRevisions []int // 可以审批的Revision的BundleRevisionID

	err = db.Transaction(func(tx *gorm.DB) error {
		actionID := *rule.Action
		// 为了方便后面筛选，2，3都属于未审批
		filterActions := []int{actionID}
		// 所有的RevisionIDs
		var allRevisionIDs []int
		if err = revisions.Select("RevisionID").Pluck("RevisionID", &allRevisionIDs).Error; err != nil {
			return err
		}
		// 安装和卸载执行审批规则需剔除还未接受许可协议的更新
		if utils.ContainInt(actionID, []int{utils.InstallAction, utils.UninstallAction}) {
			eulaRevision := revisions.Where("EulaID != '' and EulaID is not null and EulaExplicitlyAccepted is not true")
			var eulaCount int64
			if err = eulaRevision.Count(&eulaCount).Error; err != nil {
				return err
			}
			if eulaCount > 0 {
				var eulaIDs []string
				if err = eulaRevision.Select("EulaID").Pluck("EulaID", &eulaIDs).Error; err != nil {
					return err
				}
				var acceptedEulaIDs []string
				if err = db.Model(&model.EulaAcceptance{}).Where("eula_id in (?)", eulaIDs).Select("eula_id").Pluck("eula_id", &acceptedEulaIDs).Error; err != nil {
					return err
				}
				var unacceptedRevisionIDs []int
				if err = eulaRevision.Where("EulaID not in (?)", acceptedEulaIDs).Select("RevisionID").Pluck("RevisionID", &unacceptedRevisionIDs).Distinct("RevisionID").Error; err != nil {
					return err
				}
				failedCount += len(unacceptedRevisionIDs) // 未接受许可协议计入不支持审批数量
				allRevisionIDs = goset.IntMinus(allRevisionIDs, unacceptedRevisionIDs).List()
			}
		}
		// 先筛选出所有不满足当前审批条件的RevisionID
		var notStatisfiedRevisions []int
		sql := fmt.Sprintf(`select distinct tmp.RevisionID
							from (select r.RevisionID,
										 c.TargetGroupID,
										 c.TargetGroupName
								  from revision r,
									   computer_target_group c
								  where r.RevisionID in (%s)
									and c.TargetGroupID in (%s)) as tmp
									 left outer join deployment d on tmp.RevisionID = d.RevisionID and d.TargetGroupID = tmp.TargetGroupID
							where d.ActionID not in (%s)
							   or d.ActionID is null;`, GenSqlStrInt(allRevisionIDs), GenSqlStrString(targetGroupList), GenSqlStrInt(filterActions))
		if err = db.Raw(sql).Scan(&notStatisfiedRevisions).Error; err != nil {
			return err
		}
		if len(notStatisfiedRevisions) == 0 {
			return nil
		}

		if actionID == utils.InstallAction {
			// 过期更新不允许审批到安装
			canSql := fmt.Sprintf("select RevisionID from revision where RevisionID in (%s) and PublicationState != 1", GenSqlStrInt(notStatisfiedRevisions))
			if err = revisions.Raw(canSql).Scan(&canApproveRevisions).Error; err != nil {
				return err
			}
		} else if actionID == utils.UninstallAction {
			// 获取所有Revision的BundleRevisionID，为了获取是否允许卸载属性
			var firstBundleRevision []int // 每个更新的第一个Bundle的RevisionID
			var noBundleRevision []int    // 没有Bundle的RevisionID
			bundleList, err1 := global.GRedis.HMGet("b_r", utils.IntMapStr(notStatisfiedRevisions)...).Result()
			if err != nil {
				return err1
			}
			for i, revisionID := range notStatisfiedRevisions {
				if bundleList[i] != nil {
					var revisionBundles []int
					_ = json.Unmarshal([]byte(reflect.ValueOf(bundleList[i]).String()), &revisionBundles)
					firstBundleRevision = append(firstBundleRevision, revisionBundles[0])
				} else {
					noBundleRevision = append(noBundleRevision, revisionID)
				}
			}
			// 没有Bundle的，CanUninstall=True属性在自己身上，获取可以卸载的RevisionID
			var usefulRevisions1 []int
			if err = db.Model(&model.Revision{}).Where("RevisionID in (?) and CanUninstall is true;", noBundleRevision).Pluck("RevisionID", &usefulRevisions1).Error; err != nil {
				return err
			}
			// 有Bundle的Revision他的CanUninstall属性在他的Bundle身上，获取可以卸载的BundleRevisionID，通过可以卸载的BundleRevisionID再去获取对应的RevisionID
			var usefulRevisions2 []int
			if len(firstBundleRevision) > 0 {
				sql2 := fmt.Sprintf("select distinct RevisionID from bundle where BundleRevisionID in (select RevisionID from revision where RevisionID in (%s) and CanUninstall is true);", GenSqlStrInt(firstBundleRevision))
				if err = db.Raw(sql2).Scan(&usefulRevisions2).Error; err != nil {
					return err
				}
			}
			canApproveRevisions = append(usefulRevisions1, usefulRevisions2...)
		} else if actionID == utils.BlockAction {
			canApproveRevisions = notStatisfiedRevisions
			filterActions = append(filterActions, utils.PreDeploymentCheckAction)
		} else {
			failedCount = 0
			return nil
		}

		failedCount += len(notStatisfiedRevisions) - len(canApproveRevisions)
		if len(canApproveRevisions) == 0 { // 没有满足条件的更新直接返回结果
			return nil
		}
		// 获取每个可以审批的RevisionID和所有的计算机组的审批关系
		var revisionGroupDeployment []RevisionGroupDeployment
		dataSql := fmt.Sprintf(`select tmp.TargetGroupID, tmp.TargetGroupName, tmp.RevisionID, d.ActionID, d.DeploymentID
								from (select r.RevisionID,
											 c.TargetGroupID,
											 c.TargetGroupName
									  from revision r,
										   computer_target_group c
									  where r.RevisionID in (%s)
										and c.TargetGroupID in (%s)) as tmp
										 left outer join deployment d on tmp.RevisionID = d.RevisionID and d.TargetGroupID = tmp.TargetGroupID
								where d.ActionID not in (%s)
								   or d.ActionID is null;`, GenSqlStrInt(canApproveRevisions), GenSqlStrString(targetGroupList), GenSqlStrInt(filterActions))
		// 每个更新对应每个组的不满足当前ActionID审批关系，满足审批关系的直接不过滤出来，ActionID=None表明和当前组没有审批关系
		if err = db.Raw(dataSql).Scan(&revisionGroupDeployment).Error; err != nil {
			return err
		}
		// 可以审批的RevisionID的Bundle关系
		yesBundleList, err2 := global.GRedis.HMGet("b_r", utils.IntMapStr(canApproveRevisions)...).Result()
		if err2 != nil {
			return err2
		}
		var revisionBundleDict = make(map[int][]int, 0) // revision和Bundle的关系 {2: [3, 4]}
		for i, revisionID := range canApproveRevisions {
			if yesBundleList[i] != nil {
				var revisionHasBundle []int
				_ = json.Unmarshal([]byte(reflect.ValueOf(yesBundleList[i]).String()), &revisionHasBundle)
				canApproveBundleRevisions = append(canApproveBundleRevisions, revisionHasBundle...)
				revisionBundleDict[revisionID] = revisionHasBundle
			}
		}
		// 如果当前RevisionID中有和所有计算机组存在ActionID=8的deployment关系，说明之前是decline状态，需要修改其和所有计算机组的deployment关系(ActionID=2)
		var declineCount int64
		if err = db.Model(&model.Deployment{}).Where("RevisionID in (?) and TargetGroupID = ? and ActionID = 8", canApproveRevisions, utils.UUIDAllComputer).Count(&declineCount).Error; err != nil {
			return err
		}
		if declineCount > 0 {
			if err = db.Model(&model.Deployment{}).Where("RevisionID in (?) and ActionID = 8", canApproveRevisions).Updates(map[string]interface{}{"ActionID": utils.PreDeploymentCheckAction, "LastChangeTime": time.Now().UTC(), "AdminName": adminName}).Error; err != nil {
				return err
			}
			if err = db.Model(&model.Deployment{}).Where("RevisionID in (?) and ActionID = 8", canApproveBundleRevisions).Updates(map[string]interface{}{"ActionID": utils.BundleAction, "LastChangeTime": time.Now().UTC(), "AdminName": adminName}).Error; err != nil {
				return err
			}
		}

		var newDeploymentList []map[string]interface{}
		var deadDeploymentList []map[string]interface{}
		var needHandleRevisions []int
		var needDelDeploymentIDs []int // 需要删除的deployment

		for _, item := range revisionGroupDeployment {
			targetGroupID := item.TargetGroupID
			targetGroupName := item.TargetGroupName
			revisionID := item.RevisionID
			needHandleRevisions = append(needHandleRevisions, revisionID)
			// 新建RevisionID的deployment
			newDeploymentList = append(newDeploymentList, map[string]interface{}{
				"version_id":      1,
				"RevisionID":      revisionID,
				"TargetGroupID":   targetGroupID,
				"TargetGroupName": targetGroupName,
				"ActionID":        actionID,
				"DeploymentGuid":  uuid.New().String(),
				"AdminName":       adminName,
				"LastChangeTime":  time.Now().UTC(),
			})
			// RevisionID旧的Deployment放入dead deployment
			if item.ActionID != nil {
				deadDeploymentList = append(deadDeploymentList, map[string]interface{}{
					"version_id":      1,
					"RevisionID":      revisionID,
					"TargetGroupID":   targetGroupID,
					"TargetGroupName": targetGroupName,
					"ActionID":        item.ActionID,
					"DeploymentGuid":  uuid.New().String(),
					"AdminName":       adminName,
					"LastChangeTime":  time.Now().UTC(),
				})
				needDelDeploymentIDs = append(needDelDeploymentIDs, *item.DeploymentID)
			}
			// 新建Bundle RevisionID的deployment
			nowBundleRevisionIDs := revisionBundleDict[revisionID]
			for _, bundleRevision := range nowBundleRevisionIDs {
				newDeploymentList = append(newDeploymentList, map[string]interface{}{
					"version_id":      1,
					"RevisionID":      bundleRevision,
					"TargetGroupID":   targetGroupID,
					"TargetGroupName": targetGroupName,
					"ActionID":        utils.BundleAction,
					"DeploymentGuid":  uuid.New().String(),
					"AdminName":       adminName,
					"LastChangeTime":  time.Now().UTC(),
				})
				// 旧的bundle RevisionID的deployment
				if item.ActionID != nil {
					deadDeploymentList = append(deadDeploymentList, map[string]interface{}{
						"version_id":      1,
						"RevisionID":      bundleRevision,
						"TargetGroupID":   targetGroupID,
						"TargetGroupName": targetGroupName,
						"ActionID":        utils.BundleAction,
						"DeploymentGuid":  uuid.New().String(),
						"AdminName":       adminName,
						"LastChangeTime":  time.Now().UTC(),
					})
				}
			}
		}
		// 删除旧的deployment
		if len(needDelDeploymentIDs) > 0 {
			delSql1 := fmt.Sprintf("delete from deployment where DeploymentID in (%s);", GenSqlStrInt(needDelDeploymentIDs))
			if err = db.Exec(delSql1).Error; err != nil {
				return err
			}
		}
		// 删除旧的bundle的deployment
		delSql2 := fmt.Sprintf("delete from deployment where TargetGroupID in (%s) and RevisionID in (%s);", GenSqlStrString(targetGroupList), GenSqlStrInt(canApproveBundleRevisions))
		if err = db.Exec(delSql2).Error; err != nil {
			return err
		}
		// 插入deployment和dead deployment数据
		if err = tx.Model(&model.Deployment{}).CreateInBatches(&newDeploymentList, 5000).Error; err != nil {
			return err
		}
		if err = tx.Model(&model.DeadDeployment{}).CreateInBatches(&deadDeploymentList, 5000).Error; err != nil {
			return err
		}

		successCount = len(utils.RemoveRep(needHandleRevisions))

		return nil
	})

	if err != nil {
		global.GLog.Error("AutoApproveWithRule", zap.Any("err", err))
		result = append(result, map[string]interface{}{"name": rule.Name, "status": utils.ApprovalRuleStatus.FAILED, "success": 0, "failed": 0})
	} else {
		if len(canApproveRevisions) > 0 {
			// 刷缓存
			service.GenDeploymentRelationship(append(canApproveRevisions, canApproveBundleRevisions...))
			// 审批到安装下载文件，其他审批取消下载文件
			if *rule.Action == utils.InstallAction {
				download.ProcessDownload(utils.IntMapStr(canApproveRevisions), true, false)
			} else {
				download.ProcessDownload(utils.IntMapStr(canApproveRevisions), false, false)
			}
		}
		result = append(result, map[string]interface{}{"name": rule.Name, "status": utils.ApprovalRuleStatus.SUCCESS, "success": successCount, "failed": failedCount})
	}
	return result
}

// BatchDeclineRevisions 审批规则批量拒绝，拒绝更新，与所有计算机组的关系改成Decline_Action，除下游服务器组外与其他组的关系删除
func BatchDeclineRevisions(rule *model.AutoApproveRules, revisionIDs []int, adminName string, result []map[string]interface{}) []map[string]interface{} {
	db := global.GDb
	var err error
	var needHandleRevisions []int
	var bundleRevisionIds []string
	var revisionStrIDs []string

	err = db.Transaction(func(tx *gorm.DB) error {
		// 获取所有bundle RevisionID
		revisionStrIDs = utils.IntMapStr(revisionIDs)
		bundleRevisionIds, err = service.GenBundleRevisionIds(revisionStrIDs)
		if err != nil {
			return err
		}

		revisionStrIDs = append(revisionStrIDs, bundleRevisionIds...)
		searchStr := GenSqlStrString(revisionStrIDs)
		// 查找需要删除的deployment
		var needDeleteDeployment []DeclineDeployment
		selectSql := fmt.Sprintf("select RevisionID, TargetGroupID, TargetGroupName, ActionID from deployment where RevisionID in (%s) and TargetGroupID != '%s' and ActionID != %d", searchStr, utils.UUIDGroupDss, utils.DeclineAction)
		if err = tx.Raw(selectSql).Scan(&needDeleteDeployment).Error; err != nil {
			return err
		}
		var deadDeploymentList []map[string]interface{}
		for i := 0; i < len(needDeleteDeployment); i++ {
			deploy := needDeleteDeployment[i]
			deadDeploymentList = append(deadDeploymentList, map[string]interface{}{
				"version_id":      1,
				"RevisionID":      deploy.RevisionID,
				"TargetGroupID":   deploy.TargetGroupID,
				"TargetGroupName": deploy.TargetGroupName,
				"ActionID":        deploy.ActionID,
				"DeploymentGuid":  uuid.New().String(),
				"AdminName":       adminName,
				"LastChangeTime":  time.Now().UTC(),
			})
		}
		// 需要拒绝的更新的RevisionID
		revisionSql := fmt.Sprintf("select distinct RevisionID from deployment where RevisionID in (%s) and TargetGroupID != '%s' and ActionID != %d", searchStr, utils.UUIDGroupDss, utils.DeclineAction)
		if err = db.Raw(revisionSql).Scan(&needHandleRevisions).Error; err != nil {
			return err
		}
		// 删除旧的deployments
		delSql := fmt.Sprintf("delete from deployment where RevisionID in (%s) and TargetGroupID != '%s' and ActionID != %d", searchStr, utils.UUIDGroupDss, utils.DeclineAction)
		if err = tx.Exec(delSql).Error; err != nil {
			return err
		}
		// 新建和所有计算机组的拒绝关系
		var newDeploymentList []map[string]interface{}
		for _, item := range needHandleRevisions {
			newDeploymentList = append(newDeploymentList, map[string]interface{}{
				"version_id":      1,
				"RevisionID":      item,
				"TargetGroupID":   utils.UUIDAllComputer,
				"TargetGroupName": utils.NameAllComputer,
				"ActionID":        utils.DeclineAction,
				"DeploymentGuid":  uuid.New().String(),
				"AdminName":       adminName,
				"LastChangeTime":  time.Now().UTC(),
			})
		}
		// 批量插入deployment 和 dead deployment数据
		if err = tx.Model(&model.Deployment{}).CreateInBatches(&newDeploymentList, 5000).Error; err != nil {
			return err
		}
		if err = tx.Model(&model.DeadDeployment{}).CreateInBatches(&deadDeploymentList, 5000).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		global.GLog.Error("BatchDeclineRevisions", zap.Any("err", err))
		if rule != nil {
			result = append(result, map[string]interface{}{"name": rule.Name, "status": utils.ApprovalRuleStatus.FAILED, "success": 0, "failed": 0})
		}
	} else {
		// 刷缓存
		service.GenDeploymentRelationship(utils.StrMapInt(revisionStrIDs))
		// 拒绝更新停止下载其文件
		if len(revisionStrIDs) > 0 {
			download.ProcessDownload(revisionStrIDs, false, false)
		}
		if rule != nil {
			successCount := goset.StrMinus(utils.IntMapStr(needHandleRevisions), bundleRevisionIds).Count()
			result = append(result, map[string]interface{}{"name": rule.Name, "status": utils.ApprovalRuleStatus.SUCCESS, "success": successCount, "failed": 0})
		}
	}
	return result
}

// SyncHistoryUpdates 同步记录详情关联数据
func SyncHistoryUpdates(c *gin.Context, db *gorm.DB, syncHistory *model.SyncHistory) (data []map[string]interface{}, count int) {
	sort := c.DefaultQuery("sort", "RevisionID")
	updateType := c.Query("update_type")
	pageNum, pageSize := GetPageParams(c)

	var filterSql string
	limitSql := fmt.Sprintf(" LIMIT " + strconv.Itoa((pageNum-1)*pageSize) + "," + strconv.Itoa(pageSize)) // 分页

	if updateType == "NewUpdates" { // 新更新
		filterSql = fmt.Sprintf(" AND r.ImportedTime >= '%s' AND r.ImportedTime <= '%s'", syncHistory.StartTime, syncHistory.FinishTime)
		data, count = SimpleRevisionQuery(db, filterSql, sort, limitSql)
	} else if updateType == "RevisedUpdates" { // 修订更新
		// 查找所有UpdateID有重复数据的更新取出RevisionNumber最大的输出结果，根据时间段查询此次同步所包含的RevisionID
		sql := fmt.Sprintf(`select a.RevisionID, a.UpdateID, a.RevisionNumber, a.UpdateType
                      from revision a, (select UpdateID, max(RevisionNumber) RevisionNumber, count(1) count
                      from revision r where r.ImportedTime <= '%s' group by UpdateID) b where  b.count > 1 and a.UpdateID = b.UpdateID
                      and a.RevisionNumber = b.RevisionNumber and a.UpdateType = "Software"
                      and a.ImportedTime >= '%s' and a.ImportedTime <= '%s'`, syncHistory.FinishTime, syncHistory.StartTime, syncHistory.FinishTime)

		var revisionQuery []map[string]interface{}
		if err := db.Raw(sql).Scan(&revisionQuery).Error; err != nil {
			global.GLog.Error("GetSyncHistoryUpdates", zap.Any("err", err))
		}

		if len(revisionQuery) > 0 {
			var updateIDs []interface{}
			var revisionIDs []interface{}
			for i := 0; i < len(revisionQuery); i++ {
				updateIDs = append(updateIDs, revisionQuery[i]["UpdateID"])
				revisionIDs = append(revisionIDs, revisionQuery[i]["RevisionID"])
			}
			filterSql = fmt.Sprintf(`AND r.RevisionID NOT IN (%s) AND r.UpdateID IN (%s) AND r.ImportedTime <= '%s'`, GenSqlStr(revisionIDs), GenSqlStr(updateIDs), syncHistory.FinishTime)
		}
		data, count = SimpleRevisionQuery(db, filterSql, sort, limitSql)
	} else if updateType == "ExpiredUpdates" { // 替代更新 superseded
		var updateIDs []string
		sql := fmt.Sprintf("select distinct r.UpdateID from revision r join superseded s on r.RevisionID = s.RevisionID where r.ImportedTime >= '%s' and r.ImportedTime <= '%s';", syncHistory.StartTime, syncHistory.FinishTime)
		if err := db.Raw(sql).Scan(&updateIDs).Error; err != nil {
			global.GLog.Error("GetSyncHistoryUpdates", zap.Any("err", err))
		}
		if len(updateIDs) > 0 {
			filterSql = fmt.Sprintf(" AND r.UpdateID IN (%s) AND r.ImportedTime <= '%s' ", GenSqlStrString(updateIDs), syncHistory.FinishTime)
		}
		data, count = SimpleRevisionQuery(db, filterSql, sort, limitSql)
	} else if updateType == "MSExpiredUpdates" { //过期更新
		var updateIDs []string
		db.Model(&model.Revision{}).Where("UpdateType NOT IN ('Category', 'Detectoid') AND ProductRevisionID IS NOT NULL AND ClassificationRevisionID IS NOT NULL").
			Where("PublicationState = ?", utils.PublicationState.Expired).
			Where("ImportedTime >= ?", syncHistory.StartTime).
			Where("ImportedTime <= ?", syncHistory.FinishTime).Distinct("UpdateID").Select("UpdateID").Pluck("UpdateID", &updateIDs)

		if len(updateIDs) > 0 {
			filterSql = fmt.Sprintf(" AND r.UpdateID IN (%s) ", GenSqlStrString(updateIDs))
		}
		data, count = SimpleRevisionQuery(db, filterSql, sort, limitSql)
		// 这里加判断的原因是过期更新要修改原有revision的PublicationState，会导致同步记录显示没有过期更新，但是关联数据页出现过期更新
		// 暂时没有更好的办法解决
		if syncHistory.MSExpiredUpdates != count {
			data, count = []map[string]interface{}{}, 0
		}
	}
	return
}

// SimpleRevisionQuery 更新过滤
func SimpleRevisionQuery(db *gorm.DB, filterSql string, sort string, limitSql string) (data []map[string]interface{}, count int) {
	if filterSql == "" {
		return
	}

	countSql := fmt.Sprintf(`SELECT COUNT(1) AS count FROM revision r WHERE r.UpdateType NOT IN ('Category', 'Detectoid')
        AND r.ProductRevisionID IS NOT NULL AND r.ClassificationRevisionID IS NOT NULL %s;`, filterSql)
	if err := db.Raw(countSql).Scan(&count).Error; err != nil {
		global.GLog.Error("SimpleRevisionQuery", zap.Any("err", err))
	}

	baseSql := fmt.Sprintf(`SELECT a.RevisionID, a.ProductRevisionID, a.ClassificationRevisionID,
        (SELECT Title FROM property WHERE RevisionID = a.RevisionID AND Language IN ('zh-cn', 'en') ORDER BY Language DESC LIMIT 1) AS Title,
        (SELECT Title FROM property WHERE RevisionID = a.ProductRevisionID AND Language IN ('zh-cn', 'en') ORDER BY Language DESC LIMIT 1) AS ProductTitle,
        (SELECT Title FROM property WHERE RevisionID = a.ClassificationRevisionID AND Language IN ('zh-cn', 'en') ORDER BY Language DESC LIMIT 1) AS ClassificationTitle
        FROM (SELECT r.RevisionID, r.ProductRevisionID, r.ClassificationRevisionID FROM revision r
        WHERE r.UpdateType NOT IN ('Category', 'Detectoid') AND r.ProductRevisionID IS NOT NULL AND r.ClassificationRevisionID IS NOT NULL
        %s) a ORDER BY %s %s;`, filterSql, sort, limitSql)

	if err := db.Raw(baseSql).Scan(&data).Error; err != nil {
		global.GLog.Error("SimpleRevisionQuery", zap.Any("err", err))
	}
	return
}

// EulaAccepttance 接受Eula许可协议
func EulaAccepttance(userID int, userEmail string, revisionID string) (error, int) {
	db := global.GDb

	serverConf := service.GetGlobalServerConfig()
	serverName := serverConf.FullDomainName
	adminName := serverName + "/" + userEmail

	var revision struct {
		RevisionID int
		EulaID     string
	}
	err := db.Model(&model.Revision{}).Select("RevisionID", "EulaID").Where("RevisionID = ?", revisionID).First(&revision).Error
	if err == gorm.ErrRecordNotFound || revision.EulaID == "" {
		return gorm.ErrRecordNotFound, response.NotFound
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var eulaAcceptance model.EulaAcceptance
		if err = tx.Where("eula_id = ?", revision.EulaID).First(&eulaAcceptance).Error; err == gorm.ErrRecordNotFound {
			newAcceptance := model.EulaAcceptance{EulaID: revision.EulaID, AdminName: adminName, AcceptedDate: time.Now().UTC()}
			if err = tx.Create(&newAcceptance).Error; err != nil {
				return err
			}
			if err = tx.Model(&model.Revision{}).Where("EulaID = ? and RequiresReacceptanceOfEula is true", revision.EulaID).Update("EulaExplicitlyAccepted", true).Error; err != nil {
				return err
			}
			// 创建操作记录
			propertyTitle := service.GetPropertyTitle(revisionID)
			desc := fmt.Sprintf("接受【%s】许可协议", propertyTitle)
			GenOperateRecord(OperateAction.AcceptEula, desc, ResultStatus.SUCCESS, userID, userEmail)
		}
		return nil
	})
	if err != nil {
		return err, response.FAILED
	} else {
		return nil, response.SUCCESS
	}
}

func GetOperateMapping() ([]map[string]interface{}, []map[string]interface{}) {
	var actionList []map[string]interface{}
	var resultList []map[string]interface{}

	for key, val := range OperateActionDesc {
		actionList = append(actionList, map[string]interface{}{
			"key":   key,
			"label": val,
		})
	}

	for key, val := range ResultStatusMap {
		resultList = append(resultList, map[string]interface{}{
			"key":   key,
			"label": val,
		})
	}
	return actionList, resultList
}

// GetOperateRecordList 获取操作记录列表
func GetOperateRecordList(c *gin.Context, export bool) ([]model.OperateRecord, int64) {
	db := global.GDb

	pageNum, pageSize := GetPageParams(c)
	operatorID := c.Query("operator_id")
	operate := c.Query("operate")
	operateResult := c.Query("operate_result")
	startTimestamp := c.Query("start_timestamp")
	endTimestamp := c.Query("end_timestamp")

	var data []model.OperateRecord

	db = db.Model(&model.OperateRecord{})

	if operatorID != "" {
		var operatorIDs []string
		_ = json.Unmarshal([]byte(operatorID), &operatorIDs)
		db = db.Where("operator_id in (?)", operatorIDs)
	}
	if operate != "" {
		var operates []int
		_ = json.Unmarshal([]byte(operate), &operates)
		db = db.Where("operate in (?)", operates)
	}
	if operateResult != "" {
		db = db.Where("operate_result = ?", operateResult)
	}
	if startTimestamp != "" {
		startTime, _ := strconv.Atoi(startTimestamp)
		db = db.Where("operate_time >= ?", time.Unix(int64(startTime), 0).Format("2006-01-02 15:04:05"))
	}
	if endTimestamp != "" {
		endTime, _ := strconv.Atoi(endTimestamp)
		db = db.Where("operate_time <= ?", time.Unix(int64(endTime), 0).Format("2006-01-02 15:04:05"))
	}

	db = db.Order("id desc")
	// 不是导出，需要分页和查询总数
	var count int64
	if !export {
		if err := db.Count(&count).Error; err != nil {
			global.GLog.Error("GetOperateRecordList", zap.Any("err", err))
			return data, 0
		}
		db = db.Offset((pageNum - 1) * pageSize).Limit(pageSize)
	}
	if err := db.Find(&data).Error; err != nil {
		global.GLog.Error("GetOperateRecordList", zap.Any("err", err))
		return data, 0
	}

	return data, count
}

// NeedCleanRevisionIDs
func NeedCleanRevisionIDs(db *gorm.DB, revisionIDs []int) []int {
	bundleRevisionIDs, _ := service.GenBundleRevisionIds(utils.IntMapStr(revisionIDs))
	var results []struct {
		BundleRevisionID, Count int
	}
	db.Table("bundle").Select("BundleRevisionID, count(BundleRevisionID) as Count").Where("BundleRevisionID in (?)", bundleRevisionIDs).Group("BundleRevisionID").Having("Count > 1").Scan(&results)
	var repeatBundleRevisionIDs []int
	for _, item := range results {
		repeatBundleRevisionIDs = append(repeatBundleRevisionIDs, item.BundleRevisionID)
	}

	var bundleRevisionIDs2 []int
	sql := fmt.Sprintf("select BundleRevisionID from bundle where RevisionID not in (%s) and BundleRevisionID in (%s)", GenSqlStrInt(revisionIDs), GenSqlStrInt(repeatBundleRevisionIDs))
	db.Raw(sql).Scan(&bundleRevisionIDs2)

	return goset.IntMinus(utils.StrMapInt(bundleRevisionIDs), bundleRevisionIDs2).List()
}

// CleanUnusedUpdates 未使用的更新和更新修订 删除过期更新和已三个月或更长时间没有审批的更新，以及删除已30天或更长时间没有审批的早期修订更新
func CleanUnusedUpdates() error {
	db := global.GDb
	re := global.GRedis

	err := db.Transaction(func(tx *gorm.DB) error {
		// 未使用的更新和更新修订 删除过期更新和已三个月或更长时间没有审批的更新，以及删除已30天或更长时间没有审批的早期修订更新
		revisionQuery := tx.Model(&model.Revision{}).Select("RevisionID").Where("UpdateType = Software")
		// 早期修订更新更新
		var expireRevisionIDs []int
		if err := revisionQuery.Where("IsLatestRevision is false and ImportedTime <= ?", time.Now().UTC().AddDate(0, 0, -30)).Scan(&expireRevisionIDs).Error; err != nil {
			return err
		}
		// 被取代更新
		var supersedingRevisionIDs []int
		revisionQuery.Joins("left join superseded on revision.UpdateID = superseded.UpdateID").Where("revision.ImportedTime <= ?", time.Now().UTC().AddDate(0, 0, -30)).Scan(&supersedingRevisionIDs)
		revisionIDs := append(expireRevisionIDs, supersedingRevisionIDs...)
		var deploymentRevisionIDs []int
		tx.Model(&model.Deployment{}).Select("RevisionID").Where("RevisionID in (?) and ActionID in (?)", revisionIDs, []int{utils.InstallAction, utils.UninstallAction, utils.DeclineAction, utils.BlockAction}).Distinct().Scan(&deploymentRevisionIDs)
		needDeleteRevisionIDs := goset.IntMinus(revisionIDs, deploymentRevisionIDs).List()
		needDeleteRevisionIDs = append(needDeleteRevisionIDs, NeedCleanRevisionIDs(db, needDeleteRevisionIDs)...)
		// update与revision关系
		var prerequisiteIDs []int
		tx.Table("update_for_prerequisite").Select("PrerequisiteID").Joins("left join revision_prerequisite on revision_prerequisite.PrerequisiteID = update_for_prerequisite.PrerequisiteID").Where("revision_prerequisite.RevisionID in (?)", needDeleteRevisionIDs).Scan(&prerequisiteIDs)
		// 开始删除
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Or("BundleRevisionID in (?)", needDeleteRevisionIDs).Delete(&model.Bundle{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.DeadDeployment{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.Deployment{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.DownloadFiles{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.Property{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.RevisionInCategory{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.Rules{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.StatisticsForPerUpdate{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.Superseded{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.UpdateStatus{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.UpdateStatusPerComputer{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.RevisionStatement{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.RevisionInstallStatistics{})
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.ComputerRevisionInstallStats{})
		tx.Where("PrerequisiteID in (?)", prerequisiteIDs).Delete(&model.UpdateForPrerequisite{})
		tx.Where("PrerequisiteID in (?)", prerequisiteIDs).Delete(&model.RevisionPrerequisite{})
		var localUpdateIDs []int
		tx.Model(&model.Revision{}).Where("RevisionID in (?) and IsLatestRevision is true", needDeleteRevisionIDs).Pluck("LocalUpdateID", &localUpdateIDs)
		tx.Where("RevisionID in (?)", needDeleteRevisionIDs).Delete(&model.Revision{})
		tx.Where("LocalUpdateID in (?)", localUpdateIDs).Delete(&model.Update{})
		service.FlushRebuildCache()
		if len(needDeleteRevisionIDs) > 0 {
			// 将计算机更新状态报表中与被删除更新相关的数据作废
			tx.Model(&model.ComputerUpdateRelation{}).Where("revision_id in (?)", needDeleteRevisionIDs).Update("is_valid", false)
			needDeleteRevisionIDsStr := utils.IntMapStr(needDeleteRevisionIDs)
			re.HDel("r_i", needDeleteRevisionIDsStr...)
			re.HDel("d_i", needDeleteRevisionIDsStr...)
			re.HDel("d_f", needDeleteRevisionIDsStr...)
			re.HDel("r_x", needDeleteRevisionIDsStr...)
		}
		return nil
	})
	return err
}

// CleanNoConnectPC 删除已30天或更多天没有连接到服务器的计算机
func CleanNoConnectPC(c *gin.Context) error {
	db := global.GDb
	re := global.GRedis

	err := db.Transaction(func(tx *gorm.DB) error {
		deadLine := time.Now().UTC().AddDate(0, 0, -30)
		var computerTargets []struct {
			TargetID   int
			ComputerID string
		}
		sql := fmt.Sprintf("select TargetID, ComputerID from computer_target where LastReportedStatusTime <= '%s' or (LastReportedStatusTime is null and LastSyncTime <= '%s');", deadLine, deadLine)
		if err := tx.Raw(sql).Scan(&computerTargets).Error; err != nil {
			return err
		}
		var targetIDs []int
		var computerIDs []string
		for _, item := range computerTargets {
			targetIDs = append(targetIDs, item.TargetID)
			computerIDs = append(computerIDs, "c_i"+item.ComputerID)
		}
		tx.Where("TargetID in (?)", targetIDs).Delete(&model.ComputerInGroup{})
		tx.Where("TargetID in (?)", targetIDs).Delete(&model.UpdateStatusPerComputer{})
		tx.Where("TargetID in (?)", targetIDs).Delete(&model.ComputerSummaryForMicrosoftUpdates{})
		tx.Where("TargetID in (?)", targetIDs).Delete(&model.ComputerStatement{})
		tx.Where("TargetID in (?)", targetIDs).Delete(&model.ComputerRevisionInstallStats{})
		delSql := fmt.Sprintf("delete from computer_target where LastReportedStatusTime <= '%s' or (LastReportedStatusTime is null and LastSyncTime <= '%s');", deadLine, deadLine)
		if err := tx.Raw(delSql).Error; err != nil {
			return err
		}
		// 清除缓存
		keys := re.Keys("m_h:*").Val()
		if len(keys) > 0 {
			re.Del(keys...)
		}
		if len(computerIDs) > 0 {
			re.Del(computerIDs...)
		}
		return nil
	})
	// 操作日志
	userID, userEmail := GetUserInfo(c)
	if err != nil {
		GenOperateRecord(OperateAction.CleanComputer, "清理计算机失败", ResultStatus.FAILED, userID, userEmail)
	} else {
		GenOperateRecord(OperateAction.CleanComputer, "执行计算机清理规则", ResultStatus.SUCCESS, userID, userEmail)
	}

	return err
}

// CleanNoNeedUpdates 不需要的更新文件 删除更新或下游服务器不需要的更新文件
func CleanNoNeedUpdates() error {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		var files []struct {
			FileDigest, FileName string
		}
		if err := tx.Model(&model.DownloadFiles{}).Select("FileDigest", "FileName").Scan(&files).Error; err != nil {
			return err
		}
		var filesPathList []string
		for _, item := range files {
			filesPathList = append(filesPathList, GetFileUrl(item.FileDigest, item.FileName))
		}
		var localFileList []string
		GetLocalFileList(global.GConfig.CUS.DirPath, &localFileList)

		deletePathList := goset.StrMinus(localFileList, filesPathList).List()
		for _, path := range deletePathList {
			_ = os.Remove(filepath.Join(global.GConfig.CUS.DirPath, path))
		}
		return nil
	})
	return err
}

// CleanExpireUpdates 过期的更新 拒绝没有审批的更新以及Microsoft终止的更新
func CleanExpireUpdates(adminName string) error {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		var expireRevisionIDs []int
		if err := tx.Model(&model.Revision{}).Where("IsLatestRevision is false and UpdateType == 'Software'").Pluck("RevisionID", &expireRevisionIDs).Error; err != nil {
			return err
		}

		var deploymentRevisionIDs []int
		if err := tx.Model(&model.Deployment{}).Where("RevisionID in (?) and ActionID in (?)", expireRevisionIDs, []int{utils.InstallAction, utils.UninstallAction, utils.DeclineAction, utils.BlockAction}).Pluck("RevisionID", &deploymentRevisionIDs).Distinct("RevisionID").Error; err != nil {
			return err
		}

		notDeploymentRevisionIDs := goset.IntMinus(expireRevisionIDs, deploymentRevisionIDs).List()
		notDeploymentRevisionIDs = append(notDeploymentRevisionIDs, NeedCleanRevisionIDs(db, notDeploymentRevisionIDs)...)
		BatchDeclineRevisions(nil, notDeploymentRevisionIDs, adminName, []map[string]interface{}{})
		return nil
	})
	return err
}

// CleanSupersededUpdates 拒绝已30天或更长时间没有审批的更新、当前客户端都不需要的更新以及被已审批更新取代的更新
func CleanSupersededUpdates(adminName string) error {
	db := global.GDb

	err := db.Transaction(func(tx *gorm.DB) error {
		var supersedingRevisionIDs []int
		sql := fmt.Sprintf("select r.RevisionID from revision r left join superseded s on r.UpdateID = s.UpdateID where r.ImportedTime <= '%s' and r.UpdateType = 'Software';", time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02 15:04:05"))
		if err := tx.Raw(sql).Scan(&supersedingRevisionIDs).Error; err != nil {
			return err
		}

		var deploymentRevisionIDs []int
		if err := tx.Model(&model.Deployment{}).Where("RevisionID in (?) and ActionID in (?)", supersedingRevisionIDs, []int{utils.InstallAction, utils.UninstallAction, utils.DeclineAction, utils.BlockAction}).Pluck("RevisionID", &deploymentRevisionIDs).Distinct("RevisionID").Error; err != nil {
			return err
		}

		notDeploymentRevisionIDs := goset.IntMinus(supersedingRevisionIDs, deploymentRevisionIDs).List()
		notDeploymentRevisionIDs = append(notDeploymentRevisionIDs, NeedCleanRevisionIDs(db, notDeploymentRevisionIDs)...)
		BatchDeclineRevisions(nil, notDeploymentRevisionIDs, adminName, []map[string]interface{}{})
		return nil
	})
	return err
}
