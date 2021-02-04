package admin

import (
	"time"
)

// 计算机更新安装状态title mapping
var ComputerRevisionInstallStatsChoices = []map[string]interface{}{
	{"key": "TargetGroupID", "val": "计算机组", "filter_key": "TargetGroupID", "filter_type": "is", "required": true, "show": true, "filter_order": 0},
	{"key": "ActionID", "val": "更新审批状态", "filter_key": "ActionID", "filter_type": "is", "required": true, "show": true, "filter_order": 10, "filter": ApproveStatusFilterMap},
	{"key": "RevisionID", "val": "RevisionID", "required": false, "show": false, "filter_order": 30},
	{"key": "Title", "val": "更新名称", "required": true, "show": true, "filter_order": 40},
	{"key": "KBArticleID", "val": "KB 号", "required": false, "show": true, "filter_order": 50},
	{"key": "MsrcSeverity", "val": "MSRC 严重程度", "required": false, "show": true, "filter_key": "MsrcSeverity", "attr": "更新", "filter_order": 60, "filter": MSRCSeverityEnMap},
	{"key": "SecurityBulletinID", "val": "MSRC 编号", "required": false, "show": true, "filter_order": 70},
	{"key": "ProductTitle", "val": "产品", "required": false, "show": true, "filter_key": "ProductRevisionID", "filter_order": 80},
	{"key": "ClassificationTitle", "val": "分类", "required": false, "show": true, "filter_key": "ClassificationRevisionID", "filter_order": 90},
	{"key": "CreationDate", "val": "发布日期", "required": false, "show": true, "filter_key": "CreationDate", "filter_type": "range", "filter_order": 100},
	{"key": "ImportedTime", "val": "到达日期", "required": false, "show": true, "filter_order": 110},
	{"key": "LastChangedAnchor", "val": "修订日期", "required": false, "show": true, "filter_order": 120},
	{"key": "UpdateID", "val": "更新 ID", "required": true, "show": true, "filter_order": 130},
	{"key": "RevisionNumber", "val": "修订号", "required": true, "show": true, "filter_order": 140},
	// 计算机
	{"key": "TargetID", "val": "TargetID", "required": false, "show": false, "filter_order": 150},
	{"key": "FullDomainName", "val": "计算机名称", "required": true, "show": true, "filter_order": 160},
	{"key": "IPAddress", "val": "IP 地址", "required": false, "show": true, "filter_order": 170},
	{"key": "OSVersion", "val": "操作系统版本", "required": false, "show": true, "filter_order": 180},
	{"key": "ComputerMake", "val": "制造商", "required": false, "show": true, "filter_order": 190},
	{"key": "ComputerModel", "val": "型号", "required": false, "show": true, "filter_order": 200},
	{"key": "FirmwareVersion", "val": "固件", "required": false, "show": true, "filter_order": 210},
	{"key": "BiosName", "val": "BIOS 名称", "required": false, "show": true, "filter_order": 220},
	{"key": "BiosVersion", "val": "BIOS 版本", "required": false, "show": true, "filter_order": 230},
	{"key": "BiosReleaseDate", "val": "BIOS 发布时间", "required": false, "show": true, "filter_order": 240},
	{"key": "OSDescription", "val": "操作系统架构", "required": false, "show": true, "filter_order": 250},
	{"key": "OSMajorVersion", "val": "操作系统主版本号", "required": false, "show": true, "filter_order": 260},
	{"key": "OSMinorVersion", "val": "操作系统次版本号", "required": false, "show": true, "filter_order": 270},
	{"key": "OSBuildNumber", "val": "操作系统 Build 号", "required": false, "show": true, "filter_order": 280},
	{"key": "MobileOperator", "val": "移动运营商", "required": false, "show": true, "filter_order": 290},
	{"key": "ClientVersion", "val": "客户端版本", "required": false, "show": true, "filter_order": 300},
	{"key": "Status", "val": "更新安装状态", "required": true, "show": true, "filter_key": "Status", "filter_order": 20, "filter": ComputerUpdateStatusMap},
}

