package model

import "time"

// Author 文章作者
type Author struct {
	ID      int    `gorm:"primary_key;column:id" json:"id"`
	Name    string `gorm:"type:varchar(256);column:name;comment:姓名" json:"name"`
	Email   string `gorm:"type:varchar(128);column:email;comment:邮件;default:null" json:"email"`
	Website string `gorm:"type:varchar(128);column:website;comment:个人网站;default:null" json:"website"`
	TimeModelMiXin
}

func (Author) TableName() string {
	return "author"
}

// OwnerMessage 主人寄语
type OwnerMessage struct {
	ID         int       `gorm:"primary_key;column:id" json:"id"`
	Summary    string    `gorm:"type:varchar(100);column:summary;comment:简介;default:null" json:"summary"`
	Message    string    `gorm:"type:longtext;column:message;comment:邮件;default:null" json:"message"`
	Editor     int       `gorm:"column:editor;comment:编辑器类型;default:1" json:"editor"`
	CreatedAt  time.Time `gorm:"autoCreateTime;column:created_at;comment:创建时间" json:"created_at"`
	LastUpdate time.Time `gorm:"autoUpdateTime;column:last_update;comment:最后修改时间" json:"last_update"`
}

func (OwnerMessage) TableName() string {
	return "owner_message"
}

// Tag 标签
type Tag struct {
	ID   int    `gorm:"primary_key;column:id" json:"id"`
	Name string `gorm:"type:varchar(20);column:name;comment:标签名" json:"name"`
	TimeModelMiXin
}

func (Tag) TableName() string {
	return "tag"
}

// Classification 分类
type Classification struct {
	ID   int    `gorm:"primary_key;column:id" json:"id"`
	Name string `gorm:"type:varchar(25);column:name" json:"name"`
	TimeModelMiXin
}

func (Classification) TableName() string {
	return "classification"
}

// Article 文章
type Article struct {
	ID               int            `gorm:"primary_key;column:id" json:"id"`
	Title            string         `gorm:"type:varchar(100);column:title;comment:标题" json:"title"`
	AuthorID         int            `gorm:"primary_key;column:author_id" json:"author_id"`
	Author           Author         `gorm:"ForeignKey:author_id;AssociationForeignKey:author_id;not null"`
	Tags             []Tag          `gorm:"many2many:article_tags"`
	ClassificationID int            `gorm:"primary_key;column:classification_id" json:"classification_id"`
	Classification   Classification `gorm:"ForeignKey:classification_id;AssociationForeignKey:classification_id;not null"`
	PublishTime      time.Time      `gorm:"autoCreateTime;column:publish_time;comment:发表时间" json:"publish_time"`
	LastUpdate       time.Time      `gorm:"autoUpdateTime;column:last_update;comment:最后修改时间" json:"last_update"`
	Count            int            `gorm:"column:count;default:0;comment:文章点击数" json:"count"`
	Editor           int            `gorm:"column:editor;comment:编辑器类型;default:1" json:"editor"`
	Status           int            `gorm:"column:status;comment:状态;default:2" json:"status"`
}

func (Article) TableName() string {
	return "article"
}

// Links 友情链接
type Links struct {
	ID      int    `gorm:"primary_key;column:id" json:"id"`
	Name    string `gorm:"type:varchar(50);column:name;comment:网站名称" json:"name"`
	Link    string `gorm:"type:varchar(100);column:link;comment:网站地址" json:"link"`
	Avatar  string `gorm:"type:varchar(100);column:avatar;comment:网站图标;default:null" json:"avatar"`
	Desc    string `gorm:"type:varchar(200);column:desc;comment:网站描述;default:null" json:"desc"`
	Weights int    `gorm:"column:weights;comment:权重;default:10" json:"weights"`
	Email   string `gorm:"type:varchar(128);column:email;comment:邮件;default:null" json:"email"`
	TimeModelMiXin
}

func (Links) TableName() string {
	return "links"
}

