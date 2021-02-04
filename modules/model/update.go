package model

import (
	"time"
)

type Update struct {
	VersionCheckMixin
	LocalUpdateID          int       `gorm:"primary_key;AUTO_INCREMENT;column:LocalUpdateID;comment:该update在数据库内的唯一标识" json:"LocalUpdateID"`
	UpdateID               string    `gorm:"type:char(36);not null;column:UpdateID;comment:该update的GUID" json:"UpdateID"`
	UpdateType             string    `gorm:"type:varchar(256);not null;column:UpdateType;comment:该更新的类型" json:"UpdateType"`
	IsClientSelfUpdate     bool      `gorm:"type:boolean;default:null;column:IsClientSelfUpdate" json:"IsClientSelfUpdate"`
	PublisherID            string    `gorm:"type:varchar(36);default:null;column:PublisherID" json:"PublisherID"`
	IsPublic               bool      `gorm:"type:boolean;default:null;column:IsPublic" json:"IsPublic"`
	IsHidden               bool      `gorm:"type:boolean;default:null;column:IsHidden" json:"IsHidden"`
	DetectoidType          string    `gorm:"type:varchar(80);default:null;column:DetectoidType" json:"DetectoidType"`
	LegacyName             string    `gorm:"type:varchar(255);default:null;column:LegacyName" json:"LegacyName"`
	LastUndeclinedTime     time.Time `gorm:"default:null;column:LastUndeclinedTime" json:"LastUndeclinedTime"`
	IsLocallyPublished     bool      `gorm:"default:null;column:IsLocallyPublished;comment:该更新是否本地发布" json:"IsLocallyPublished"`
	ImportedTime           time.Time `gorm:"autoCreateTime;column:ImportedTime" json:"ImportedTime"`
	IsCategory             bool      `gorm:"default:null;column:IsCategory;comment:该更新的更新类型是否为category" json:"IsCategory"`
	CategoryIndex          int       `gorm:"type:int;default:null;column:CategoryIndex" json:"CategoryIndex"`
	CategoryID             int       `gorm:"type:int;default:null;column:CategoryID" json:"CategoryID"`
	ParentCategoryID       int       `gorm:"type:int;default:null;column:ParentCategoryID" json:"ParentCategoryID"`
	CategoryType           string    `gorm:"type:varchar(256);default:null;column:CategoryType;comment:当更新的类型为category，那么category的子类型是什么" json:"CategoryType"`
	CategoryTypeLevel      int       `gorm:"type:int;default:null;column:CategoryTypeLevel" json:"CategoryTypeLevel"`
	LastChange             time.Time `gorm:"default:null;column:LastChange" json:"LastChange"`
	ProhibitsSubcategories bool      `gorm:"type:boolean;default:null;column:ProhibitsSubcategories" json:"ProhibitsSubcategories"`
	ProhibitsUpdates       bool      `gorm:"type:boolean;default:null;column:ProhibitsUpdates" json:"ProhibitsUpdates"`
	DisplayOrder           int       `gorm:"type:int;default:null;column:DisplayOrder" json:"DisplayOrder"`
}

func (Update) TableName() string {
	return "update"
}