// 审批状态报表过滤
var ApproveStatusFilterMap = map[string]string{
	"0": "安装",
	"2": "未审批",
	"1": "卸载",
	"3": "未审批",
	"8": "拒绝",
}

// MSRC Severity严重程度
var MSRCSeverityEnMap = map[string]string{
	"Unspecified": "未指定",
	"Low":         "低",
	"Moderate":    "中",
	"Important":   "重要",
	"Critical":    "关键",
}

// 计算机的更新的状态
var ComputerUpdateStatusMap = map[int]string{
	0: "状态未知或不适用",
	2: "需要安装",
	3: "已下载但未安装",
	4: "已安装",
	5: "安装失败",
	6: "安装但需要重启",
}

// 更新报表title mapping
var UpdateTitleChoices = []map[string]interface{}{
	{"key": "RevisionID", "val": "RevisionID", "required": false, "show": false, "filter_order": 0},
	{"key": "Title", "val": "更新名称", "required": true, "show": true, "filter_order": 10},
	{"key": "KBArticleID", "val": "KB 号", "required": true, "show": true, "filter_order": 20},
	{"key": "MsrcSeverity", "val": "MSRC 严重程度", "required": false, "show": true, "filter_key": "MsrcSeverity", "filter": MsrcSeverityCnMap, "filter_order": 30},
	{"key": "SecurityBulletinID", "val": "MSRC 编号", "required": false, "show": true, "filter_order": 40},
	{"key": "ProductTitle", "val": "产品", "required": false, "show": true, "filter_key": "ProductRevisionID", "filter_order": 50},
	{"key": "ClassificationTitle", "val": "分类", "required": false, "show": true, "filter_key": "ClassificationRevisionID", "filter_order": 60},
	{"key": "TargetGroupName", "val": "计算机组", "required": false, "show": true, "filter_key": "TargetGroupID", "filter_order": 70},
	{"key": "ApproveStatus", "val": "审批状态", "required": false, "show": true, "filter_key": "ActionID", "filter": ApproveStatusFilterMap, "filter_order": 80},
	{"key": "AdminName", "val": "审批人", "required": false, "show": true, "filter_order": 90},
	{"key": "DeploymentTime", "val": "审批时间", "required": false, "show": true, "filter_order": 100},
	{"key": "CreationDate", "val": "发布日期", "required": false, "show": true, "filter_key": "CreationDate", "filter_type": "range", "filter_order": 110},
	{"key": "ImportedTime", "val": "到达日期", "required": false, "show": true, "filter_order": 120},
	{"key": "UpdateID", "val": "更新 ID", "required": true, "show": true, "filter_order": 130},
	{"key": "RevisionNumber", "val": "修订号", "required": true, "show": true, "filter_order": 140},
	{"key": "LastChangedAnchor", "val": "修订日期", "required": false, "show": true, "filter_order": 150},
}

// 更新严重程度报表过滤
var MsrcSeverityCnMap = map[string]string{
	"未指定": "未指定",
	"低":   "低",
	"中":   "中",
	"重要":  "重要",
	"关键":  "关键",
}

