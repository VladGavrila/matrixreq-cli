package api

// ProjectType represents a Matrix project.
type ProjectType struct {
	ID           int    `json:"id"`
	Label        string `json:"label"`
	ShortLabel   string `json:"shortLabel"`
	ProjectLogo  string `json:"projectLogo,omitempty"`
	QMSProject   bool   `json:"qmsProject,omitempty"`
	AccessType   string `json:"accessType,omitempty"`
	UniqueIDs    bool   `json:"uniqueIds,omitempty"`
}

// CategoryType represents a category within a project.
type CategoryType struct {
	ID         int    `json:"id"`
	Label      string `json:"label"`
	ShortLabel string `json:"shortLabel"`
	MaxSerial  int    `json:"maxSerial,omitempty"`
}

// CategoryExtendedType includes field list and enabled categories.
type CategoryExtendedType struct {
	Category  CategoryType  `json:"category"`
	FieldList FieldListType `json:"fieldList"`
	Enable    []string      `json:"enable,omitempty"`
}

// CategoryExtendedListWrapper handles the API's wrapped format:
// {"categoryExtended": [...]} instead of a bare array.
type CategoryExtendedListWrapper struct {
	CategoryExtended []CategoryExtendedType `json:"categoryExtended"`
}

// FieldType represents a field definition.
type FieldType struct {
	ID        int    `json:"id"`
	Order     int    `json:"order"`
	FieldType string `json:"fieldType"`
	Parameter string `json:"parameter,omitempty"`
	Label     string `json:"label"`
}

// FieldListType wraps a list of fields.
type FieldListType struct {
	Field []FieldType `json:"field,omitempty"`
}

// SettingType represents a configuration setting.
type SettingType struct {
	Value  string `json:"value"`
	Key    string `json:"key"`
	Secret bool   `json:"secret,omitempty"`
}

// ServerStatus represents the instance status response from GET /all/status.
type ServerStatus struct {
	ExceptionStatus *ExceptionStatus `json:"exceptionStatus,omitempty"`
	Version         string           `json:"version"`
	PublicURL       string           `json:"publicUrl,omitempty"`
}

// ExceptionStatus holds exception info.
type ExceptionStatus struct {
	NbExceptionsStillStart int                `json:"nbExceptionsStillStart"`
	LastHourExceptions     []ExceptionItemIso `json:"lastHourExceptions,omitempty"`
}

// ExceptionItemIso represents a single exception entry.
type ExceptionItemIso struct {
	Date string `json:"date"`
	Text string `json:"text"`
}

// MatrixLicense represents license information.
type MatrixLicense struct {
	MaxUsers       int    `json:"maxUsers,omitempty"`
	CurrentUsers   int    `json:"currentUsers,omitempty"`
	LicenseType    string `json:"licenseType,omitempty"`
	ExpirationDate string `json:"expirationDate,omitempty"`
}

// MainAndBranch represents a branch relationship.
type MainAndBranch struct {
	MainProject   string `json:"mainProject,omitempty"`
	BranchProject string `json:"branchProject,omitempty"`
	BranchDate    string `json:"branchDate,omitempty"`
	BranchTag     string `json:"branchTag,omitempty"`
}

// GetDateAck represents date format info.
type GetDateAck struct {
	DateFormat     string `json:"dateFormat,omitempty"`
	TimeFormat     string `json:"timeFormat,omitempty"`
	DateTimeFormat string `json:"dateTimeFormat,omitempty"`
	TimeZone       string `json:"timeZone,omitempty"`
}
