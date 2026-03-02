package api

// GroupType represents a Matrix group with its members.
type GroupType struct {
	GroupID    int              `json:"groupId"`
	GroupName  string           `json:"groupName"`
	Membership []UserTypeSimple `json:"membership,omitempty"`
}

// GroupPermissionType represents group permissions on a project.
type GroupPermissionType struct {
	GroupName  string           `json:"groupName"`
	Permission int              `json:"permission"`
	GroupID    int              `json:"groupId"`
	Membership []UserTypeSimple `json:"membership,omitempty"`
}