// 计算机报表title mapping
var ComputerTitleChoices = []map[string]interface{}{
	{"key": "TargetID", "val": "TargetID", "required": false, "show": false, "filter_order": 0},
	{"key": "FullDomainName", "val": "计算机名称", "required": true, "show": true, "filter_order": 10},
	{"key": "IPAddress", "val": "IP 地址", "required": false, "show": true, "filter_order": 20},
	{"key": "OSVersion", "val": "操作系统版本", "required": false, "show": true, "filter_order": 30},
	{"key": "ClientVersion", "val": "客户端版本", "required": false, "show": true, "filter_order": 40},
	{"key": "LastSyncTime", "val": "上次同步时间", "required": true, "show": true, "filter_order": 50},
	{"key": "LastReportedStatusTime", "val": "上次报告状态时间", "required": true, "show": true, "filter_order": 60},
	{"key": "NotInstalled", "val": "需要安装的更新总数", "required": false, "show": true, "filter_order": 70},
	{"key": "Downloaded", "val": "下载但未安装的更新总数", "required": false, "show": true, "filter_order": 80},
	{"key": "InstalledPendingReboot", "val": "安装但需要重启计算机的更新总数", "required": false, "show": true, "filter_order": 90},
	{"key": "Failed", "val": "安装失败的更新总数", "required": false, "show": true, "filter_order": 100},
	{"key": "Installed", "val": "已安装的更新总数", "required": false, "show": true, "filter_order": 110},
	{"key": "Unknown", "val": "状态未知和不适用的更新总数", "required": false, "show": true, "filter_order": 120},
	{"key": "ComputerMake", "val": "制造商", "required": false, "show": true, "filter_order": 130},
	{"key": "ComputerModel", "val": "型号", "required": false, "show": true, "filter_order": 140},
	{"key": "FirmwareVersion", "val": "固件", "required": false, "show": true, "filter_order": 150},
	{"key": "BiosName", "val": "BIOS 名称", "required": false, "show": true, "filter_order": 160},
	{"key": "BiosVersion", "val": "BIOS 版本", "required": false, "show": true, "filter_order": 170},
	{"key": "BiosReleaseDate", "val": "BIOS 发布时间", "required": false, "show": true, "filter_order": 180},
	{"key": "OSDescription", "val": "操作系统架构", "required": false, "show": true, "filter_order": 190},
	{"key": "OSMajorVersion", "val": "操作系统主版本号", "required": false, "show": true, "filter_order": 200},
	{"key": "OSMinorVersion", "val": "操作系统次版本号", "required": false, "show": true, "filter_order": 210},
	{"key": "OSBuildNumber", "val": "操作系统 Build 号", "required": false, "show": true, "filter_order": 220},
	{"key": "MobileOperator", "val": "移动运营商", "required": false, "show": true, "filter_order": 230},
}

// 更新安装统计title map
var UpdateInstallStatsTitleChoices = []map[string]interface{}{
	{"key": "RevisionID", "val": "RevisionID", "required": false, "show": false, "filter_order": 0},
	{"key": "Title", "val": "更新名称", "required": true, "show": true, "filter_order": 10},
	{"key": "ComputerCount", "val": "计算机总数", "required": false, "show": true, "filter_order": 20},
	{"key": "Downloaded", "val": "下载但未安装的计算机总数", "required": false, "show": true, "filter_order": 30},
	{"key": "Failed", "val": "安装失败的计算机总数", "required": false, "show": true, "filter_order": 40},
	{"key": "InstalledPendingReboot", "val": "安装但需要重启的计算机总数", "required": false, "show": true, "filter_order": 50},
	{"key": "Installed", "val": "已安装的计算机总数", "required": false, "show": true, "filter_order": 60},
	{"key": "NotInstalled", "val": "需要安装的计算机总数", "required": false, "show": true, "filter_order": 70},
	{"key": "Unknown", "val": "状态未知和不适用的计算机总数", "required": false, "show": true, "filter_order": 80},
}

