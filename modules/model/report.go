package model

import "time"

type BasicData struct {
	/*
		ReportingWebService中的基础数据表
	*/
	ID              uint      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	TargetID        string    `gorm:"type:char(36);column:TargetID;index;comment:客户端计算机的标识(与clientID参数相同);unique_index:teurr" json:"TargetID"`
	SequenceNumber  int       `gorm:"column:SequenceNumber;default:0;comment:必须设置为0,收到后必须忽略" json:"SequenceNumber"`
	TimeAtTarget    time.Time `gorm:"column:TimeAtTarget;not null;comment:客户端记录事件时的(UTC)时间" json:"TimeAtTarget"`
	EventInstanceID string    `gorm:"column:EventInstanceID;default:null;comment:客户端生成的GUID,用于唯一标识此事件的发生" json:"EventInstanceID"`
	NamespaceID     int       `gorm:"column:NamespaceID;default:1;comment:所有客户端必须设置为1" json:"NamespaceID"`
	EventID         int       `gorm:"column:EventID;not default:null;comment:用于标识客户端上发生的事件的类型;unique_index:teurr" json:"EventID"`
	SourceID        int       `gorm:"column:SourceID;default:null;comment:定义生成事件的客户端中的子组件" json:"SourceID"`
	UpdateID        string    `gorm:"column:UpdateID;type:char(36);index;not null;comment:该修订版本对应的更新的GUID;unique_index:teurr" json:"UpdateID"`
	RevisionNumber  int       `gorm:"column:RevisionNumber;not null;comment:单一更新的修订版本号，用来标识同一更新的不同修订版本;unique_index:teurr" json:"RevisionNumber"`
	Win32HResult    int       `gorm:"column:Win32HResult;default:null;comment:可选择为与故障相对应的事件指定Win32 HRESULT代码" json:"Win32HResult"`
	AppName         string    `gorm:"column:AppName;default:null;comment:触发客户端执行操作的应用程序的名称" json:"AppName"`
	MiscData        string    `gorm:"column:MiscData;type:longtext;default:null;comment:reporting的杂项数据" json:"MiscData"`
	ReportedTime    time.Time `gorm:"autoCreateTime;column:ReportedTime" json:"ReportedTime"`
	ReportedDate    time.Time `gorm:"column:ReportedDate;comment:数据上报日期;unique_index:teurr" json:"ReportedDate"` // 该字段是为了保证同一天相同TargetID，EventID，UpdateID，RevisionNumber的数据只记录一次
}

func (BasicData) TableName() string {
	return "basic_data"
}

type ExtendedData struct {
	/*
		ReportingWebService中的扩展数据表
	*/
	ID                    uint      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	BasicID               uint      `gorm:"index;column:BasicID;comment:关联basic_data中的id字段" json:"BasicID"`
	BasicData             BasicData `gorm:"ForeignKey:BasicID;AssociationForeignKey:BasicID;not null"`
	ReplacementStrings    string    `gorm:"column:ReplacementStrings;type:longtext;default:null;comment:占位符" json:"ReplacementStrings"`
	ComputerBrand         string    `gorm:"column:ComputerBrand;type:varchar(256);default:null;comment:客户端计算机制造商" json:"ComputerBrand"`
	ComputerModel         string    `gorm:"column:ComputerModel;type:varchar(256);default:null;comment:客户端计算机型号名称" json:"ComputerModel"`
	BiosRevision          string    `gorm:"column:BiosRevision;type:varchar(256);default:null;comment:客户端BIOS固件版本" json:"BiosRevision"`
	ProcessorArchitecture string    `gorm:"column:ProcessorArchitecture;type:varchar(256);default:null;comment:客户端计算机CPU的体系结构" json:"ProcessorArchitecture"`
	// Unknown/X86Compatible/IA64Compatible/Amd64Compatible
	// OSVersion:客户端操作系统版本
	Major            int `gorm:"column:Major;default:null" json:"Major"`
	Minor            int `gorm:"column:Minor;default:null" json:"Minor"`
	Build            int `gorm:"column:Build;default:null" json:"Build"`
	Revision         int `gorm:"column:Revision;default:null" json:"Revision"`
	ServicePackMajor int `gorm:"column:ServicePackMajor;default:null" json:"ServicePackMajor"`
	ServicePackMinor int `gorm:"column:ServicePackMinor;default:null" json:"ServicePackMinor"`

	OSLocaleID int    `gorm:"column:OSLocaleID;default:null;comment:客户端操作系统区域设置" json:"OSLocaleID"`
	DeviceID   string `gorm:"column:DeviceID;default:null;type:char(36);comment:依赖于event的字符串" json:"DeviceID"`
	// PrivateData: 此字段必须存在且为空,并且必须在收到时忽略。
	ComputerDnsName string `gorm:"column:ComputerDnsName;default:null;type:char(36)" json:"ComputerDnsName"`
	UserAccountName string `gorm:"column:UserAccountName;default:null;type:char(36)" json:"UserAccountName"`
}

