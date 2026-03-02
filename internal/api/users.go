package api

// UserType represents a Matrix user.
type UserType struct {
	ID                int           `json:"id"`
	Login             string        `json:"login"`
	Email             string        `json:"email"`
	FirstName         string        `json:"firstName,omitempty"`
	LastName          string        `json:"lastName,omitempty"`
	SignatureImage    string        `json:"signatureImage,omitempty"`
	SignaturePassword string        `json:"signaturePassword,omitempty"`
	CustomerAdmin     int           `json:"customerAdmin,omitempty"`
	PasswordAgeInDays int           `json:"passwordAgeInDays,omitempty"`
	BadLogins         int           `json:"badLogins,omitempty"`
	BadLoginsBefore   int           `json:"badLoginsBefore,omitempty"`
	SuperAdmin        int           `json:"superAdmin,omitempty"`
	UserStatus        string        `json:"userStatus,omitempty"`
	UserSettingsList  []SettingType `json:"userSettingsList,omitempty"`
	TokenList         []TokenType   `json:"tokenList,omitempty"`
	GroupList         []int         `json:"groupList,omitempty"`
}

// UserTypeSimple is a minimal user representation.
type UserTypeSimple struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	Email string `json:"email,omitempty"`
}

// TokenType represents an API token.
type TokenType struct {
	UserID            int    `json:"userId"`
	TokenID           int    `json:"tokenId"`
	Purpose           string `json:"purpose,omitempty"`
	Reason            string `json:"reason,omitempty"`
	Value             string `json:"value,omitempty"`
	ValidTo           string `json:"validTo,omitempty"`
	ValidToUserFormat string `json:"validToUserFormat,omitempty"`
}

// UserPermissionType represents user permissions on a project.
type UserPermissionType struct {
	ID         int    `json:"id"`
	Login      string `json:"login"`
	Email      string `json:"email,omitempty"`
	Permission int    `json:"permission"`
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
}