// 下游服务器title mapping
var DssTitleChoices = []map[string]interface{}{
	{"key": "FullDomainName", "val": "名称", "required": true, "show": true, "filter_order": 0},
	{"key": "ServerId", "val": "ID", "required": true, "show": true, "filter_order": 10},
	{"key": "IsReplica", "val": "同步模式", "required": false, "show": true, "filter_key": "IsReplica", "filter": SyncModeFilter, "filter_order": 20},
	{"key": "LastSyncTime", "val": "上次同步时间", "required": true, "show": true, "filter_order": 30},
	{"key": "LastRollupTime", "val": "上次汇总时间", "required": false, "show": true, "filter_order": 40},
	{"key": "Version", "val": "版本", "required": false, "show": true, "filter_order": 50},
	{"key": "UpdateCount", "val": "更新总数", "required": false, "show": true, "filter_order": 60},
	{"key": "ApprovedUpdateCount", "val": "已审批的更新总数", "required": false, "show": true, "filter_order": 70},
	{"key": "NotApprovedUpdateCount", "val": "未审批的更新总数", "required": false, "show": true, "filter_order": 80},
	{"key": "CriticalOrSecurityUpdatesNotApprovedForInstallCount", "val": "未审批的关键更新总数", "required": false, "show": true, "filter_order": 90},
	{"key": "ExpiredUpdateCount", "val": "已过期的更新总数", "required": false, "show": true, "filter_order": 100},
	{"key": "DeclinedUpdateCount", "val": "已拒绝的更新总数", "required": false, "show": true, "filter_order": 110},
	{"key": "UpdatesUpToDateCount", "val": "需要安装的更新总数", "required": false, "show": true, "filter_order": 120},
	{"key": "UpdatesNeedingFilesCount", "val": "需要下载文件的更新总数", "required": false, "show": true, "filter_order": 130},
	{"key": "CustomComputerTargetGroupCount", "val": "自定义组总数", "required": false, "show": true, "filter_order": 140},
	{"key": "ComputerTargetCount", "val": "计算机总数", "required": false, "show": true, "filter_order": 150},
	{"key": "ComputerTargetsNeedingUpdatesCount", "val": "需要安装更新的计算机总数", "required": false, "show": true, "filter_order": 160},
}

// 同步模式报表过滤
var SyncModeFilter = map[string]string{
	"自治": "自治",
	"副本": "副本",
}

// 报表服务类型
var ServiceType = struct {
	Updates, Computer, UpdateInstallStats, DSS, ComputerUpdatesStatus int
}{
	Updates:               1, // 更新
	Computer:              2, // 计算机
	UpdateInstallStats:    3, // 更新安装统计
	DSS:                   4, // 下游服务器
	ComputerUpdatesStatus: 5, // 计算机更新安装状态
}

var ServiceTypeMap = map[int]string{
	ServiceType.Updates:               "更新",
	ServiceType.Computer:              "计算机",
	ServiceType.UpdateInstallStats:    "更新安装统计",
	ServiceType.DSS:                   "下游服务器",
	ServiceType.ComputerUpdatesStatus: "计算机更新安装状态",
}

// 操作记录相关action
var OperateAction = struct {
	LOGIN               int
	LOGOUT              int
	ChangePwd           int
	ImportUpdate        int
	CleanUpdate         int
	CreatApproval       int
	RuleApproval        int
	DeployUpdate        int
	BeginDownload       int
	EndDownload         int
	INSTALL             int
	UNINSTALL           int
	PreDeploymentCheck  int
	DECLINE             int
	CleanComputer       int
	DeleteComputer      int
	DealComputer        int
	CancelDealComputer  int
	DeleteComputerGroup int
	AddComputerGroup    int
	EditComputerGroup   int
	DeleteDssServer     int
	CreateReport        int
	ExportReport        int
	DeleteReport        int
	BeginSync           int
	CancelSync          int
	SetSyncSource       int
	SetSyncPlan         int
	SetUpdateLanguage   int
	SetUpdateProduct    int
	SetUpdateClassfiy   int
	AcceptEula          int
	DownloadApproval    int
	DownloadExpress     int
	ExportOperate       int
	ExportImage         int
	AddDownloadList     int
	RemoveDownloadList  int
}{
	LOGIN:               2,
	LOGOUT:              4,
	ChangePwd:           6,
	ImportUpdate:        8,
	CleanUpdate:         10,
	CreatApproval:       12,
	RuleApproval:        14,
	DeployUpdate:        16,
	BeginDownload:       18,
	EndDownload:         20,
	INSTALL:             22,
	UNINSTALL:           24,
	PreDeploymentCheck:  26,
	DECLINE:             28,
	CleanComputer:       30,
	DeleteComputer:      32,
	DealComputer:        34,
	CancelDealComputer:  36,
	DeleteComputerGroup: 38,
	AddComputerGroup:    40,
	EditComputerGroup:   42,
	DeleteDssServer:     44,
	CreateReport:        46,
	ExportReport:        48,
	DeleteReport:        50,
	BeginSync:           52,
	CancelSync:          54,
	SetSyncSource:       56,
	SetSyncPlan:         58,
	SetUpdateLanguage:   60,
	SetUpdateProduct:    62,
	SetUpdateClassfiy:   64,
	AcceptEula:          66,
	DownloadApproval:    68,
	DownloadExpress:     70,
	ExportOperate:       72,
	ExportImage:         74,
	AddDownloadList:     76,
	RemoveDownloadList:  78,
}

