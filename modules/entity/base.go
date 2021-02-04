package entity

import (
	"encoding/xml"
	"time"
)

type Int struct {
	Int []int `xml:"int"`
}

type String struct {
	String []string `xml:"string"`
}

type GUID struct {
	GUID []string `xml:"guid"`
}

type Base64Binary struct {
	Base64Binary []string `xml:"base64Binary"`
}

type UpdateIdentity struct {
	UpdateID       string `xml:"UpdateID"`
	RevisionNumber int    `xml:"RevisionNumber"`
}

type UpdateIdentityAttr struct {
	UpdateID       string `xml:"UpdateID,attr,omitempty"`
	RevisionNumber int    `xml:"RevisionNumber,attr,omitempty"`
}

type Cookie struct {
	Expiration    string `xml:"Expiration"`
	EncryptedData string `xml:"EncryptedData"`
}

type AuthorizationCookie struct {
	PlugInID   string `xml:"PlugInId"`
	CookieData string `xml:"CookieData"`
}

type Deployment struct {
	DeploymentID         int    `xml:"ID"`
	Action               string `xml:"Action"`
	IsAssigned           bool   `xml:"IsAssigned"`
	LastChangeTime       string `xml:"LastChangeTime"`
	AutoSelect           int    `xml:"AutoSelect"`
	AutoDownload         int    `xml:"AutoDownload"`
	SupersedenceBehavior int    `xml:"SupersedenceBehavior"`
	FlagBitmask          int    `xml:"FlagBitmask,omitempty"`
}

type UpdateInfo struct {
	ID         int         `xml:"ID"`
	Deployment Deployment  `xml:"Deployment"`
	IsLeaf     bool        `xml:"IsLeaf"`
	Xml        interface{} `xml:"Xml,omitempty"`
}

type Update struct {
	ID  int         `xml:"ID"`
	Xml interface{} `xml:"Xml"`
}

type FileLocation struct {
	FileDigest string `xml:"FileDigest"`
	Url        string `xml:"Url"`
}

type FileLocations struct {
	FileLocation []FileLocation `xml:"FileLocation"`
}

type CuspMsg struct {
	ClientID           string
	ServerID           string
	LastDeploymentTime time.Time
	ProtocolVersion    string
	CookieData         string
}

type LocalizedProperties struct {
	XMLName     xml.Name `xml:"LocalizedProperties"`
	Language    string   `xml:"Language"`
	Title       string   `xml:"Title"`
	Description string   `xml:"Description"`
}

type EulaFile struct {
	XMLName          xml.Name         `xml:"EulaFile"`
	Digest           string           `xml:"Digest"`
	DigestAlgorithm  string           `xml:"DigestAlgorithm"`
	FileName         string           `xml:"FileName"`
	Language         string           `xml:"Language"`
	Size             string           `xml:"Size"`
	AdditionalDigest AdditionalDigest `xml:"AdditionalDigest"`
}

type AdditionalDigest struct {
	Algorithm string `xml:"Algorithm,attr"`
	Value     string `xml:",chardata"`
}
