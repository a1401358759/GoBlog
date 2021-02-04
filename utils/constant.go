package utils

const (
	// config
	ConfigEnv  = "G_CONFIG"
	ConfigFile = "config.ini"
	// computer group id
	UUIDAllComputer     = "A0A08746-4DBE-4A37-9ADF-9E7652C0B421" // 所有计算机组
	UUIDGroupUnassigned = "B73CA6ED-5727-47F3-84DE-015E03F6A88A" // 未分配组
	UUIDGroupDss        = "D374F42A-9BE2-4163-A0FA-3C86A401B7A7" // 下游服务器组
	NameAllComputer     = "All Computers"
	NameGroupUnassigned = "Unassigned Computers"
	NameGroupDss        = "Downstream Servers"
	TargetGroup         = "0"
	PlugInID            = "SimpleTargeting"
	// omega
	RevisionNumPerResp = 50
	SpecialUpdateID    = "00000000-0000-0000-0000-000000000000" // event is not associated with any particular update
	// version
	CuspProtocolVersion          = "3.2"
	ProtocolComponentNo          = 2 // 协议版本格式 x.y 2表示点分割后的字段个数
	UssAuthCookieComponentNo     = 3
	ProtocolMajorOffset          = 0
	ProtocolMajor                = "1"
	UssCookieComponentNo         = 6
	UssCookieExpirationOffset    = 3
	UssCookieProtocolOffset      = 4
	ProtocolVersion              = "1.23" // 上下游同步协议
	ReplicaDeploymentsPerRequest = 1000   // 副本模式每次返回给下游的deployment的数量
	// xml namespace soap
	SOAP         = "http://schemas.xmlsoap.org/soap/envelope/"
	XSI          = "http://www.w3.org/2001/XMLSchema-instance"
	XSD          = "http://www.w3.org/2001/XMLSchema"
	Xmlns        = "http://www.microsoft.com/SoftwareDistribution/Server/ClientWebService"
	AuthXmlns    = "http://www.microsoft.com/SoftwareDistribution/Server/SimpleAuthWebService"
	DssAuthXmlns = "http://www.microsoft.com/SoftwareDistribution/Server/DssAuthWebService"
	// root dns
	DefaultRootDns = "http://www.microsoft.com/SoftwareDistribution"

	// time format
	TimeFormat      = "2006-01-02T15:04:05.999999Z07:00"
	OnlyTimeFormat  = "15:04:05"
	ExpirationDelta = 365 // cookie的有效时长 ：365天
	// actionID
	InstallAction            = 0 // 审批到安装
	UninstallAction          = 1 // 审批到卸载
	PreDeploymentCheckAction = 2 // 未审批
	BlockAction              = 3 // 未审批的一种
	EvaluateAction           = 4 // 不提供更新、不报告状态
	BundleAction             = 5 // 更新不提供安装，它只是部署，因为它是捆绑其他显式部署的更新
	DssAction                = 7 // 下有服务器
	DeclineAction            = 8 // 拒绝的更新
	// patchingType
	Unspecified       = "0"
	SelfContained     = "1"
	Express           = "2"
	MSPBinaryDelta    = "3"
	Setup360Installer = "5"
	Setup360WIM       = "6"
	Setup360Servicing = "7"
	// us source
	SourceCUSID  = 1
	SourceWSUSID = 2
	SourceCUS    = "CUS"
	SourceWSUS   = "WSUS"
)

var PublicationState = struct {
	Published int
	Expired   int
}{
	Published: 0,
	Expired:   1,
}

var (
	WSUSBlackList          = [1]string{"A3C2375D-0C8A-42F9-BCE0-28333E198407"}
	NeedLanguage           = [2]string{"zh-cn", "en"}
	RevisionClause         = make(map[int][][]int, 0)
	BackRevisionClause     = make(map[int][]int, 0)
	NonePrerequisiteClause = make([]int, 0)
)

var ApprovalRuleStatus = struct {
	SUCCESS, FAILED, INVALID int
}{
	SUCCESS: 0, // 成功
	FAILED:  1, // 失败
	INVALID: 2, // 规则失效
}