var OperateActionDesc = map[int]string{
	OperateAction.LOGIN:               "登录系统",
	OperateAction.LOGOUT:              "注销登录",
	OperateAction.ChangePwd:           "修改密码",
	OperateAction.ImportUpdate:        "导入更新",
	OperateAction.CleanUpdate:         "清理更新",
	OperateAction.CreatApproval:       "创建审批规则",
	OperateAction.RuleApproval:        "按规则审批",
	OperateAction.DeployUpdate:        "配置更新修订",
	OperateAction.BeginDownload:       "开始下载",
	OperateAction.EndDownload:         "取消下载",
	OperateAction.INSTALL:             "审批到安装",
	OperateAction.UNINSTALL:           "审批到卸载",
	OperateAction.PreDeploymentCheck:  "取消审批",
	OperateAction.DECLINE:             "拒绝更新",
	OperateAction.CleanComputer:       "清理计算机",
	OperateAction.DeleteComputer:      "删除计算机",
	OperateAction.DealComputer:        "分配计算机",
	OperateAction.CancelDealComputer:  "取消计算机分配",
	OperateAction.DeleteComputerGroup: "删除计算机组",
	OperateAction.AddComputerGroup:    "添加计算机组",
	OperateAction.EditComputerGroup:   "修改计算机组",
	OperateAction.DeleteDssServer:     "删除下游服务器",
	OperateAction.CreateReport:        "创建报表",
	OperateAction.ExportReport:        "导出报表",
	OperateAction.DeleteReport:        "删除报表",
	OperateAction.BeginSync:           "开始同步",
	OperateAction.CancelSync:          "取消同步",
	OperateAction.SetSyncSource:       "设置同步来源",
	OperateAction.SetSyncPlan:         "设置同步计划",
	OperateAction.SetUpdateLanguage:   "设置更新语言",
	OperateAction.SetUpdateProduct:    "设置更新产品",
	OperateAction.SetUpdateClassfiy:   "设置更新分类",
	OperateAction.AcceptEula:          "接受许可协议",
	OperateAction.DownloadApproval:    "设置仅下载已审批",
	OperateAction.DownloadExpress:     "设置下载快速安装文件",
	OperateAction.ExportOperate:       "导出操作记录",
	OperateAction.ExportImage:         "导出报表图片",
	OperateAction.AddDownloadList:     "移入下载队列",
	OperateAction.RemoveDownloadList:  "移出下载队列",
}

// 操作记录执行结果
var ResultStatus = struct {
	SUCCESS, FAILED int
}{
	SUCCESS: 0,
	FAILED:  1,
}

var ResultStatusMap = map[int]string{
	ResultStatus.SUCCESS: "成功",
	ResultStatus.FAILED:  "失败",
}

// EULA状态
var EulaStatus = struct {
	NotEula, EulaNoFile, EulaFile, EulaAccept int
}{
	NotEula:    0,
	EulaNoFile: 2,
	EulaFile:   4,
	EulaAccept: 6,
}

// 重启行为
var RebootBehavior = struct {
	NeverReboots, AlwaysRequiresReboot, CanRequestReboot string
}{
	NeverReboots:         "0",
	AlwaysRequiresReboot: "1",
	CanRequestReboot:     "2",
}