func (ExtendedData) TableName() string {
	return "extended_data"
}

type MiscData struct {
	ID            uint      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	BasicID       uint      `gorm:"index;column:BasicID;comment:关联basic_data中的id字段" json:"BasicID"`
	BasicData     BasicData `gorm:"ForeignKey:BasicID;AssociationForeignKey:BasicID;not null"`
	MiscDataTag   string    `gorm:"column:MiscDataTag;default:null;type:char(255);comment:数据标签" json:"MiscDataTag"`
	MiscDataValue string    `gorm:"column:MiscDataValue;default:null;type:varchar(5000);comment:misc data tag对应的值" json:"MiscDataValue"`
}

func (MiscData) TableName() string {
	return "misc_data"
}

type UpdateStatusPerComputer struct {
	/*
		电脑的更新状态
	*/
	ID                     uint           `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	Status                 int            `gorm:"column:Status;default:null" json:"Status"`
	RevisionID             int            `gorm:"column:RevisionID" json:"RevisionID"`
	Revision               Revision       `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	TargetID               int            `gorm:"column:TargetID" json:"TargetID"`
	ComputerTarget         ComputerTarget `gorm:"ForeignKey:TargetID;AssociationForeignKey:TargetID;not null"`
	LastChangeTime         time.Time      `gorm:"column:LastChangeTime" json:"LastChangeTime"`
	LastChangeTimeOnServer time.Time      `gorm:"column:LastChangeTimeOnServer" json:"LastChangeTimeOnServer"`
}

func (UpdateStatusPerComputer) TableName() string {
	return "update_status_per_computer"
}