type Revision struct {
	VersionCheckMixin
	RevisionID                    int       `gorm:"primary_key;AUTO_INCREMENT;column:RevisionID;comment:更新的修订版本在数据库内的唯一标识" json:"RevisionID"`
	LocalUpdateID                 int       `gorm:"index;not null;column:LocalUpdateID;comment:该更新的本地标识" json:"LocalUpdateID"`
	Update                        Update    `gorm:"ForeignKey:LocalUpdateID;AssociationForeignKey:LocalUpdateID;not null;"`
	RevisionNumber                int       `gorm:"index;not null;column:RevisionNumber;comment:单一更新的修订版本号，用来标识同一更新的不同修订版本" json:"RevisionNumber"`
	UpdateID                      string    `gorm:"type:char(36);not null;index;column:UpdateID;comment:该修订版本对应的更新的GUID" json:"UpdateID"`
	UpdateType                    string    `gorm:"type:varchar(256);default:null;column:UpdateType;comment:该修订版本对应更新的更新类型" json:"UpdateType"`
	CategoryType                  string    `gorm:"type:varchar(256);default:null;column:CategoryType;comment:该修订版本的category类型" json:"CategoryType"`
	ImportedTime                  time.Time `gorm:"autoCreateTime;index;column:ImportedTime" json:"ImportedTime"`
	CheckedForSyncFromUss         bool      `gorm:"default 0;column:CheckedForSyncFromUss" json:"CheckedForSyncFromUss"`
	LastChangedAnchor             time.Time `gorm:"autoCreateTime;column:LastChangedAnchor" json:"LastChangedAnchor"`
	LastIsLeafChange              time.Time `gorm:"default:null;column:LastIsLeafChange" json:"LastIsLeafChange"`
	IsLeaf                        bool      `gorm:"default:null;default:1;column:IsLeaf" json:"IsLeaf"`
	IsBeta                        bool      `gorm:"default:null;default:0;column:IsBeta" json:"IsBeta"`
	IsApproved                    bool      `gorm:"default:null;default:0;column:IsApproved" json:"IsApproved"`
	HasBundle                     bool      `gorm:"default:null;default:0;column:HasBundle;comment:表示该revision是否包含bundle update" json:"HasBundle"`
	TimeToGoLiveOnCatalog         time.Time `gorm:"default:null;column:TimeToGoLiveOnCatalog" json:"TimeToGoLiveOnCatalog"`
	CheckedInFrontend             bool      `gorm:"default:null;default:1;column:CheckedInFrontend;comment:表示该产品是否在前端被勾选" json:"CheckedInFrontend"`
	CheckedInClassification       bool      `gorm:"default:null;default:1;column:CheckedInClassification" json:"CheckedInClassification"`
	ProductRevisionID             int       `gorm:"default:null;index;column:ProductRevisionID;comment:如果是software，记录属于的产品的RevisionID" json:"ProductRevisionID"`
	ClassificationRevisionID      int       `gorm:"default:null;index;column:ClassificationRevisionID;comment:如果是software，记录属于的分类的RevisionID" json:"ClassificationRevisionID"`
	ImportCompleted               bool      `gorm:"default:null;default:0;column:import_completed;comment:这里表示该revision是否完整的import" json:"ImportCompleted"`
	RowID                         string    `gorm:"type:char(36);default:null;column:RowID" json:"RowID"`
	State                         int       `gorm:"type:int;default:null;column:State" json:"State"`
	Origin                        int       `gorm:"type:int;default:null;column:Origin" json:"Origin"`
	IsCritical                    bool      `gorm:"default:null;column:IsCritical" json:"IsCritical"`
	LanguageMask                  int64     `gorm:"type:bigint;default:null;column:LanguageMask" json:"LanguageMask"`
	IsLatestRevision              bool      `gorm:"default:null;column:IsLatestRevision" json:"IsLatestRevision"`
	IsMandatory                   bool      `gorm:"default:null;column:IsMandatory" json:"IsMandatory"`
	PublicationState              int       `gorm:"default:null;column:PublicationState" json:"PublicationState"`
	CreationDate                  time.Time `gorm:"default:null;column:CreationDate" json:"CreationDate"`
	ReceivedFromCreatorService    time.Time `gorm:"default:null;column:ReceivedFromCreatorService" json:"ReceivedFromCreatorService"`
	ExplicitlyDeployable          bool      `gorm:"default:null;column:ExplicitlyDeployable" json:"ExplicitlyDeployable"`
	CanInstall                    bool      `gorm:"default:null;column:CanInstall" json:"CanInstall"`
	InstallationImpact            string    `gorm:"type:varchar(256);default:null;column:InstallationImpact" json:"InstallationImpact"`
	InstallRequiresConnectivity   bool      `gorm:"default:null;column:InstallRequiresConnectivity" json:"InstallRequiresConnectivity"`
	InstallRequiresUserInput      bool      `gorm:"default:null;column:InstallRequiresUserInput" json:"InstallRequiresUserInput"`
	InstallRebootBehavior         string    `gorm:"type:varchar(256);default:null;column:InstallRebootBehavior" json:"InstallRebootBehavior"`
	CanUninstall                  bool      `gorm:"default:null;column:CanUninstall" json:"CanUninstall"`
	UninstallImpact               string    `gorm:"type:varchar(256);default:null;column:UninstallImpact" json:"UninstallImpact"`
	UninstallRequiresConnectivity bool      `gorm:"default:null;column:UninstallRequiresConnectivity" json:"UninstallRequiresConnectivity"`
	UninstallRequiresUserInput    bool      `gorm:"default:null;column:UninstallRequiresUserInput" json:"UninstallRequiresUserInput"`
	UninstallRebootBehavior       string    `gorm:"type:varchar(256);default:null;column:UninstallRebootBehavior" json:"UninstallRebootBehavior"`
	HandlerID                     int       `gorm:"default:null;column:HandlerID" json:"HandlerID"`
	EulaID                        string    `gorm:"type:char(36);default:null;column:EulaID" json:"EulaID"`
	RequiresReacceptanceOfEula    bool      `gorm:"default:null;column:RequiresReacceptanceOfEula" json:"RequiresReacceptanceOfEula"`
	DefaultPropertiesLanguageID   int       `gorm:"default:null;column:DefaultPropertiesLanguageID" json:"DefaultPropertiesLanguageID"`
	EulaExplicitlyAccepted        bool      `gorm:"default:null;column:EulaExplicitlyAccepted" json:"EulaExplicitlyAccepted"`
	MsrcSeverity                  string    `gorm:"type:char(20);default:null;column:MsrcSeverity" json:"MsrcSeverity"`
	CompatibleProtocolVersion     string    `gorm:"type:char(20);default:null;column:CompatibleProtocolVersion" json:"CompatibleProtocolVersion"`
	OemOnlyDriver                 bool      `gorm:"default:null;column:OemOnlyDriver" json:"OemOnlyDriver"`
	DriverMinInstalledVersion     int64     `gorm:"type:bigint;default:null;column:DriverMinInstalledVersion" json:"DriverMinInstalledVersion"`
	DriverMaxInstalledVersion     int64     `gorm:"type:bigint;default:null;column:DriverMaxInstalledVersion" json:"DriverMaxInstalledVersion"`
	SecurityBulletinID            string    `gorm:"type:varchar(15);default:null;column:SecurityBulletinID" json:"SecurityBulletinID"`
	KBArticleID                   string    `gorm:"type:varchar(15);default:null;column:KBArticleID" json:"KBArticleID"`
	SyncSource                    string    `gorm:"type:enum('CUS', 'WSUS', 'FileImport', 'Unknown');column:SyncSource;default:'CUS'"`
	SupportUrl                    string    `gorm:"type:varchar(256);default:null;column:SupportUrl" json:"SupportUrl"`
	MoreInfoUrl                   string    `gorm:"type:varchar(256);default:null;column:MoreInfoUrl" json:"MoreInfoUrl"`
}