var RebootBehaviorMsg = map[string]string{
	"0": "不需要重新启动",
	"1": "总是需要重新启动",
	"2": "可以请求重新启动",
}

// ReportUpdatesData 更新报表使用的结构体
type ReportUpdatesData struct {
	Title, UpdateID, ApproveStatus, AdminName, TargetGroupID, TargetGroupName, MsrcSeverity, ProductTitle, ClassificationTitle, KBArticleID, SecurityBulletinID string
	RevisionID, RevisionNumber, ActionID, ProductRevisionID, ClassificationRevisionID                                                                           *int
	CreationDate, ImportedTime, LastChangedAnchor, LastChangeTime                                                                                               *time.Time
}

// ReportComputerData 计算机报表使用的结构体
type ReportComputerData struct {
	TargetID, NotInstalled, Downloaded, InstalledPendingReboot, Failed, Installed, Unknown                                                                                   int
	OSBuildNumber, OSMajorVersion, OSMinorVersion                                                                                                                            *int
	IPAddress, ComputerMake, ComputerModel, FirmwareVersion, BiosName, BiosVersion, BiosReleaseDate, OSDescription, FullDomainName, OSVersion, MobileOperator, ClientVersion *string
	LastSyncTime, LastReportedStatusTime                                                                                                                                     *time.Time
}

// ReportDssData 下游服务器报表使用的结构体
type ReportDssData struct {
	ID                                                  int
	ServerId, FullDomainName, Version                   *string
	IsReplica                                           bool
	LastRollupTime, LastSyncTime                        *time.Time
	UpdateCount                                         int
	ApprovedUpdateCount                                 int
	NotApprovedUpdateCount                              int
	CriticalOrSecurityUpdatesNotApprovedForInstallCount int
	ExpiredUpdateCount                                  int
	DeclinedUpdateCount                                 int
	UpdatesUpToDateCount                                int
	UpdatesNeedingFilesCount                            int
	CustomComputerTargetGroupCount                      int
	ComputerTargetCount                                 int
	ComputerTargetsNeedingUpdatesCount                  int
}

// ReportUpdateInstallStats 更新安装统计报表结构体
type ReportUpdateInstallStats struct {
	Title                                                                                                   string
	RevisionID, ComputerCount, Downloaded, InstalledPendingReboot, Failed, Installed, NotInstalled, Unknown int
}

// ReportComputerUpdatesStatus 计算机更新安装状态结构体
type ReportComputerUpdatesStatus struct {
	// 更新
	RevisionID, RevisionNumber                                                                                                 *int
	Title, UpdateID, TargetGroupID, KBArticleID, MsrcSeverity, SecurityBulletinID, ProductTitle, ClassificationTitle, ActionID *string
	CreationDate, ImportedTime, LastChangedAnchor                                                                              *time.Time
	// 计算机
	TargetID, OSMajorVersion, OSMinorVersion, OSBuildNumber, Status                                                                                         *int
	FullDomainName, IPAddress, OSVersion, ComputerMake, ComputerModel, FirmwareVersion, BiosName, BiosVersion, OSDescription, MobileOperator, ClientVersion *string
	BiosReleaseDate                                                                                                                                         *time.Time
}

// 计算机更新报表必须提前查到的字段
var MustSelectFields = []string{
	"RevisionID", "KBArticleID", "SecurityBulletinID", "CreationDate", "ImportedTime", "LastChangedAnchor", "UpdateID",
	"RevisionNumber", "ProductRevisionID", "ClassificationRevisionID", "TargetID", "FullDomainName", "IPAddress", "OSVersion",
	"ComputerMake", "ComputerModel", "FirmwareVersion", "BiosName", "BiosVersion", "BiosReleaseDate", "OSDescription", "OSMajorVersion",
	"OSMinorVersion", "OSBuildNumber", "MobileOperator", "ClientVersion",
}

var NotShowFields = []string{"TargetID", "RevisionID"}