type RevisionStatement struct {
	ID uint `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	// 必选
	Title          string   `gorm:"column:Title;type:varchar(256);default:null;comment:更新名称" json:"Title"`
	RevisionID     int      `gorm:"column:RevisionID" json:"RevisionID"`
	Revision       Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	UpdateID       string   `gorm:"column:UpdateID;type:char(36);comment:该update的GUID" json:"UpdateID"`
	RevisionNumber int      `gorm:"column:RevisionNumber" json:"RevisionNumber"`
	// 可筛选
	MsrcSeverity             string    `gorm:"column:MsrcSeverity;type:char(36);comment:MSRC严重程度" json:"MsrcSeverity"`
	ProductRevisionID        int       `gorm:"index;column:ProductRevisionID;comment:如果是software，记录属于的产品的RevisionID" json:"ProductRevisionID"`
	ClassificationRevisionID int       `gorm:"index;column:ClassificationRevisionID;comment:如果是software，记录属于的分类的RevisionID" json:"ClassificationRevisionID"`
	ProductTitle             string    `gorm:"column:ProductTitle;type:varchar(256);default:null;comment:所属产品名称" json:"ProductTitle"`
	ClassificationTitle      string    `gorm:"column:ClassificationTitle;type:varchar(256);default:null;comment:所属分类名称" json:"ClassificationTitle"`
	KBArticleID              string    `gorm:"column:KBArticleID;type:char(36);default:null;comment:KB号" json:"KBArticleID"`
	SecurityBulletinID       string    `gorm:"column:SecurityBulletinID;type:char(36);default:null;comment:MSRC编号" json:"SecurityBulletinID"`
	CreationDate             time.Time `gorm:"column:CreationDate;default:null;comment:发布日期" json:"CreationDate"`
	ImportedTime             time.Time `gorm:"column:ImportedTime;default:null;comment:到达日期" json:"ImportedTime"`
	LastChangedAnchor        time.Time `gorm:"column:LastChangedAnchor;default:null;comment:修订日期" json:"LastChangedAnchor"`
	// 组合
	ActionID        int       `gorm:"column:ActionID;default:null" json:"ActionID"`
	ApproveStatus   string    `gorm:"column:ApproveStatus;default:null;type:varchar(64);comment:审批状态" json:"ApproveStatus"`
	DeploymentTime  time.Time `gorm:"column:DeploymentTime;default:null;comment:审批时间" json:"DeploymentTime"`
	AdminName       string    `gorm:"column:AdminName;default:null;type:varchar(385);comment:审批人" json:"AdminName"`
	TargetGroupID   string    `gorm:"column:TargetGroupID;type:char(36);comment:计算机组ID" json:"TargetGroupID"`
	TargetGroupName string    `gorm:"column:TargetGroupName;type:varchar(256);comment:计算机组名称" json:"TargetGroupName"`
	// 服务提出，后续可能加筛选
	IsSuperseded bool      `gorm:"column:IsSuperseded;default:null;comment:是否被取代" json:"IsSuperseded"`
	FileStatus   bool      `gorm:"column:FileStatus;default:null;comment:文件状态: 完成, 未完成" json:"FileStatus"`
	StatsTime    time.Time `gorm:"autoCreateTime;column:StatsTime;comment:统计时间" json:"StatsTime"`
}

func (RevisionStatement) TableName() string {
	return "revision_statement"
}

type ComputerStatement struct {
	ID             uint           `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	TargetID       int            `gorm:"column:TargetID" json:"TargetID"`
	ComputerTarget ComputerTarget `gorm:"ForeignKey:TargetID;AssociationForeignKey:TargetID;not null"`
	// 必须
	FullDomainName         string    `gorm:"column:FullDomainName;default:null;type:varchar(256);comment:计算机名称" json:"FullDomainName"`
	LastSyncTime           time.Time `gorm:"column:LastSyncTime;default:null;comment:上次同步时间" json:"LastSyncTime"`
	LastReportedStatusTime time.Time `gorm:"column:LastReportedStatusTime;default:null;comment:上次报告状态时间" json:"LastReportedStatusTime"`
	// 可筛选
	OSVersion              string `gorm:"column:OSVersion;default:null;type:varchar(256);comment:操作系统版本" json:"OSVersion"`
	NotInstalled           int    `gorm:"column:NotInstalled;default:null;comment:需要安装的更新总数" json:"NotInstalled"`
	Downloaded             int    `gorm:"column:Downloaded;default:null;comment:下载但未安装的更新总数" json:"Downloaded"`
	InstalledPendingReboot int    `gorm:"column:InstalledPendingReboot;default:null;comment:安装但需要重启的计算机更新总数" json:"InstalledPendingReboot"`
	Failed                 int    `gorm:"column:Failed;default:null;comment:安装失败的更新总数" json:"Failed"`
	Installed              int    `gorm:"column:Installed;default:null;comment:已安装的更新总数" json:"Installed"`
	Unknown                int    `gorm:"column:Unknown;default:null;comment:状态未知的更新总数" json:"Unknown"`
	// 自定义报表中可以按条件选
	IPAddress       string    `gorm:"column:IPAddress;default:null;type:varchar(56);comment:ip地址" json:"IPAddress"`
	ComputerMake    string    `gorm:"column:ComputerMake;default:null;type:varchar(64);comment:制造商" json:"ComputerMake"`
	ComputerModel   string    `gorm:"column:ComputerModel;default:null;type:varchar(64);comment:型号" json:"ComputerModel"`
	FirmwareVersion string    `gorm:"column:FirmwareVersion;default:null;type:varchar(64);comment:固件" json:"FirmwareVersion"`
	BiosName        string    `gorm:"column:BiosName;default:null;type:varchar(64);comment:BIOS名称" json:"BiosName"`
	BiosVersion     string    `gorm:"column:BiosVersion;default:null;type:varchar(64);comment:BIOS版本" json:"BiosVersion"`
	BiosReleaseDate string    `gorm:"column:BiosReleaseDate;default:null;type:varchar(64);comment:BIOS发布时间" json:"BiosReleaseDate"`
	OSDescription   string    `gorm:"column:OSDescription;default:null;type:varchar(256);comment:操作系统架构" json:"OSDescription"`
	OSMajorVersion  int       `gorm:"column:OSMajorVersion;default:null;comment:操作系统主版本号" json:"OSMajorVersion"`
	OSMinorVersion  int       `gorm:"column:OSMinorVersion;default:null;comment:操作系统次版本号" json:"OSMinorVersion"`
	OSBuildNumber   int       `gorm:"column:OSBuildNumber;default:null;comment:操作系统Build号" json:"OSBuildNumber"`
	MobileOperator  string    `gorm:"column:MobileOperator;default:null;type:varchar(15);comment:移动运营商" json:"MobileOperator"`
	ClientVersion   string    `gorm:"column:ClientVersion;default:null;type:varchar(23);comment:客户端版本" json:"ClientVersion"`
	StatsTime       time.Time `gorm:"autoCreateTime;column:StatsTime;comment:统计时间" json:"StatsTime"`
}

