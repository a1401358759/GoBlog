package model

import (
	"time"
)

type ComputerTarget struct {
	VersionCheckMixin
	TargetID                       int       `gorm:"primary_key;AUTO_INCREMENT;column:TargetID" json:"TargetID"`
	IsNewClient                    bool      `gorm:"default:null;column:IsNewClient" json:"IsNewClient"`
	ComputerID                     string    `gorm:"type:varchar(256);default:null;index;column:ComputerID" json:"ComputerID"`
	SID                            string    `gorm:"type:varchar(256);default:null;column:SID" json:"SID"`
	LastSyncTime                   time.Time `gorm:"default:null;column:LastSyncTime" json:"LastSyncTime"`
	LastReportedStatusTime         time.Time `gorm:"default:null;column:LastReportedStatusTime" json:"LastReportedStatusTime"`
	LastReportedRebootTime         time.Time `gorm:"default:null;column:LastReportedRebootTime" json:"LastReportedRebootTime"`
	IPAddress                      string    `gorm:"type:varchar(56);default:null;column:IPAddress" json:"IPAddress"`
	FullDomainName                 string    `gorm:"type:varchar(256);default:null;column:FullDomainName" json:"FullDomainName"`
	IsRegistered                   bool      `gorm:"default:false;column:IsRegistered" json:"IsRegistered"`
	LastInventoryTime              time.Time `gorm:"default:null;column:LastInventoryTime" json:"LastInventoryTime"`
	LastNameChangeTime             time.Time `gorm:"default:null;column:LastNameChangeTime" json:"LastNameChangeTime"`
	EffectiveLastDetectionTime     time.Time `gorm:"default:null;column:EffectiveLastDetectionTime" json:"EffectiveLastDetectionTime"`
	ParentServerID                 string    `gorm:"type:varchar(256);default:null;column:ParentServerID" json:"ParentServerID"`
	LastSyncResult                 int       `gorm:"default:0;column:LastSyncResult" json:"LastSyncResult"`
	OSMajorVersion                 int       `gorm:"default:null;column:OSMajorVersion" json:"OSMajorVersion"`
	OSMinorVersion                 int       `gorm:"default:null;column:OSMinorVersion" json:"OSMinorVersion"`
	OSBuildNumber                  int       `gorm:"default:null;column:OSBuildNumber" json:"OSBuildNumber"`
	OSServicePackMajorNumber       int       `gorm:"default:null;column:OSServicePackMajorNumber" json:"OSServicePackMajorNumber"`
	OSServicePackMinorNumber       int       `gorm:"default:null;column:OSServicePackMinorNumber" json:"OSServicePackMinorNumber"`
	OSLocale                       string    `gorm:"type:varchar(10);default:null;column:OSLocale" json:"OSLocale"`
	ComputerMake                   string    `gorm:"type:varchar(64);default:null;column:ComputerMake" json:"ComputerMake"`
	ComputerModel                  string    `gorm:"type:varchar(64);default:null;column:ComputerModel" json:"ComputerModel"`
	BiosVersion                    string    `gorm:"type:varchar(64);default:null;column:BiosVersion" json:"BiosVersion"`
	BiosName                       string    `gorm:"type:varchar(64);default:null;column:BiosName" json:"BiosName"`
	BiosReleaseDate                string    `gorm:"type:varchar(64);default:null;column:BiosReleaseDate" json:"BiosReleaseDate"`
	ProcessorArchitecture          string    `gorm:"type:varchar(64);default:null;column:ProcessorArchitecture" json:"ProcessorArchitecture"`
	LastStatusRollupTime           time.Time `gorm:"default:null;column:LastStatusRollupTime" json:"LastStatusRollupTime"`
	LastReceivedStatusRollupNumber int       `gorm:"default:0;column:LastReceivedStatusRollupNumber" json:"LastReceivedStatusRollupNumber"`
	LastSentStatusRollupNumber     int       `gorm:"default:0;column:LastSentStatusRollupNumber" json:"LastSentStatusRollupNumber"`
	SamplingValue                  int       `gorm:"default:0;column:SamplingValue" json:"SamplingValue"`
	CreatedTime                    time.Time `gorm:"autoCreateTime;column:CreatedTime" json:"CreatedTime"`
	SuiteMask                      int       `gorm:"default:null;SMALLINT;column:SuiteMask" json:"SuiteMask"`
	OldProductType                 int       `gorm:"default:null;SMALLINT;column:OldProductType" json:"OldProductType"`
	NewProductType                 int       `gorm:"default:null;column:NewProductType" json:"NewProductType"`
	SystemMetrics                  int       `gorm:"default:null;column:SystemMetrics" json:"SystemMetrics"`
	ClientVersion                  string    `gorm:"type:varchar(23);default:null;column:ClientVersion" json:"ClientVersion"`
	TargetGroupMembershipChanged   bool      `gorm:"default:false;column:TargetGroupMembershipChanged" json:"TargetGroupMembershipChanged"`
	OSFamily                       string    `gorm:"type:varchar(256);default:null;column:OSFamily" json:"OSFamily"`
	OSDescription                  string    `gorm:"type:varchar(256);default:null;column:OSDescription" json:"OSDescription"`
	OSVersion                      string    `gorm:"type:varchar(256);default:null;column:OSVersion" json:"OSVersion"`
	OEM                            string    `gorm:"type:varchar(64);default:null;column:OEM" json:"OEM"`
	DeviceType                     string    `gorm:"type:varchar(64);default:null;column:DeviceType" json:"DeviceType"`
	FirmwareVersion                string    `gorm:"type:varchar(64);default:null;column:FirmwareVersion" json:"FirmwareVersion"`
	MobileOperator                 string    `gorm:"type:varchar(512);default:null;column:MobileOperator" json:"MobileOperator"`
	IsDelete                       bool      `gorm:"default:false;column:IsDelete" json:"IsDelete"`
	LastRollupTime                 time.Time `gorm:"default:null;column:LastRollupTime" json:"LastRollupTime"`
	RequestedTargetGroupNames      string    `gorm:"type:varchar(1024);default:null;column:RequestedTargetGroupNames" json:"RequestedTargetGroupNames"`
	TargetGroupIDList              string    `gorm:"type:varchar(1024);default:null;column:TargetGroupIDList" json:"TargetGroupIDList"`
	HasDetailsChanged              bool      `gorm:"default:false;column:HasDetailsChanged" json:"HasDetailsChanged"`
}