// 计算机活跃度统计周期
var ComputerLiveStatsCycle = struct {
	DAY, WEEK, MONTH int
}{
	DAY:   0, // 每天
	WEEK:  1, // 每周
	MONTH: 2, // 每月
}

// LiveStats 活跃度统计
type LiveStats struct {
	OSVersion string
	Count     int
}

// ReportRuleAdd 自定义报表规则创建编辑
type ReportRuleAdd struct {
	RuleID        int                            `json:"rule_id"`
	Name          string                         `json:"name"`
	ServerType    int                            `json:"service_type"`
	Show          []string                       `json:"show"`
	Screen        map[string]map[string][]string `json:"screen"`
	Desc          string                         `json:"desc"`
	TargetGroupID string                         `json:"TargetGroupID"`
	ActionID      int                            `json:"ActionID"`
	RevisionIDs   []int                          `json:"revision_ids"`
}

// ApproveRuleAdd 自定义审批规则创建编辑
type ApproveRuleAdd struct {
	RuleID               int    `json:"id"`
	Name                 string `json:"name"`
	TargetGroups         string `json:"target_groups"`
	TargetGroupNames     string `json:"target_group_names"`
	ProductIds           string `json:"product_ids"`
	ClassificationIds    string `json:"classification_ids"`
	Action               int    `json:"action"`
	DateOffset           int    `json:"date_offset"`
	MinutesAfterMidnight string `json:"minutes_after_midnight"`
}

// RevisionGroupDeployment 更新和组的审批关系的结构体，批量审批用
type RevisionGroupDeployment struct {
	TargetGroupID   string
	TargetGroupName string
	RevisionID      int
	ActionID        *int
	DeploymentID    *int
}

// CleanRules 清理规则
var CleanRules = struct {
	UnusedUpdates, NoConnectPC, NoNeedUpdates, ExpireUpdates, SupersededUpdates int
}{
	UnusedUpdates:     2,
	NoConnectPC:       4,
	NoNeedUpdates:     6,
	ExpireUpdates:     8,
	SupersededUpdates: 10,
}

// DeclineDeployment 批量审批拒绝时使用
type DeclineDeployment struct {
	RevisionID      int
	TargetGroupID   string
	TargetGroupName string
	ActionID        int
}

var CleanRulesDesc = map[int]string{
	CleanRules.UnusedUpdates:     "30天或更长时间未被审批的且已被取代的更新和早期修订更新。",
	CleanRules.NoConnectPC:       "删除大于30天未连接本服务器的计算机记录。",
	CleanRules.NoNeedUpdates:     "更新内容文件无对应的更新元文件，无法被计算机获取。",
	CleanRules.ExpireUpdates:     "将未审批的早期修订更新以及被终止的更新的审批状态设为拒绝。",
	CleanRules.SupersededUpdates: "将30天或更长时间未被审批的被取代的更新，审批状态设为拒绝。",
}

var CleanRulesLabel = map[int]string{
	CleanRules.UnusedUpdates:     "删除以下更新的元文件。",
	CleanRules.NoConnectPC:       "没有连接到服务器的计算机",
	CleanRules.NoNeedUpdates:     "删除无用的更新内容文件",
	CleanRules.ExpireUpdates:     "拒绝早期修订更新",
	CleanRules.SupersededUpdates: "拒绝被取代的更新",
}

var CleanRulesModule = map[int]string{
	CleanRules.UnusedUpdates:     "update",
	CleanRules.NoConnectPC:       "computer",
	CleanRules.NoNeedUpdates:     "update",
	CleanRules.ExpireUpdates:     "update",
	CleanRules.SupersededUpdates: "update",
}

// SyncDetailUpdatesMap 同步记录详情更新类型
var SyncDetailUpdatesMap = map[string]string{
	"NewUpdates":       "新更新",
	"RevisedUpdates":   "被修订更新",
	"ExpiredUpdates":   "替代更新",
	"MSExpiredUpdates": "过期更新",
}
