package model

type Property struct {
	VersionCheckMixin
	Id             int      `gorm:"primary_key;column:id" json:"id"`
	RevisionID     int      `gorm:"null;column:RevisionID;index" json:"RevisionID"`
	Revision       Revision `gorm:"ForeignKey:RevisionID;AssociationForeignKey:RevisionID;not null"`
	Language       string   `gorm:"type:varchar(32);default:null;column:Language" json:"Language"`
	Title          string   `gorm:"type:varchar(200);default:null;column:Title;index" json:"Title"`
	Description    string   `gorm:"type:varchar(1500);default:null;column:Description" json:"Description"`
	MoreInfoUrl    string   `gorm:"type:varchar(256);default:null;column:more_info_url" json:"more_info_url"`
	SupportUrl     string   `gorm:"type:varchar(256);default:null;column:support_url" json:"support_url"`
	UninstallNotes string   `gorm:"type:varchar(1000);default:null;column:uninstall_notes" json:"uninstall_notes"`
}

func (Property) TableName() string {
	return "property"
}

type UpdateLanguage struct {
	VersionCheckMixin
	LanguageIndex int    `gorm:"primary_key;column:LanguageIndex" json:"LanguageIndex"`
	LanguageID    int    `gorm:"not null;column:LanguageID;unique_index" json:"LanguageID"`
	Enabled       bool   `gorm:"default:false;column:Enabled" json:"Enabled"`
	LongName      string `gorm:"type:varchar(256);default:null;column:LongName" json:"LongName"`
	ShortName     string `gorm:"type:varchar(256);default:null;column:ShortName" json:"ShortName"`
}

func (UpdateLanguage) TableName() string {
	return "update_language"
}
