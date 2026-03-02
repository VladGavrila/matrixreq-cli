package api

// ListProjectAndSettings is the response from GET /.
type ListProjectAndSettings struct {
	CurrentUser         string          `json:"currentUser,omitempty"`
	CustomerAdmin       int             `json:"customerAdmin,omitempty"`
	SuperAdmin          int             `json:"superAdmin,omitempty"`
	DateInfo            *GetDateAck     `json:"dateInfo,omitempty"`
	CustomerSettings    []SettingType   `json:"customerSettings,omitempty"`
	License             *MatrixLicense  `json:"license,omitempty"`
	ReadWriteUsers      []string        `json:"readWriteUsers,omitempty"`
	AllUsers            []UserType      `json:"allUsers,omitempty"`
	LicenseStatus       string          `json:"licenseStatus,omitempty"`
	TodoCounts          []TodoCount     `json:"todoCounts,omitempty"`
	AllTodos            []Todo          `json:"allTodos,omitempty"`
	CurrentUserSettings []SettingType   `json:"currentUserSettings,omitempty"`
	Branches            []MainAndBranch `json:"branches,omitempty"`
	ServiceEmail        string          `json:"serviceEmail,omitempty"`
	Project             []ProjectType   `json:"project,omitempty"`
	ServerVersion       string          `json:"serverVersion,omitempty"`
	BaseURL             string          `json:"baseUrl,omitempty"`
	RestURL             string          `json:"restUrl,omitempty"`
}

// ProjectInfo is the response from GET /{project}.
type ProjectInfo struct {
	Label            string                       `json:"label,omitempty"`
	ShortLabel       string                       `json:"shortLabel,omitempty"`
	Access           *Access                      `json:"access,omitempty"`
	CategoryList     CategoryExtendedListWrapper  `json:"categoryList,omitempty"`
	GroupPermission  []GroupPermissionType         `json:"groupPermission,omitempty"`
	UserPermission   []UserPermissionType         `json:"userPermission,omitempty"`
	PluginSettings   []PluginSetting              `json:"pluginSettings,omitempty"`
	ProjectSettings  []SettingType                `json:"projectSettings,omitempty"`
}

// Access represents project access settings.
type Access struct {
	StartDate8601 string `json:"startDate8601,omitempty"`
	EndDate8601   string `json:"endDate8601,omitempty"`
	ReadWrite     int    `json:"readWrite,omitempty"`
	VisitorOnly   bool   `json:"visitorOnly,omitempty"`
}

// PluginSetting represents a plugin configuration.
type PluginSetting struct {
	PluginID     int                `json:"pluginId,omitempty"`
	PluginName   string             `json:"pluginName,omitempty"`
	Capabilities *PluginCapabilities `json:"capabilities,omitempty"`
	Settings     []SettingType      `json:"settings,omitempty"`
}

// PluginCapabilities describes what a plugin can do.
type PluginCapabilities struct {
	CanCreate            bool `json:"canCreate,omitempty"`
	CanFind              bool `json:"canFind,omitempty"`
	NeedSetup            bool `json:"needSetup,omitempty"`
	HandleAsLink         bool `json:"handleAsLink,omitempty"`
	One2OneMapping       bool `json:"one2OneMapping,omitempty"`
	HasMeta              bool `json:"hasMeta,omitempty"`
	CanCreateBacklinks   bool `json:"canCreateBacklinks,omitempty"`
	Messaging            bool `json:"messaging,omitempty"`
	RestToken            bool `json:"restToken,omitempty"`
	Impersonate          bool `json:"impersonate,omitempty"`
	ExtendedSettings     bool `json:"extendedSettings,omitempty"`
	HideInProjectSettings bool `json:"hideInProjectSettings,omitempty"`
}

// AddItemAck is the response after creating an item.
type AddItemAck struct {
	ItemID      int          `json:"itemId"`
	Serial      int          `json:"serial"`
	OriginTag   string       `json:"originTag,omitempty"`
	CleanupFail *CleanupFail `json:"cleanupFail,omitempty"`
}

// CopyItemAck is the response after copying items.
type CopyItemAck struct {
	ItemsAndFoldersCreated []string `json:"itemsAndFoldersCreated,omitempty"`
}

// CreateReportJobAck is the response after creating a report job.
type CreateReportJobAck struct {
	JobID int `json:"jobId"`
}

// ExportItemsAck is the response after starting an export.
type ExportItemsAck struct {
	JobID int `json:"jobId"`
}

// SignItemAck is the response after signing an item.
type SignItemAck struct {
	Result string `json:"result,omitempty"`
	OK     bool   `json:"ok"`
}

// UndeleteAnswer is the response after restoring an item.
type UndeleteAnswer struct {
	NewParent string `json:"newParent,omitempty"`
	NewOrder  int    `json:"newOrder,omitempty"`
}

// SettingAndValue represents a global setting.
type SettingAndValue struct {
	Setting   string `json:"setting"`
	Value     string `json:"value"`
	Encrypted bool   `json:"encrypted,omitempty"`
}
