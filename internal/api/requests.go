package api

// CreateItemRequest is the body for creating an item.
type CreateItemRequest struct {
	Title  string            `json:"title"`
	Folder string            `json:"folder,omitempty"`
	Reason string            `json:"reason"`
	Fields []FieldValSetType `json:"fieldVal,omitempty"`
	Labels []string          `json:"labels,omitempty"`
	Author string            `json:"author,omitempty"`
}

// UpdateItemRequest is the body for updating an item.
type UpdateItemRequest struct {
	Title    string            `json:"title,omitempty"`
	Reason   string            `json:"reason"`
	Fields   []FieldValSetType `json:"fieldVal,omitempty"`
	Labels   []string          `json:"labels,omitempty"`
	OnlyThose bool            `json:"onlyThoseFields,omitempty"`
	OnlyLabels bool           `json:"onlyThoseLabels,omitempty"`
}

// FieldValSetType is used to set a field value on create/update.
type FieldValSetType struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

// CreateFolderRequest is the body for creating a folder.
type CreateFolderRequest struct {
	Label  string `json:"label"`
	Parent string `json:"parent"`
	Reason string `json:"reason"`
}

// CreateCategoryRequest is the body for creating a category.
type CreateCategoryRequest struct {
	Label      string `json:"label"`
	ShortLabel string `json:"shortLabel"`
	Reason     string `json:"reason,omitempty"`
}

// MoveItemRequest is the body for moving an item.
type MoveItemRequest struct {
	Reason  string `json:"reason"`
	NewFolder string `json:"newFolder"`
	CopyLabels int  `json:"copyLabels,omitempty"`
}

// CreateLinkRequest is the body for creating a traceability link.
type CreateLinkRequest struct {
	UpItemRef   string `json:"upItemRef"`
	DownItemRef string `json:"downItemRef"`
	Reason      string `json:"reason"`
}

// CreateTodoRequest is the body for creating a todo.
type CreateTodoRequest struct {
	Login    string `json:"login,omitempty"`
	ItemRef  string `json:"itemRef,omitempty"`
	FieldID  int    `json:"fieldId,omitempty"`
	Auto     bool   `json:"auto,omitempty"`
	Text     string `json:"text,omitempty"`
	TodoType string `json:"todoType,omitempty"`
	Future   bool   `json:"future,omitempty"`
}

// CreateTokenRequest is the body for creating an API token.
type CreateTokenRequest struct {
	Purpose string `json:"purpose"`
	Reason  string `json:"reason"`
	ValidTo string `json:"validTo,omitempty"`
}

// CreateGroupRequest is the body for creating a group.
type CreateGroupRequest struct {
	GroupName string `json:"groupName"`
}

// ReportRequest is the body for generating a report.
type ReportRequest struct {
	ItemRef    string `json:"itemRef"`
	ReportID   string `json:"reportId,omitempty"`
	Format     string `json:"format,omitempty"`
	IsSignedReport bool `json:"isSignedReport,omitempty"`
}

// SignItemRequest is the body for signing an item.
type SignItemRequest struct {
	Password string `json:"password"`
	Meaning  string `json:"meaning,omitempty"`
}

// CrossProjectLinkRequest is the body for creating a cross-project link.
type CrossProjectLinkRequest struct {
	UpProjectShort   string `json:"upProjectShort"`
	UpItemRef        string `json:"upItemRef"`
	DownProjectShort string `json:"downProjectShort"`
	DownItemRef      string `json:"downItemRef"`
	Relation         string `json:"relation,omitempty"`
}
