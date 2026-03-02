package api

// TrimItem represents a full item response.
type TrimItem struct {
	Title              string               `json:"title"`
	ItemRef            string               `json:"itemRef"`
	FolderRef          string               `json:"folderRef,omitempty"`
	UpLinkList         []TrimLink           `json:"upLinkList,omitempty"`
	DownLinkList       []TrimLink           `json:"downLinkList,omitempty"`
	FieldValList       *FieldValListType    `json:"fieldValList,omitempty"`
	Labels             []string             `json:"labels,omitempty"`
	ItemHistoryList    *ItemHistoryListType `json:"itemHistoryList,omitempty"`
	MaxVersion         int                  `json:"maxVersion,omitempty"`
	Disabled           int                  `json:"disabled,omitempty"`
	IsFolder           int                  `json:"isFolder,omitempty"`
	AvailableFormats   []string             `json:"availableFormats,omitempty"`
	ItemID             int                  `json:"itemId,omitempty"`
	ModDate            string               `json:"modDate,omitempty"`
	ModDateUserFormat  string               `json:"modDateUserFormat,omitempty"`
	RequireSubTree     []CategoryAndRoot    `json:"requireSubTree,omitempty"`
	SelectSubTree      []CategoryAndRoot    `json:"selectSubTree,omitempty"`
	IsUnselected       int                  `json:"isUnselected,omitempty"`
	Downloads          []UserAndTime        `json:"downloads,omitempty"`
	DocHasPackage      bool                 `json:"docHasPackage,omitempty"`
	CleanupFail        *CleanupFail         `json:"cleanupFail,omitempty"`
	ContextTree        *FancyLeaf           `json:"contextTree,omitempty"`
	CrossLinks         []CrossProjectLink   `json:"crossLinks,omitempty"`
}

// TrimFolder represents a folder item (recursive tree).
type TrimFolder struct {
	ItemRef           string               `json:"itemRef"`
	Title             string               `json:"title"`
	Partial           int                  `json:"partial,omitempty"`
	ItemList          []TrimFolder         `json:"itemList,omitempty"`
	FieldValList      *FieldValListType    `json:"fieldValList,omitempty"`
	IsFolder          int                  `json:"isFolder,omitempty"`
	IsUnselected      int                  `json:"isUnselected,omitempty"`
	ItemHistoryList   *ItemHistoryListType `json:"itemHistoryList,omitempty"`
	MaxVersion        int                  `json:"maxVersion,omitempty"`
	ModDate           string               `json:"modDate,omitempty"`
	ModDateUserFormat string               `json:"modDateUserFormat,omitempty"`
	ItemID            int                  `json:"itemId,omitempty"`
	Disabled          int                  `json:"disabled,omitempty"`
	ContextTree       *FancyLeaf           `json:"contextTree,omitempty"`
	CrossLinks        []CrossProjectLink   `json:"crossLinks,omitempty"`
}

// FieldValListType wraps a list of field values.
type FieldValListType struct {
	FieldVal []FieldValType `json:"fieldVal,omitempty"`
}

// FieldValType represents a single field value.
type FieldValType struct {
	ID         int    `json:"id"`
	Value      string `json:"value"`
	Hide       int    `json:"hide,omitempty"`
	Restricted int    `json:"restricted,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	FieldType  string `json:"fieldType,omitempty"`
}

// TrimLink represents an up/down link.
type TrimLink struct {
	UpLinkList        []TrimLink `json:"upLinkList,omitempty"`
	DownLinkList      []TrimLink `json:"downLinkList,omitempty"`
	ItemRef           string     `json:"itemRef"`
	Title             string     `json:"title,omitempty"`
	ModDate           string     `json:"modDate,omitempty"`
	ModDateUserFormat string     `json:"modDateUserFormat,omitempty"`
}

// ItemHistoryListType wraps a list of item history entries.
type ItemHistoryListType struct {
	ItemHistory []ItemHistoryType `json:"itemHistory,omitempty"`
}

// ItemHistoryType represents a version history entry.
type ItemHistoryType struct {
	Version            int    `json:"version"`
	CreatedAt          string `json:"createdAt,omitempty"`
	CreatedAtUserFormat string `json:"createdAtUserFormat,omitempty"`
	DeletedAt          string `json:"deletedAt,omitempty"`
	DeletedAtUserFormat string `json:"deletedAtUserFormat,omitempty"`
	Title              string `json:"title,omitempty"`
	CreatedByUserID    int    `json:"createdByUserId,omitempty"`
	CreatedByUserLogin string `json:"createdByUserLogin,omitempty"`
	Reason             string `json:"reason,omitempty"`
	AuditID            int    `json:"auditId,omitempty"`
	AuditAction        string `json:"auditAction,omitempty"`
}

// CategoryAndRoot links a category to its root folder.
type CategoryAndRoot struct {
	Category   string `json:"category"`
	RootFolder string `json:"rootFolder"`
}

// UserAndTime represents a user action with timestamp.
type UserAndTime struct {
	UserID         int    `json:"userId"`
	Login          string `json:"login"`
	FirstName      string `json:"firstName,omitempty"`
	LastName       string `json:"lastName,omitempty"`
	Email          string `json:"email,omitempty"`
	Date           string `json:"date,omitempty"`
	DateUserFormat string `json:"dateUserFormat,omitempty"`
}

// CleanupFail indicates a cleanup issue.
type CleanupFail struct {
	Fields             []CleanupField `json:"fields,omitempty"`
	TitleCleanedUp     bool           `json:"titleCleanedUp,omitempty"`
	TitleBeforeCleanup string         `json:"titleBeforeCleanup,omitempty"`
	TitleAfterCleanup  string         `json:"titleAfterCleanup,omitempty"`
	ItemRef            string         `json:"itemRef,omitempty"`
}

// CleanupField represents a field cleanup issue.
type CleanupField struct {
	FieldID       int    `json:"fieldId"`
	FieldLabel    string `json:"fieldLabel,omitempty"`
	FieldType     string `json:"fieldType,omitempty"`
	BeforeCleanup string `json:"beforeCleanup,omitempty"`
	AfterCleanup  string `json:"afterCleanup,omitempty"`
}

// FancyLeaf represents a tree node.
type FancyLeaf struct {
	ID           string `json:"id,omitempty"`
	Title        string `json:"title,omitempty"`
	Type         string `json:"type,omitempty"`
	IsUnselected int    `json:"isUnselected,omitempty"`
	Version      string `json:"version,omitempty"`
	Mode         string `json:"mode,omitempty"`
}

// CrossProjectLink represents a cross-project link.
type CrossProjectLink struct {
	UpItem          *OneItem `json:"upItem,omitempty"`
	DownItem        *OneItem `json:"downItem,omitempty"`
	Relation        string   `json:"relation,omitempty"`
	CreationDate    string   `json:"creationDate,omitempty"`
	ImportHistoryID int      `json:"importHistoryId,omitempty"`
}

// OneItem represents a single item reference.
type OneItem struct {
	ItemID             int    `json:"itemId"`
	Version            int    `json:"version,omitempty"`
	ProjectShort       string `json:"projectShort,omitempty"`
	ItemRefWithVersion string `json:"itemRefWithVersion,omitempty"`
	ItemTitle          string `json:"itemTitle,omitempty"`
}

// ItemSimpleType is a simplified item representation.
type ItemSimpleType struct {
	Author  string `json:"author,omitempty"`
	Birth   string `json:"birth,omitempty"`
	Ref     string `json:"ref"`
	Title   string `json:"title"`
	Version int    `json:"version,omitempty"`
}
