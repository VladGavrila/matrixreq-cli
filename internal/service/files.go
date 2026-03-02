package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// FileService manages file operations.
type FileService interface {
	List(project string) ([]api.ProjectFileType, error)
	Get(project string, fileNo int, key string) (*http.Response, error)
	Upload(project, fileName string, fileData io.Reader) (*api.AddFileAck, error)
}

// GetProjectFileListAck wraps the file list response.
type GetProjectFileListAck struct {
	FileList []api.ProjectFileType `json:"fileList,omitempty"`
}

type fileService struct {
	client *client.Client
}

func (s *fileService) List(project string) ([]api.ProjectFileType, error) {
	data, err := s.client.Get("/" + url.PathEscape(project) + "/file")
	if err != nil {
		return nil, err
	}
	var resp GetProjectFileListAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing file list: %w", err)
	}
	return resp.FileList, nil
}

func (s *fileService) Get(project string, fileNo int, key string) (*http.Response, error) {
	path := fmt.Sprintf("/%s/file/%d?key=%s",
		url.PathEscape(project), fileNo, url.QueryEscape(key))
	return s.client.GetRaw(path)
}

func (s *fileService) Upload(project, fileName string, fileData io.Reader) (*api.AddFileAck, error) {
	path := "/" + url.PathEscape(project) + "/file"
	data, err := s.client.PostForm(path, nil, fileName, fileData)
	if err != nil {
		return nil, err
	}
	var resp api.AddFileAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing upload response: %w", err)
	}
	return &resp, nil
}