func (Revision) TableName() string {
	return "revision"
}

type RevisionInCategory struct {
	VersionCheckMixin
	ID             int      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	RevisionID     int      `gorm:"not null;column:RevisionID" json:"RevisionID"`
	Revision       Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	UpdateID       string   `gorm:"not null;type:char(36);column:UpdateID" json:"UpdateID"`
	RevisionNumber int      `gorm:"not null;column:RevisionNumber" json:"RevisionNumber"`
	Type           string   `gorm:"type:varchar(256);default:null;column:Type" json:"Type"`
	Value          string   `gorm:"type:char(36);default:null;column:Value" json:"Value"`
}

func (RevisionInCategory) TableName() string {
	return "revision_in_category"
}

type RevisionPrerequisite struct {
	VersionCheckMixin
	PrerequisiteID int      `gorm:"primary_key;column:PrerequisiteID;AUTO_INCREMENT" json:"PrerequisiteID"`
	RevisionID     int      `gorm:"not null;column:RevisionID;index" json:"RevisionID"`
	Revision       Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
}

func (RevisionPrerequisite) TableName() string {
	return "revision_prerequisite"
}

type UpdateForPrerequisite struct {
	VersionCheckMixin
	ID                   int                  `gorm:"primary_key;column:id" json:"id"`
	PrerequisiteID       int                  `gorm:"column:PrerequisiteID;index" json:"PrerequisiteID"`
	RevisionPrerequisite RevisionPrerequisite `gorm:"ForeignKey:PrerequisiteID;AssociationForeignKey:PrerequisiteID;not null"`
	LocalUpdateID        int                  `gorm:"column:LocalUpdateID;index" json:"LocalUpdateID"`
	Update               Update               `gorm:"ForeignKey:LocalUpdateID;AssociationForeignKey:LocalUpdateID;not null"`
}

func (UpdateForPrerequisite) TableName() string {
	return "update_for_prerequisite"
}

