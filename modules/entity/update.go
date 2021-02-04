package entity

type RevisionInfo struct {
	RevisionID       int    `json:"RevisionID"`
	LastIsLeafChange string `json:"LastIsLeafChange"`
	IsLeaf           bool   `json:"IsLeaf"`
	IsLatestRevision bool   `json:"IsLatestRevision"`
}

type DownloadFileInfo struct {
	RevisionID   int    `json:"RevisionID"`
	PatchingType string `json:"PatchingType"`
	IsOnServer   bool   `json:"IsOnServer"`
}

type DeploymentInfo struct {
	ActionID             int    `json:"ActionID"`
	DeploymentID         int    `json:"DeploymentID"`
	AutoSelect           int    `json:"AutoSelect"`
	AutoDownload         int    `json:"AutoDownload"`
	SupersedenceBehavior int    `json:"SupersedenceBehavior"`
	LastChangeTime       string `json:"LastChangeTime"`
	RevisionID           int    `json:"RevisionID"`
	TargetGroupID        string `json:"TargetGroupID"`
}

type RuleInfo struct {
	RevisionID      int    `json:"RevisionID"`
	RootElementType int    `json:"RootElementType"`
	RootElementXml  string `json:"RootElementXml"`
}

type ComputerInfo struct {
	TargetID     int    `json:"TargetID"`
	ClientID     string `json:"ClientID"`
	LastSyncTime string `json:"LastSyncTime"`
}

type SyncHistoryInfo struct {
	ID              int    `json:"ID"`
	ParentServerID  string `json:"ParentServerID"`
	ParentServerIP  string `json:"ParentServerIP"`
	LastSyncTime    string `json:"LastSyncTime"`
	ImportedTime    string `json:"ImportedTime"`
	StartTime       string `json:"StartTime"`
	FinishTime      string `json:"FinishTime"`
	SyncType        int    `json:"SyncType"`
	SyncStatus      int    `json:"SyncStatus"`
	NewUpdates      int    `json:"NewUpdates"`
	RevisedUpdates  int    `json:"RevisedUpdates"`
	ExpiredUpdates  int    `json:"ExpiredUpdates"`
	Pending         bool   `json:"Pending"`
	ReplicationMode string `json:"ReplicationMode"`
	SyncCategory    int    `json:"SyncCategory"`
}