func (ComputerTarget) TableName() string {
	return "computer_target"
}

type ComputerTargetGroup struct {
	/*
		这个表只用于存储计算机组信息，不用于记录计算机与计算机组关系
		这里的target group会有最基本的三个组：
		B73CA6ED-5727-47F3-84DE-015E03F6A88A: Unassigned Computers
		D374F42A-9BE2-4163-A0FA-3C86A401B7A7: Downstream Servers
		A0A08746-4DBE-4A37-9ADF-9E7652C0B421: All Computers
		这里注意，前两个计算机组，都属于最后一个计算机组，targetingroup的时候，使用前两个
		deployment的时候，wsus使用Downstream Servers和All Computers
	*/
	TargetGroupID   string `gorm:"primary_key;type:char(36);column:TargetGroupID" json:"TargetGroupID"`
	TargetGroupName string `gorm:"type:varchar(256);default:null;column:TargetGroupName" json:"TargetGroupName"`
	Description     string `gorm:"type:varchar(256);default:null;column:Description" json:"Description"`
	ParentGroupID   string `gorm:"type:char(36);default:null;column:ParentGroupID" json:"ParentGroupID"`
	IsBuiltin       bool   `gorm:"default:null;column:IsBuiltin" json:"IsBuiltin"`
	OrderValue      int    `gorm:"default:null;column:OrderValue" json:"OrderValue"`
}