func (ComputerStatement) TableName() string {
	return "computer_statement"
}

type DSSStatement struct {
	// 必须
	ID             uint      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	DssID          int       `gorm:"column:DssID" json:"DssID"`
	Dss            Dss       `gorm:"ForeignKey:DssID;AssociationForeignKey:DssID;not null"`
	LastSyncTime   time.Time `gorm:"column:LastSyncTime;default:null;comment:上次同步时间" json:"LastSyncTime"`
	ServerId       string    `gorm:"column:ServerId;type:char(36);comment:服务器ID" json:"ServerId"`
	FullDomainName string    `gorm:"column:FullDomainName;default:null;type:varchar(256);comment:服务器名称" json:"FullDomainName"`
	// 可筛选
	IsReplica string `gorm:"column:IsReplica;type:char(36);default:null;comment:同步模式" json:"IsReplica"`
	// 按条件筛选
	LastRollupTime time.Time `gorm:"column:LastRollupTime;default:null;comment:上次汇总时间" json:"LastRollupTime"`
	Version        string    `gorm:"column:Version;default:null;type:varchar(256);comment:版本" json:"Version"`

	UpdateCount                                         int       `gorm:"column:UpdateCount;default:null;comment:更新的总数" json:"UpdateCount"`
	ApprovedUpdateCount                                 int       `gorm:"column:ApprovedUpdateCount;default:null;comment:已审批的更新的总数" json:"ApprovedUpdateCount"`
	NotApprovedUpdateCount                              int       `gorm:"column:NotApprovedUpdateCount;default:null;comment:未审批的更新的总数" json:"NotApprovedUpdateCount"`
	CriticalOrSecurityUpdatesNotApprovedForInstallCount int       `gorm:"column:CriticalOrSecurityUpdatesNotApprovedForInstallCount;default:null;comment:未审批的关键更新的总数" json:"CriticalOrSecurityUpdatesNotApprovedForInstallCount"`
	ExpiredUpdateCount                                  int       `gorm:"column:ExpiredUpdateCount;default:null;comment:已过期的更新的总数" json:"ExpiredUpdateCount"`
	DeclinedUpdateCount                                 int       `gorm:"column:DeclinedUpdateCount;default:null;comment:已拒绝的更新总数" json:"DeclinedUpdateCount"`
	UpdatesUpToDateCount                                int       `gorm:"column:UpdatesUpToDateCount;default:null;comment:需要安装的更新的总数" json:"UpdatesUpToDateCount"`
	UpdatesNeedingFilesCount                            int       `gorm:"column:UpdatesNeedingFilesCount;default:null;comment:需要下载文件的更新的总数" json:"UpdatesNeedingFilesCount"`
	CustomComputerTargetGroupCount                      int       `gorm:"column:CustomComputerTargetGroupCount;default:null;comment:自定义组的总数" json:"CustomComputerTargetGroupCount"`
	ComputerTargetCount                                 int       `gorm:"column:ComputerTargetCount;default:null;comment:计算机的总数" json:"ComputerTargetCount"`
	ComputerTargetsNeedingUpdatesCount                  int       `gorm:"column:ComputerTargetsNeedingUpdatesCount;default:null;comment:需要安装更新的计算机的总数" json:"ComputerTargetsNeedingUpdatesCount"`
	StatsTime                                           time.Time `gorm:"autoCreateTime;column:StatsTime;comment:统计时间" json:"StatsTime"`
}

