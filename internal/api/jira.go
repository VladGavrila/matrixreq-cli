package api

// JiraPluginIDDefault is the Matrix plugin ID for the Jira Cloud add-on.
const JiraPluginIDDefault = 212

// JiraMatrixItem identifies a Matrix item by project + ref in wfgw payloads.
type JiraMatrixItem struct {
	Project    string `json:"project"`
	MatrixItem string `json:"matrixItem"`
}

// JiraExternalItem describes a Jira issue link in wfgw request payloads.
type JiraExternalItem struct {
	ExternalItemID    string   `json:"externalItemId"`
	ExternalItemTitle string   `json:"externalItemTitle"`
	ExternalItemURL   string   `json:"externalItemUrl"`
	Plugin            int      `json:"plugin"`
	MatrixItemIDs     []string `json:"matrixItemIds,omitempty"` // required on BreakLinks
}

// JiraGetIssuesPayload is the query payload for the GetIssues action.
type JiraGetIssuesPayload struct {
	Action     string         `json:"action"`
	MatrixItem JiraMatrixItem `json:"matrixItem"`
}

// JiraCreateLinksPayload is the form payload for the CreateLinks action.
type JiraCreateLinksPayload struct {
	Action        string             `json:"action"`
	MatrixItem    JiraMatrixItem     `json:"matrixItem"`
	ExternalItems []JiraExternalItem `json:"externalItems"`
}

// JiraBreakLinksBody is the JSON body for a BreakLinks DELETE.
type JiraBreakLinksBody struct {
	PluginID      int                `json:"pluginId"`
	Action        string             `json:"action"`
	MatrixItem    JiraMatrixItem     `json:"matrixItem"`
	ExternalItems []JiraExternalItem `json:"externalItems"`
}

// JiraLinkedIssue is a single link returned by GetIssues.
type JiraLinkedIssue struct {
	ExternalItemID           string `json:"externalItemId"`
	ExternalItemTitle        string `json:"externalItemTitle"`
	ExternalItemURL          string `json:"externalItemUrl"`
	ExternalLinkCreationDate string `json:"externalLinkCreationDate"`
	ExternalMeta             string `json:"externalMeta"` // JSON string: {issueType, status, ...}
	Plugin                   int    `json:"plugin"`
}

// JiraGetIssuesResponseEntry is one entry in the top-level GetIssues response array.
type JiraGetIssuesResponseEntry struct {
	Links []JiraLinkedIssue `json:"links"`
}

// JiraExternalMeta is the decoded shape of JiraLinkedIssue.ExternalMeta.
type JiraExternalMeta struct {
	IssueType string `json:"issueType"`
	Status    string `json:"status"`
	Summary   string `json:"summary"`
}