func (ComputerTargetGroup) TableName() string {
	return "computer_target_group"
}

type ComputerInGroup struct {
	/*
		该表用于记录每一个计算机组内有哪些计算机
	*/
	TargetGroupID       string              `gorm:"primary_key;type:char(36);column:TargetGroupID" json:"TargetGroupID"`
	ComputerTargetGroup ComputerTargetGroup `gorm:"ForeignKey:TargetGroupID;AssociationForeignKey:TargetGroupID;not null"`
	TargetID            int                 `gorm:"primary_key;column:TargetID;index" json:"TargetID"`
	ComputerTarget      ComputerTarget      `gorm:"ForeignKey:TargetID;AssociationForeignKey:TargetID;not null"`
	ComputerID          string              `gorm:"type:varchar(256);default:null;index;column:ComputerID" json:"ComputerID"`
	UpdateTime          time.Time           `gorm:"autoCreateTime;autoUpdateTime;column:UpdateTime" json:"UpdateTime"`
}

func (ComputerInGroup) TableName() string {
	return "computer_in_group"
}

type ClientComputerSummaryRollup struct {
	VersionCheckMixin
	ClientSummaryID          int    `gorm:"primary_key;column:ClientSummaryID;AUTO_INCREMENT" json:"ClientSummaryID"`
	ClientCount              int    `gorm:"default:null;column:ClientCount" json:"ClientCount"`
	OSMajorVersion           int    `gorm:"default:null;column:OSMajorVersion" json:"OSMajorVersion"`
	OSMinorVersion           int    `gorm:"default:null;column:OSMinorVersion" json:"OSMinorVersion"`
	OSBuildNumber            int    `gorm:"default:null;column:OSBuildNumber" json:"OSBuildNumber"`
	OSServicePackMajorNumber int    `gorm:"default:null;column:OSServicePackMajorNumber" json:"OSServicePackMajorNumber"`
	OSServicePackMinorNumber int    `gorm:"default:null;column:OSServicePackMinorNumber" json:"OSServicePackMinorNumber"`
	OSLocale                 string `gorm:"type:varchar(10);column:OSLocale;default:null" json:"OSLocale"`
	SuiteMask                int    `gorm:"default:null;SMALLINT;column:SuiteMask" json:"SuiteMask"`
	NewProductType           int    `gorm:"default:null;column:NewProductType" json:"NewProductType"`
	OldProductType           int    `gorm:"default:null;SMALLINT;column:OldProductType" json:"OldProductType"`
	SystemMetrics            int    `gorm:"default:null;column:SystemMetrics" json:"SystemMetrics"`
	ProcessorArchitecture    string `gorm:"type:varchar(50);column:ProcessorArchitecture;default:null" json:"ProcessorArchitecture"`
}

func (ClientComputerSummaryRollup) TableName() string {
	return "client_computer_summary_rollup"
}

type ClientComputerActivityRollup struct {
	VersionCheckMixin
	ID                          int                         `gorm:"primary_key;column:id" json:"id"`
	ClientSummaryID             int                         `gorm:"null;column:ClientSummaryID" json:"ClientSummaryID"`
	ClientComputerSummaryRollup ClientComputerSummaryRollup `gorm:"ForeignKey:ClientSummaryID;AssociationForeignKey:ClientSummaryID;not null;"`
	UpdateID                    string                      `gorm:"type:varchar(36);default:null;column:UpdateID" json:"UpdateID"`
	RevisionNumber              int                         `gorm:"index;not null;column:RevisionNumber" json:"RevisionNumber"`
	ServerID                    string                      `gorm:"type:char(36);column:ServerID;default:null" json:"ServerID"`
	InstallSuccessCount         int                         `gorm:"default:null;column:InstallSuccessCount" json:"InstallSuccessCount"`
	InstallFailureCount         int                         `gorm:"default:null;column:InstallFailureCount" json:"InstallFailureCount"`
}

