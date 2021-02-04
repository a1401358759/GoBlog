package model

import "time"

type ServerConfiguration struct {
	VersionCheckMixin
	ID                                  int       `gorm:"primary_key;column:id" json:"ID"`
	ServerID                            string    `gorm:"type:char(36);column:ServerID;unique;comment:服务器唯一标识UUID" json:"ServerID"`
	FullDomainName                      string    `gorm:"type:varchar(256);column:FullDomainName;not null;comment:服务器的机器名称" json:"FullDomainName"`
	ServerVersion                       string    `gorm:"type:varchar(256);column:ServerVersion;default:null" json:"ServerVersion"`
	RollupResetGuid                     string    `gorm:"type:char(36);column:RollupResetGuid;default:null" json:"RollupResetGuid"`
	ReplicationMode                     string    `gorm:"type:varchar(256);column:ReplicationMode;default:null;default:'Auto';comment:运行模式为自治模式(Auto)或者副本模式(Replica)" json:"ReplicationMode"`
	CatalogOnlySync                     bool      `gorm:"default:null;default:0;column:CatalogOnlySync" json:"CatalogOnlySync"`
	LazySync                            bool      `gorm:"default:null;column:LazySync" json:"LazySync"`
	ServerHostsPsfFiles                 bool      `gorm:"default:null;column:ServerHostsPsfFiles" json:"ServerHostsPsfFiles"`
	MaxNumberOfUpdatesPerRequest        int       `gorm:"default:null;column:MaxNumberOfUpdatesPerRequest;default:100" json:"MaxNumberOfUpdatesPerRequest"`
	DoDetailedRollup                    bool      `gorm:"default:null;column:DoDetailedRollup;default:0;comment:是否要收集下游服务器的计算机详细信息" json:"DoDetailedRollup"`
	ConfigAnchor                        time.Time `gorm:"default:null;column:ConfigAnchor" json:"ConfigAnchor"`
	LastRollupTime                      time.Time `gorm:"default:null;column:LastRollupTime" json:"LastRollupTime"`
	ProtocolVersion                     string    `gorm:"type:varchar(256);column:ProtocolVersion;default:null" json:"ProtocolVersion"`
	RollupDownstreamServersMaxBatchSize int       `gorm:"default:null;column:RollupDownstreamServersMaxBatchSize;default:100" json:"RollupDownstreamServersMaxBatchSize"`
	RollupComputersMaxBatchSize         int       `gorm:"default:null;column:RollupComputersMaxBatchSize;default:100" json:"RollupComputersMaxBatchSize"`
	GetOutOfSyncComputersMaxBatchSize   int       `gorm:"default:null;column:GetOutOfSyncComputersMaxBatchSize;default:100" json:"GetOutOfSyncComputersMaxBatchSize"`
	RollupComputerStatusMaxBatchSize    int       `gorm:"default:null;column:RollupComputerStatusMaxBatchSize;default:100" json:"RollupComputerStatusMaxBatchSize"`
	SyncMode                            int       `gorm:"default:null;column:SyncMode;default:0;comment:与上游服务器的同步方式:0-手动 1-定时" json:"SyncMode"`
	SyncFirstTime                       string    `gorm:"default:null;column:SyncFirstTime;comment:指定自动同步模式下的第一次开始时间" json:"SyncFirstTime"`
	SyncTimesPerDay                     int       `gorm:"default:1;column:SyncTimesPerDay;comment:自动同步模式下每天的同步次数" json:"SyncTimesPerDay"`
	AutoApproval                        bool      `gorm:"default:null;column:AutoApproval;default:0;comment:是否自动审批为Approval" json:"AutoApproval"`
	EnableExpress                       bool      `gorm:"default:null;column:EnableExpress;default:0;comment:是否支持Express模式" json:"EnableExpress"`
	IsDownloadApproved                  bool      `gorm:"default:null;column:IsDownloadApproved;default:1;comment:是否仅下载已审批的更新文件" json:"IsDownloadApproved"`
	IsAutoDeclineExpired                bool      `gorm:"default:null;column:IsAutoDeclineExpired;default:1;comment:是否自动拒绝已过期的更新" json:"IsAutoDeclineExpired"`
	IsAutoApproveRevision               bool      `gorm:"default:null;column:IsAutoApproveRevision;default:1;comment:是否自动审批已审批的新修订" json:"IsAutoApproveRevision"`
}

func (ServerConfiguration) TableName() string {
	return "server_configuration"
}