type Rules struct {
	VersionCheckMixin
	XmlID                    int      `gorm:"primary_key;AUTO_INCREMENT;column:XmlID" json:"XmlID"`
	RevisionID               int      `gorm:"not null;column:RevisionID" json:"RevisionID"`
	Revision                 Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:Revision;not null"`
	RootElementXml           string   `gorm:"type:longtext;default:null;column:RootElementXml" json:"RootElementXml"`
	RootElementType          int      `gorm:"default:null;column:RootElementType" json:"RootElementType"`
	RootElementXmlCompressed []byte   `gorm:"type:longblob;default:null;column:RootElementXmlCompressed" json:"RootElementXmlCompressed"`
}

func (Rules) TableName() string {
	return "rules"
}

type DownloadFiles struct {
	VersionCheckMixin
	ID                    int       `gorm:"primary_key;column:id" json:"id"`
	FileDigest            string    `gorm:"type:varchar(256);not null;column:FileDigest" json:"FileDigest"`
	RevisionID            int       `gorm:"column:RevisionID;index" json:"RevisionID"`
	Revision              Revision  `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID"`
	RevisionNumber        int       `gorm:"index;not null;column:RevisionNumber" json:"RevisionNumber"`
	UpdateID              string    `gorm:"type:char(36);default:null;index;column:UpdateID" json:"UpdateID"`
	PatchingType          string    `gorm:"type:varchar(256);default:null;index;column:PatchingType" json:"PatchingType"`
	FileName              string    `gorm:"type:varchar(256);default:null;index;column:FileName" json:"FileName"`
	Modified              time.Time `gorm:"default:null;column:Modified" json:"Modified"`
	Size                  int64     `gorm:"default:null;column:Size" json:"Size"`
	IsEula                bool      `gorm:"default:null;column:IsEula" json:"IsEula"`
	Language              string    `gorm:"type:varchar(32);default:null;column:Language" json:"Language"`
	MUURL                 string    `gorm:"type:varchar(1024);default:null;column:MUURL" json:"MUURL"`
	USSURL                string    `gorm:"type:varchar(1024);default:null;column:USSURL" json:"USSURL"`
	IsExternalCab         bool      `gorm:"default:null;column:IsExternalCab" json:"IsExternalCab"`
	IsSecure              bool      `gorm:"default:null;column:IsSecure" json:"IsSecure"`
	IsEncrypted           bool      `gorm:"default:null;column:IsEncrypted" json:"IsEncrypted"`
	DecryptionKey         string    `gorm:"type:varchar(256);default:null;column:DecryptionKey" json:"DecryptionKey"`
	DecryptionFileDigest  string    `gorm:"type:varchar(256);default:null;column:DecryptionFileDigest" json:"DecryptionFileDigest"`
	TotalBytesForDownload int64     `gorm:"default:null;column:TotalBytesForDownload" json:"TotalBytesForDownload"`
	BytesDownloaded       int64     `gorm:"default:null;column:BytesDownloaded" json:"BytesDownloaded"`
	IsOnServer            bool      `gorm:"default:0;column:IsOnServer" json:"IsOnServer"`
	ConfigurationID       int       `gorm:"default:null;column:ConfigurationID" json:"ConfigurationID"`
	DesiredState          int       `gorm:"default:null;column:DesiredState" json:"DesiredState"`
	ActualState           int       `gorm:"default:null;column:ActualState" json:"ActualState"`
	TimeAdded             time.Time `gorm:"default:null;column:TimeAdded" json:"TimeAdded"`
	DownloadRequired      bool      `gorm:"default:null;column:DownloadRequired" json:"DownloadRequired"`
}

func (DownloadFiles) TableName() string {
	return "files"
}

type Bundle struct {
	VersionCheckMixin
	ID               int      `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
	RevisionID       int      `gorm:"not null;column:RevisionID;index" json:"RevisionID"`
	Revision         Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	BundleRevisionID int      `gorm:"not null;column:BundleRevisionID;index" json:"BundleRevisionID"`
}

func (Bundle) TableName() string {
	return "bundle"
}

type Superseded struct {
	VersionCheckMixin
	SupersededID int      `gorm:"primary_key;AUTO_INCREMENT;column:SupersededID" json:"SupersededID"`
	RevisionID   int      `gorm:"not null;column:RevisionID;index" json:"RevisionID"`
	Revision     Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	UpdateID     string   `gorm:"not null;type:char(36);column:UpdateID" json:"UpdateID"`
}

func (Superseded) TableName() string {
	return "superseded"
}