func (ClientComputerActivityRollup) TableName() string {
	return "client_computer_activity_rollup"
}

type ComputerSummaryForMicrosoftUpdates struct {
	VersionCheckMixin
	TargetID               int            `gorm:"primary_key;column:TargetID;index" json:"TargetID"`
	ComputerTarget         ComputerTarget `gorm:"ForeignKey:TargetID;AssociationForeignKey:TargetID;not null"`
	Unknown                int            `gorm:"default:null;column:Unknown" json:"Unknown"`
	NotInstalled           int            `gorm:"default:null;column:NotInstalled" json:"NotInstalled"`
	Downloaded             int            `gorm:"default:null;column:Downloaded" json:"Downloaded"`
	Installed              int            `gorm:"default:null;column:Installed" json:"Installed"`
	Failed                 int            `gorm:"default:null;column:Failed" json:"Failed"`
	InstalledPendingReboot int            `gorm:"default:null;column:InstalledPendingReboot" json:"InstalledPendingReboot"`
	LastChangeTime         time.Time      `gorm:"default:null;column:LastChangeTime" json:"LastChangeTime"`
	MaxReportTime          time.Time      `gorm:"default:null;column:MaxReportTime" json:"MaxReportTime"`
}

func (ComputerSummaryForMicrosoftUpdates) ComputerSummaryForMicrosoftUpdates() string {
	return "computer_summary_for_microsoft_updates"
}