type ParentUssConfig struct {
	VersionCheckMixin
	ID                                  int       `gorm:"primary_key;column:id" json:"id"`
	AuthCookie                          string    `gorm:"type:varchar(1024);column:authCookie;default:null" json:"authCookie"`
	Expiration                          string    `gorm:"type:varchar(128);column:expiration;default:null;comment:USS服务器返回的Cookie过期时间" json:"expiration"`
	EncryptedCookie                     string    `gorm:"type:varchar(1024);column:encryptedCookie;default:null;comment:USS服务器返回的Cookie" json:"EncryptedCookie"`
	ServerID                            string    `gorm:"type:char(36);column:ServerID;default:null" json:"ServerID"`
	ServerIP                            string    `gorm:"type:varchar(256);column:ServerIP;default:null" json:"ServerIP"`
	ServerPort                          string    `gorm:"type:varchar(256);column:ServerPort;default:null" json:"ServerPort"`
	ServerUseTls                        bool      `gorm:"default:null;column:ServerUseTls;default:0" json:"ServerUseTls"`
	LastConfigAnchor                    string    `gorm:"type:varchar(256);column:LastConfigAnchor;default:null" json:"LastConfigAnchor"`
	LastConfigSyncanchor                string    `gorm:"type:varchar(256);column:LastConfigSyncanchor;default:null;comment:用于记录getConfig为True时USS返回的Anchor" json:"LastConfigSyncanchor"`
	LastSyncanchor                      string    `gorm:"type:varchar(256);column:LastSyncanchor;default:null;comment:用于记录getConfig为False时USS返回的Anchor" json:"LastSyncanchor"`
	LastSyncCategory                    string    `gorm:"type:varchar(5000);column:LastSyncCategory;default:null;comment:用于记录最后一次成功同步的category内容" json:"LastSyncCategory"`
	LastDeploymentanchor                string    `gorm:"type:varchar(256);column:LastDeploymentanchor;default:null" json:"LastDeploymentanchor"`
	MaxNumberOfUpdatesPerRequest        int       `gorm:"default:null;column:MaxNumberOfUpdatesPerRequest" json:"MaxNumberOfUpdatesPerRequest"`
	CatalogOnlySync                     bool      `gorm:"default:null;column:CatalogOnlySync" json:"CatalogOnlySync"`
	LazySync                            bool      `gorm:"default:null;column:LazySync" json:"LazySync"`
	ServerHostsPsfFiles                 bool      `gorm:"default:null;column:ServerHostsPsfFiles" json:"ServerHostsPsfFiles"`
	UpdateTime                          time.Time `gorm:"autoCreateTime;column:UpdateTime" json:"UpdateTime"`
	DoDetailedRollup                    bool      `gorm:"default:null;column:DoDetailedRollup" json:"DoDetailedRollup"`
	RollupResetGuid                     string    `gorm:"type:char(36);column:RollupResetGuid;default:null" json:"RollupResetGuid"`
	ProtocolVersion                     string    `gorm:"type:varchar(256);column:ProtocolVersion;default:null" json:"ProtocolVersion"`
	RollupDownstreamServersMaxBatchSize int       `gorm:"default:null;column:RollupDownstreamServersMaxBatchSize" json:"RollupDownstreamServersMaxBatchSize"`
	RollupComputersMaxBatchSize         int       `gorm:"default:null;column:RollupComputersMaxBatchSize" json:"RollupComputersMaxBatchSize"`
	GetOutOfSyncComputersMaxBatchSize   int       `gorm:"default:null;column:GetOutOfSyncComputersMaxBatchSize" json:"GetOutOfSyncComputersMaxBatchSize"`
	RollupComputerStatusMaxBatchSize    int       `gorm:"default:null;column:RollupComputerStatusMaxBatchSize" json:"RollupComputerStatusMaxBatchSize"`
	ProxyHost                           string    `gorm:"type:varchar(128);column:proxy_host;default:null" json:"proxy_host"`
	ProxyPort                           string    `gorm:"type:varchar(128);column:proxy_port;default:null" json:"proxy_port"`
	ProxyUserName                       string    `gorm:"type:varchar(128);column:proxy_user_name;default:null" json:"proxy_user_name"`
	ProxyPassword                       string    `gorm:"type:varchar(128);column:proxy_password;default:null" json:"proxy_password"`
	ProxyDomain                         string    `gorm:"type:varchar(128);column:proxy_domain;default:null" json:"proxy_domain"`
	ProxyType                           string    `gorm:"type:varchar(128);column:proxy_type;default:null" json:"proxy_type"`
	ProxyIsUsed                         bool      `gorm:"default:0;column:proxy_is_used" json:"proxy_is_used"`
}

func (ParentUssConfig) TableName() string {
	return "parent_uss_config"
}
