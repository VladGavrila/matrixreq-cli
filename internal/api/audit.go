package api

import "encoding/json"

// AuditItemRef represents the before/after state in an audit entry.
// The API returns this as an object with itemOrFolderRef and title.
type AuditItemRef struct {
	ItemOrFolderRef string `json:"itemOrFolderRef,omitempty"`
	Title           string `json:"title,omitempty"`
}

// String returns a readable representation.
func (r AuditItemRef) String() string {
	if r.ItemOrFolderRef != "" {
		return r.ItemOrFolderRef
	}
	return ""
}

// TrimAudit represents an audit log entry.
type TrimAudit struct {
	AuditID      int            `json:"auditId"`
	Action       string         `json:"action,omitempty"`
	Entity       string         `json:"entity,omitempty"`
	UserLogin    string         `json:"userLogin,omitempty"`
	DateTime     string         `json:"dateTime,omitempty"`
	DateUser     string         `json:"dateTimeUserFormat,omitempty"`
	Reason       string         `json:"reason,omitempty"`
	ProjectLabel string         `json:"projectLabel,omitempty"`
	ItemBefore   *AuditItemRef  `json:"itemBefore,omitempty"`
	ItemAfter    *AuditItemRef  `json:"itemAfter,omitempty"`
	TechAudit    any            `json:"techAudit,omitempty"`
}

// ItemRef returns the most relevant item reference from either before or after.
func (a TrimAudit) ItemRef() string {
	if a.ItemBefore != nil && a.ItemBefore.ItemOrFolderRef != "" {
		return a.ItemBefore.ItemOrFolderRef
	}
	if a.ItemAfter != nil && a.ItemAfter.ItemOrFolderRef != "" {
		return a.ItemAfter.ItemOrFolderRef
	}
	return ""
}

// TrimAuditList wraps a list of audit entries with pagination info.
type TrimAuditList struct {
	StartAt      int         `json:"startAt,omitempty"`
	MaxResults   int         `json:"maxResults,omitempty"`
	TotalResults int         `json:"totalResults,omitempty"`
	Audit        []TrimAudit `json:"audit,omitempty"`
}

// MarshalJSON implements custom marshaling to produce clean JSON output.
func (a AuditItemRef) MarshalJSON() ([]byte, error) {
	type Alias AuditItemRef
	return json.Marshal(Alias(a))
}