type Deployment struct {
	VersionCheckMixin
	DeploymentID         int                 `gorm:"primary_key;column:DeploymentID" json:"DeploymentID"`
	RevisionID           int                 `gorm:"not null;column:RevisionID" json:"RevisionID"`
	Revision             Revision            `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	TargetGroupID        string              `gorm:"type:char(36);column:TargetGroupID;index" json:"TargetGroupID"`
	ComputerTargetGroup  ComputerTargetGroup `gorm:"ForeignKey:TargetGroupID;AssociationForeignKey:TargetGroupID;not null"`
	TargetGroupName      string              `gorm:"type:varchar(256);default:null;column:TargetGroupName" json:"TargetGroupName"`
	LastChangeTime       time.Time           `gorm:"autoCreateTime;column:LastChangeTime" json:"LastChangeTime"`
	LastChangeNumber     int                 `gorm:"default:null;column:LastChangeNumber" json:"LastChangeNumber"`
	DeploymentStatus     int                 `gorm:"default:null;SMALLINT;column:DeploymentStatus" json:"DeploymentStatus"`
	ActionID             int                 `gorm:"default:null;column:ActionID" json:"ActionID"`
	DeploymentTime       time.Time           `gorm:"default:null;column:DeploymentTime" json:"DeploymentTime"`
	GoLiveTime           time.Time           `gorm:"default:null;column:GoLiveTime" json:"GoLiveTime"`
	Deadline             time.Time           `gorm:"default:null;column:Deadline" json:"Deadline"`
	AdminName            string              `gorm:"type:varchar(385);default:null;column:AdminName" json:"AdminName"`
	DownloadPriority     int                 `gorm:"default:null;SMALLINT;column:DownloadPriority" json:"DownloadPriority"`
	DeploymentGuid       string              `gorm:"type:char(36);column:DeploymentGuid;default:null" json:"DeploymentGuid"`
	IsAssigned           bool                `gorm:"default:null;column:IsAssigned" json:"IsAssigned"`
	IsLeaf               bool                `gorm:"default:null;column:IsLeaf" json:"IsLeaf"`
	UpdateType           string              `gorm:"type:varchar(256);default:null;column:UpdateType" json:"UpdateType"`
	IsCritical           bool                `gorm:"default:null;column:IsCritical" json:"IsCritical"`
	Priority             int                 `gorm:"default:null;column:Priority" json:"Priority"`
	IsFeatured           bool                `gorm:"default:null;column:IsFeatured" json:"IsFeatured"`
	AutoSelect           int                 `gorm:"default:null;SMALLINT;column:AutoSelect" json:"AutoSelect"`
	AutoDownload         int                 `gorm:"default:null;SMALLINT;column:AutoDownload" json:"AutoDownload"`
	SupersedenceBehavior int                 `gorm:"default:null;SMALLINT;column:SupersedenceBehavior" json:"SupersedenceBehavior"`
	IsPartOfSet          bool                `gorm:"default:null;column:IsPartOfSet" json:"IsPartOfSet"`
	SetCreationTime      time.Time           `gorm:"default:null;column:SetCreationTime" json:"SetCreationTime"`
}

func (Deployment) TableName() string {
	return "deployment"
}

type DeadDeployment struct {
	VersionCheckMixin
	DeploymentID         int       `gorm:"primary_key;column:DeploymentID" json:"DeploymentID"`
	RevisionID           int       `gorm:"not null;column:RevisionID" json:"RevisionID"`
	TargetGroupID        string    `gorm:"type:char(36);column:TargetGroupID;index" json:"TargetGroupID"`
	TargetGroupName      string    `gorm:"type:varchar(256);default:null;column:TargetGroupName" json:"TargetGroupName"`
	LastChangeTime       time.Time `gorm:"autoCreateTime;column:LastChangeTime" json:"LastChangeTime"`
	LastChangeNumber     int64     `gorm:"default:null;column:LastChangeNumber" json:"LastChangeNumber"`
	DeploymentStatus     int       `gorm:"default:null;SMALLINT;column:DeploymentStatus" json:"DeploymentStatus"`
	ActionID             int       `gorm:"default:null;column:ActionID" json:"ActionID"`
	DeploymentTime       time.Time `gorm:"default:null;column:DeploymentTime" json:"DeploymentTime"`
	GoLiveTime           time.Time `gorm:"default:null;column:GoLiveTime" json:"GoLiveTime"`
	Deadline             time.Time `gorm:"default:null;column:Deadline" json:"Deadline"`
	AdminName            string    `gorm:"type:varchar(385);default:null;column:AdminName" json:"AdminName"`
	DownloadPriority     int       `gorm:"default:null;SMALLINT;column:DownloadPriority" json:"DownloadPriority"`
	DeploymentGuid       string    `gorm:"type:char(36);column:DeploymentGuid;default:null" json:"DeploymentGuid"`
	IsAssigned           bool      `gorm:"default:null;column:IsAssigned" json:"IsAssigned"`
	IsLeaf               bool      `gorm:"default:null;column:IsLeaf" json:"IsLeaf"`
	UpdateType           string    `gorm:"type:varchar(256);default:null;column:UpdateType" json:"UpdateType"`
	IsCritical           bool      `gorm:"default:null;column:IsCritical" json:"IsCritical"`
	Priority             int       `gorm:"default:null;column:Priority" json:"Priority"`
	IsFeatured           bool      `gorm:"default:null;column:IsFeatured" json:"IsFeatured"`
	AutoSelect           int       `gorm:"default:null;SMALLINT;column:AutoSelect" json:"AutoSelect"`
	AutoDownload         int       `gorm:"default:null;SMALLINT;column:AutoDownload" json:"AutoDownload"`
	SupersedenceBehavior int       `gorm:"default:null;SMALLINT;column:SupersedenceBehavior" json:"SupersedenceBehavior"`
	IsPartOfSet          bool      `gorm:"default:null;column:IsPartOfSet" json:"IsPartOfSet"`
	SetCreationTime      time.Time `gorm:"default:null;column:SetCreationTime" json:"SetCreationTime"`
}

func (DeadDeployment) TableName() string {
	return "dead_deployment"
}

type UpdateStatus struct {
	VersionCheckMixin
	RevisionID             int       `gorm:"primary_key;column:RevisionID;index" json:"RevisionID"`
	Revision               Revision  `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	UpdateID               string    `gorm:"type:varchar(36);default:null;column:UpdateID" json:"UpdateID"`
	UpdateState            int       `gorm:"default:null;column:UpdateState;comment:计算机针对该更新的更新结果,0=失败/1=成功" json:"UpdateState"`
	ComputerID             string    `gorm:"type:varchar(256);primary_key;column:ComputerID" json:"ComputerID"`
	LastRefreshTime        time.Time `gorm:"default:null;column:LastRefreshTime" json:"LastRefreshTime"`
	LastChangeTime         time.Time `gorm:"default:null;column:LastChangeTime" json:"LastChangeTime"`
	LastChangeTimeOnServer time.Time `gorm:"default:null;column:LastChangeTimeOnServer" json:"LastChangeTimeOnServer"`
}