func (DSSStatement) TableName() string {
	return "dss_statement"
}

type RevisionInstallStatistics struct {
	/*
		更新安装统计
	*/
	ID                     uint      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	RevisionID             int       `gorm:"column:RevisionID" json:"RevisionID"`
	Revision               Revision  `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	Title                  string    `gorm:"column:Title;type:varchar(256);default:null;comment:更新名称" json:"Title"`
	ComputerCount          int       `gorm:"column:ComputerCount;default:null;comment:计算机总数" json:"ComputerCount"`
	Downloaded             int       `gorm:"column:Downloaded;default:null;comment:下载但未安装的计算机总数" json:"Downloaded"`
	InstalledPendingReboot int       `gorm:"column:InstalledPendingReboot;default:null;comment:安装但需要重启的计算机总数" json:"InstalledPendingReboot"`
	Failed                 int       `gorm:"column:Failed;default:null;comment:安装失败的计算机总数" json:"Failed"`
	Installed              int       `gorm:"column:Installed;default:null;comment:已安装的更新计算机总数" json:"Installed"`
	NotInstalled           int       `gorm:"column:NotInstalled;default:null;comment:需要安装的计算机总数" json:"NotInstalled"`
	Unknown                int       `gorm:"column:Unknown;default:null;comment:状态未知的计算机总数" json:"Unknown"`
	StatsTime              time.Time `gorm:"autoCreateTime;column:StatsTime;comment:统计时间" json:"StatsTime"`
}

func (RevisionInstallStatistics) TableName() string {
	return "revision_install_statistics"
}

type ComputerRevisionInstallStats struct {
	/*
		计算机更新安装状态（一台计算机对应一个更新的关系 computer * revision）
	*/
	ID uint `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	// 更新
	RevisionID     int      `gorm:"column:RevisionID" json:"RevisionID"`
	Revision       Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	UpdateID       string   `gorm:"column:UpdateID;type:char(36);index;not null;comment:该修订版本对应的更新的GUID" json:"UpdateID"`
	RevisionNumber int      `gorm:"column:RevisionNumber;not null;comment:单一更新的修订版本号，用来标识同一更新的不同修订版本" json:"RevisionNumber"`
	Title          string   `gorm:"column:Title;type:varchar(256);default:null;comment:更新名称" json:"Title"`
	// 计算机
	TargetID               int            `gorm:"column:TargetID" json:"TargetID"`
	ComputerTarget         ComputerTarget `gorm:"ForeignKey:TargetID;AssociationForeignKey:TargetID;not null"`
	FullDomainName         string         `gorm:"column:FullDomainName;default:null;type:varchar(256);comment:计算机名称" json:"FullDomainName"`
	LastSyncTime           time.Time      `gorm:"column:LastSyncTime;default:null;comment:上次同步时间" json:"LastSyncTime"`
	LastReportedStatusTime time.Time      `gorm:"column:LastReportedStatusTime;default:null;comment:上次报告状态时间" json:"LastReportedStatusTime"`
	Status                 int            `gorm:"column:Status;default:null;comment:安装状态" json:"Status"`
	StatsTime              time.Time      `gorm:"autoCreateTime;column:StatsTime;comment:统计时间" json:"StatsTime"`
}