// CarouselImg 轮播图管理
type CarouselImg struct {
	ID          int    `gorm:"primary_key;column:id" json:"id"`
	Name        string `gorm:"type:varchar(50);column:name;comment:图片名称" json:"name"`
	Description string `gorm:"type:varchar(100);column:description;comment:图片描述" json:"description"`
	Path        string `gorm:"type:varchar(100);column:path;comment:图片地址" json:"path"`
	Link        string `gorm:"type:varchar(200);column:link;comment:图片外链;default:null" json:"link"`
	Weights     int    `gorm:"column:weights;comment:图片权重;default:10" json:"weights"`
	ImgType     int    `gorm:"column:img_type;comment:类型;default:1" json:"img_type"`
	TimeModelMiXin
}

func (CarouselImg) TableName() string {
	return "carousel_img"
}

// Music 背景音乐
type Music struct {
	ID     int    `gorm:"primary_key;column:id" json:"id"`
	Name   string `gorm:"type:varchar(50);column:name;comment:音乐名称" json:"name"`
	Url    string `gorm:"type:varchar(100);column:url;comment:音乐地址" json:"url"`
	Cover  string `gorm:"type:varchar(100);column:cover;comment:音乐封面" json:"cover"`
	Artist string `gorm:"type:varchar(100);column:artist;comment:艺术家;default:null" json:"artist"`
	Lrc    string `gorm:"type:varchar(100);column:lrc;comment:音乐歌词;default:null" json:"lrc"`
	TimeModelMiXin
}

func (Music) TableName() string {
	return "music"
}

// Subscription 邮件订阅
type Subscription struct {
	ID    int    `gorm:"primary_key;column:id" json:"id"`
	Email string `gorm:"type:varchar(128);column:email;comment:订阅邮箱" json:"email"`
	TimeModelMiXin
}

func (Subscription) TableName() string {
	return "subscription"
}

// Visitor 访客表
type Visitor struct {
	ID       int    `gorm:"primary_key;column:id" json:"id"`
	NickName string `gorm:"type:varchar(50);column:nickname" json:"nickname"`
	Avatar   string `gorm:"type:varchar(100);column:avatar" json:"avatar"`
	Email    string `gorm:"type:varchar(128);column:email;comment:邮件;default:null" json:"email"`
	Website  string `gorm:"type:varchar(128);column:website;comment:个人网站;default:null" json:"website"`
	Blogger  bool   `gorm:"column:blogger;default:false" json:"blogger"`
	TimeModelMiXin
}

func (Visitor) TableName() string {
	return "visitor"
}

// Comments 评论表
type Comments struct {
	ID        int       `gorm:"primary_key;column:id" json:"id"`
	UserID    int       `gorm:"primary_key;column:user_id" json:"user_id"`
	Visitor   Visitor   `gorm:"ForeignKey:user_id;AssociationForeignKey:user_id;not null"`
	ReplyToID int       `gorm:"primary_key;column:reply_to_id;null" json:"reply_to_id"`
	ReplyTo   Visitor   `gorm:"ForeignKey:reply_to_id;AssociationForeignKey:reply_to_id;null"`
	Content   string    `gorm:"type:longtext;column:content" json:"content"`
	ParentID  int       `gorm:"primary_key;column:parent_id" json:"parent_id"`
	Parent    *Comments `gorm:"ForeignKey:parent_id;AssociationForeignKey:parent_id;null"`
	Target    string    `gorm:"type:varchar(100);column:target;default:null" json:"target"`
	Anchor    string    `gorm:"type:varchar(20);column:anchor;default:null" json:"anchor"`
	IPAddress string    `gorm:"type:varchar(20);column:ip_address;default:null" json:"ip_address"`
	Country   string    `gorm:"type:varchar(20);column:country;default:null" json:"country"`
	Province  string    `gorm:"type:varchar(30);column:province;default:null" json:"province"`
	City      string    `gorm:"type:varchar(30);column:city;default:null" json:"city"`
	TimeModelMiXin
}

func (Comments) TableName() string {
	return "comment"
}