func (UpdateStatus) TableName() string {
	return "update_status"
}

type StatisticsForPerUpdate struct {
	VersionCheckMixin
	RevisionID             int       `gorm:"primary_key;column:RevisionID" json:"RevisionID"`
	Revision               Revision  `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	Unknown                int       `gorm:"default:null;column:Unknown" json:"Unknown"`
	NotInstalled           int       `gorm:"default:null;column:NotInstalled" json:"NotInstalled"`
	Downloaded             int       `gorm:"default:null;column:Downloaded" json:"Downloaded"`
	Installed              int       `gorm:"default:null;column:Installed" json:"Installed"`
	Failed                 int       `gorm:"default:null;column:Failed" json:"Failed"`
	InstalledPendingReboot int       `gorm:"default:null;column:InstalledPendingReboot" json:"InstalledPendingReboot"`
	LastChangeTime         time.Time `gorm:"default:null;column:LastChangeTime;" json:"LastChangeTime"`
	MaxReportTime          time.Time `gorm:"default:null;column:MaxReportTime" json:"MaxReportTime"`
}

func (StatisticsForPerUpdate) TableName() string {
	return "statistics_for_per_update"
}

type MSRCSeverityStatistics struct {
	ID             int       `gorm:"primary_key;column:id;AUTO_INCREMENT" json:"id"`
	UnspecifiedNum int       `gorm:"default:null;column:unspecified_num" json:"unspecified_num"`
	LowNum         int       `gorm:"default:null;column:low_num" json:"low_num"`
	ModerateNum    int       `gorm:"default:null;column:moderate_num" json:"moderate_num"`
	ImportantNum   int       `gorm:"default:null;column:important_num" json:"important_num"`
	CriticalNum    int       `gorm:"default:null;column:critical_num" json:"critical_num"`
	LastUpdate     time.Time `gorm:"not null;column:last_update" json:"last_update"`
	MaxStaticsTime time.Time `gorm:"default:null;column:max_statics_time" json:"max_statics_time"`
}

func (MSRCSeverityStatistics) TableName() string {
	return "msrc_severity_statistics"
}

type SyncHistory struct {
	VersionCheckMixin
	ID               int       `gorm:"primary_key;column:id" json:"id"`
	ParentServerID   string    `gorm:"type:char(36);column:ParentServerID;default:null" json:"ParentServerID"`
	ParentServerIP   string    `gorm:"type:varchar(256);column:ParentServerIP;default:null" json:"ParentServerIP"`
	LastSyncTime     time.Time `gorm:"default:null;column:LastSyncTime" json:"LastSyncTime"`
	ImportedTime     time.Time `gorm:"autoCreateTime;column:ImportedTime" json:"ImportedTime"`
	StartTime        time.Time `gorm:"default:null;column:StartTime;index" json:"StartTime"`
	FinishTime       time.Time `gorm:"default:null;column:FinishTime;index" json:"FinishTime"`
	SyncType         int       `gorm:"default:null;column:SyncType" json:"SyncType"`
	SyncStatus       int       `gorm:"default:null;column:SyncStatus" json:"SyncStatus"`
	NewUpdates       int       `gorm:"default:null;column:NewUpdates" json:"NewUpdates"`
	RevisedUpdates   int       `gorm:"default:null;column:RevisedUpdates" json:"RevisedUpdates"`
	ExpiredUpdates   int       `gorm:"default:null;column:ExpiredUpdates" json:"ExpiredUpdates"`
	MSExpiredUpdates int       `gorm:"default:null;column:MSExpiredUpdates" json:"MSExpiredUpdates"`
	Pending          bool      `gorm:"default:null;column:Pending" json:"Pending"`
	ReplicationMode  string    `gorm:"type:varchar(20);column:ReplicationMode;default:null;default:'Auto'" json:"ReplicationMode"`
	SyncCategory     int       `gorm:"default:null;column:SyncCategory" json:"SyncCategory"`
}

func (SyncHistory) TableName() string {
	return "sync_history"
}

type Dss struct {
	ID                                                  int       `gorm:"primary_key;column:id" json:"id"`
	ServerId                                            string    `gorm:"type:char(36);column:ServerId;default:null" json:"ServerId"`
	FullDomainName                                      string    `gorm:"type:varchar(256);column:FullDomainName;default:null" json:"FullDomainName"`
	ParentServerId                                      string    `gorm:"type:char(36);column:ParentServerId;default:null" json:"ParentServerId"`
	LastRollupTime                                      time.Time `gorm:"default:null;column:LastRollupTime" json:"LastRollupTime"`
	InstallSuccessCount                                 int       `gorm:"default:null;column:InstallSuccessCount" json:"InstallSuccessCount"`
	InstallFailureCount                                 int       `gorm:"default:null;column:InstallFailureCount" json:"InstallFailureCount"`
	LastSyncTime                                        time.Time `gorm:"default:null;column:LastSyncTime" json:"LastSyncTime"`
	Version                                             string    `gorm:"type:varchar(256);column:Version;default:null" json:"Version"`
	IsReplica                                           bool      `gorm:"default:null;column:IsReplica" json:"IsReplica"`
	UpdateCount                                         int       `gorm:"default:null;column:UpdateCount" json:"UpdateCount"`
	DeclinedUpdateCount                                 int       `gorm:"default:null;column:DeclinedUpdateCount" json:"DeclinedUpdateCount"`
	ApprovedUpdateCount                                 int       `gorm:"default:null;column:ApprovedUpdateCount" json:"ApprovedUpdateCount"`
	NotApprovedUpdateCount                              int       `gorm:"default:null;column:NotApprovedUpdateCount" json:"NotApprovedUpdateCount"`
	UpdatesWithStaleUpdateApprovalsCount                int       `gorm:"default:null;column:UpdatesWithStaleUpdateApprovalsCount" json:"UpdatesWithStaleUpdateApprovalsCount"`
	ExpiredUpdateCount                                  int       `gorm:"default:null;column:ExpiredUpdateCount" json:"ExpiredUpdateCount"`
	CriticalOrSecurityUpdatesNotApprovedForInstallCount int       `gorm:"default:null;column:CriticalOrSecurityUpdatesNotApprovedForInstallCount" json:"CriticalOrSecurityUpdatesNotApprovedForInstallCount"`
	WsusInfrastructureUpdatesNotApprovedForInstallCount int       `gorm:"default:null;column:WsusInfrastructureUpdatesNotApprovedForInstallCount" json:"WsusInfrastructureUpdatesNotApprovedForInstallCount"`
	UpdatesWithClientErrorsCount                        int       `gorm:"default:null;column:UpdatesWithClientErrorsCount" json:"UpdatesWithClientErrorsCount"`
	UpdatesWithServerErrorsCount                        int       `gorm:"default:null;column:UpdatesWithServerErrorsCount" json:"UpdatesWithServerErrorsCount"`
	UpdatesNeedingFilesCount                            int       `gorm:"default:null;column:UpdatesNeedingFilesCount" json:"UpdatesNeedingFilesCount"`
	UpdatesNeededByComputersCount                       int       `gorm:"default:null;column:UpdatesNeededByComputersCount" json:"UpdatesNeededByComputersCount"`
	UpdatesUpToDateCount                                int       `gorm:"default:null;column:UpdatesUpToDateCount" json:"UpdatesUpToDateCount"`
	CustomComputerTargetGroupCount                      int       `gorm:"default:null;column:CustomComputerTargetGroupCount" json:"CustomComputerTargetGroupCount"`
	ComputerTargetCount                                 int       `gorm:"default:null;column:ComputerTargetCount" json:"ComputerTargetCount"`
	ComputerTargetsNeedingUpdatesCount                  int       `gorm:"default:null;column:ComputerTargetsNeedingUpdatesCount" json:"ComputerTargetsNeedingUpdatesCount"`
	ComputerTargetsWithUpdateErrorsCount                int       `gorm:"default:null;column:ComputerTargetsWithUpdateErrorsCount" json:"ComputerTargetsWithUpdateErrorsCount"`
	ComputersUpToDateCount                              int       `gorm:"default:null;column:ComputersUpToDateCount" json:"ComputersUpToDateCount"`
	ImportedTime                                        time.Time `gorm:"autoCreateTime;column:ImportedTime" json:"ImportedTime"`
}

func (Dss) TableName() string {
	return "dss"
}

type EulaAcceptance struct {
	VersionCheckMixin
	ID           int       `gorm:"primary_key;column:id" json:"id"`
	EulaID       string    `gorm:"unique;type:varchar(60);comment:'许可协议ID" json:"eula_id"`
	AcceptedDate time.Time `gorm:"comment:接受时间" json:"accepted_date"`
	AdminName    string    `gorm:"type:varchar(60);comment:接受人" json:"admin_name"`
}

func (EulaAcceptance) TableName() string {
	return "eula_acceptance"
}

type AutoApproveRules struct {
	VersionCheckMixin
	ID                   int    `gorm:"primary_key;column:id" json:"id"`
	Name                 string `gorm:"unique;type:varchar(256)" json:"name"`
	UserID               int    `gorm:"default:null;column:user_id;comment:User表的id字段，这里没有使用外键，因为怕删除User表时报错" json:"user_id"`
	TargetGroups         string `gorm:"type:varchar(5000);comment:组ID" json:"target_groups"`
	TargetGroupNames     string `gorm:"type:varchar(5000);comment:组名称" json:"target_group_names"`
	ProductIds           string `gorm:"type:varchar(256);column:product_ids;comment:产品" json:"product_ids"`
	ClassificationIds    string `gorm:"type:varchar(256);comment:分类" json:"classification_ids"`
	Action               *int   `gorm:"comment:审批行为" json:"action"`
	DateOffset           int    `gorm:"default:null" json:"date_offset"`
	MinutesAfterMidnight int    `gorm:"default:null" json:"minutes_after_midnight"`
	IsValid              bool   `gorm:"default:null;default:1;comment:是否可用" json:"is_valid"`
}

func (AutoApproveRules) TableName() string {
	return "auto_approve_rules"
}

type OperateRecord struct {
	VersionCheckMixin
	ID            int       `gorm:"primary_key;column:id" json:"id"`
	OperateID     int       `gorm:"column:operator_id;comment:操作人ID" json:"operate_id"`
	Operator      string    `gorm:"type:varchar(60);comment:操作人" json:"operator"`
	OperateTime   time.Time `gorm:"autoCreateTime;comment:操作时间" json:"operate_time"`
	Operate       int       `gorm:"comment:操作" json:"operate"`
	OperateDesc   string    `gorm:"type:varchar(1024);comment:操作描述" json:"operate_desc"`
	OperateResult int       `gorm:"comment:操作结果" json:"operate_result"`
}

func (OperateRecord) TableName() string {
	return "operate_record"
}