func (ComputerRevisionInstallStats) TableName() string {
	return "computer_revision_stats"
}

type CustomReportRules struct {
	/*
		自定义报表规则
	*/
	TimeModelMiXin
	ID            uint   `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	Name          string `gorm:"column:name;type:varchar(256)" json:"name"`
	UserID        int    `gorm:"column:user_id;default:null" json:"user_id"`
	ServiceType   int    `gorm:"column:service_type;comment:服务类型" json:"service_type"`
	Show          string `gorm:"column:show;type:varchar(2000);comment:展示字段" json:"show"`
	Screen        string `gorm:"column:screen;default:null;type:varchar(2000);comment:筛选字段" json:"screen"`
	BuiltIn       *bool  `gorm:"column:built_in;default:true;comment:判断是否是默认规则" json:"built_in"`
	Desc          string `gorm:"column:desc;type:varchar(256);default:null;comment:规则描述" json:"desc"`
	TargetGroupID string `gorm:"column:target_group_id;default:null;type:char(36);comment:用于筛选的计算机组" json:"target_group_id"`
	ActionID      int    `gorm:"column:action_id;default:null;comment:用于筛选的审批状态" json:"action_id"`
	IsValid       bool   `gorm:"column:is_valid;default:true;comment:当前规则是否有效" json:"is_valid"`
}

func (CustomReportRules) TableName() string {
	return "custom_report_rules"
}

type ComputerLiveStats struct {
	/*
		计算机活跃量
	*/
	TimeModelMiXin
	ID            uint      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	Date          time.Time `gorm:"column:date;comment:当前统计日期" json:"date"`
	StatsTime     time.Time `gorm:"autoCreateTime;column:stats_time;comment:统计时间" json:"stats_time"`
	ComputerCount int       `gorm:"column:computer_count;default:null;comment:计算机活跃数量" json:"computer_count"`
	Cycle         int       `gorm:"column:cycle;default:0;comment:统计周期" json:"cycle"`
	Year          int       `gorm:"column:year;default:null;comment:统计时的年份" json:"year"`
	Month         int       `gorm:"column:month;default:null;comment:统计时的月份" json:"month"`
	Week          int       `gorm:"column:week;default:null;comment:统计时的周数" json:"week"`
	OSversion     string    `gorm:"column:os_version;default:null;type:varchar(256);comment:计算机操作系统版本" json:"os_version"`
}

func (ComputerLiveStats) TableName() string {
	return "live_stats"
}

type ComputerUpdateRelation struct {
	TimeModelMiXin
	ID            uint   `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	RuleID        uint   `gorm:"column:rule_id;comment:对应报表ID" json:"rule_id"`
	RevisionID    int    `gorm:"column:revision_id;comment:更新ID" json:"revision_id"`
	TargetGroupID string `gorm:"column:group_id;type:char(36);comment:计算机组ID" json:"group_id"`
	IsValid       bool   `gorm:"column:is_valid;default:true;comment:当前关系是否有效" json:"is_valid"`
}

func (ComputerUpdateRelation) TableName() string {
	return "computer_update_relation"
}
