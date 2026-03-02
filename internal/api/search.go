package api

// TrimNeedle represents a search request.
type TrimNeedle struct {
	Search       string `json:"search,omitempty"`
	ID           string `json:"id,omitempty"`
	Treeorder    int    `json:"treeorder,omitempty"`
	Filter       string `json:"filter,omitempty"`
	FieldList    string `json:"fieldList,omitempty"`
	Labels       string `json:"labels,omitempty"`
	FromDate     string `json:"fromDate,omitempty"`
	ToDate       string `json:"toDate,omitempty"`
	CrossProject string `json:"crossProject,omitempty"`
}

// TrimNeedleItem represents a single search result item.
type TrimNeedleItem struct {
	ItemOrFolderRef string       `json:"itemOrFolderRef"`
	Title           string       `json:"title"`
	Project         string       `json:"project,omitempty"`
	FieldVal        []FieldValType `json:"fieldVal,omitempty"`
	Labels          string       `json:"labels,omitempty"`
	LastModDate     string       `json:"lastModDate,omitempty"`
	CreationDate    string       `json:"creationDate,omitempty"`
	UpLinkList      []TrimLink   `json:"upLinkList,omitempty"`
	DownLinkList    []TrimLink   `json:"downLinkList,omitempty"`
}

// SearchResult wraps search response data.
type SearchResult struct {
	Items []TrimNeedleItem `json:"needleResponse,omitempty"`
}
