package utils

const (
	// config
	ConfigEnv  = "G_CONFIG"
	ConfigFile = "config.ini"
	// time format
	TimeFormat     = "2006-01-02 15:04:05"
	OnlyTimeFormat = "15:04:05"
	// Domain
	Domain = "https://yangsihan.com"
)

var BlogStatus = struct {
	DRAFT, PUBLISHED int
}{
	DRAFT:     1, // 草稿
	PUBLISHED: 2, // 已发布
}

var CarouselImgType = struct {
	BANNER, ADS int
}{
	BANNER: 1, // banner
	ADS:    2, // ads
}

var EditorKind = struct {
	RichText, Markdown int
}{
	RichText: 1, // 富文本编辑器
	Markdown: 2, // Markdown编辑器
}
