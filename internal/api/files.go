package api

// ProjectFileType represents a file attached to a project.
type ProjectFileType struct {
	FileID    int    `json:"fileId"`
	LocalName string `json:"localName"`
	FullPath  string `json:"fullPath,omitempty"`
	MimeType  string `json:"mimeType,omitempty"`
	Key       string `json:"key,omitempty"`
}

// AddFileAck is the response after uploading a file.
type AddFileAck struct {
	FileID       int    `json:"fileId"`
	FileFullPath string `json:"fileFullPath,omitempty"`
	Key          string `json:"key,omitempty"`
}
