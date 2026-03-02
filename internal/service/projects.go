package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// ProjectService manages project operations.
type ProjectService interface {
	List() ([]api.ProjectType, error)
	Get(project string) (*api.ProjectInfo, error)
	Create(label, shortLabel string) error
	Delete(project, confirm string) error
	Tree(project string, filter string) ([]byte, error)
	Access(project string) (*GetAccessAck, error)
	Audit(project string, startAt, maxResults int) (*api.TrimAuditList, error)
	Hide(project, reason string) error
	Unhide(project, newShort, reason string) error
}

// GetAccessAck wraps the access response.
type GetAccessAck struct {
	GroupPermission []api.GroupPermissionType `json:"groupPermission,omitempty"`
	UserPermission  []api.UserPermissionType `json:"userPermission,omitempty"`
}

type projectService struct {
	client *client.Client
}

func (s *projectService) List() ([]api.ProjectType, error) {
	data, err := s.client.Get("/")
	if err != nil {
		return nil, err
	}
	var resp api.ListProjectAndSettings
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing project list: %w", err)
	}
	return resp.Project, nil
}

func (s *projectService) Get(project string) (*api.ProjectInfo, error) {
	data, err := s.client.Get("/" + url.PathEscape(project))
	if err != nil {
		return nil, err
	}
	var resp api.ProjectInfo
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing project info: %w", err)
	}
	return &resp, nil
}

func (s *projectService) Create(label, shortLabel string) error {
	path := fmt.Sprintf("/?label=%s&shortLabel=%s",
		url.QueryEscape(label), url.QueryEscape(shortLabel))
	_, err := s.client.Post(path, nil)
	return err
}

func (s *projectService) Delete(project, confirm string) error {
	path := fmt.Sprintf("/%s?confirm=%s",
		url.PathEscape(project), url.QueryEscape(confirm))
	_, err := s.client.Delete(path)
	return err
}

func (s *projectService) Tree(project string, filter string) ([]byte, error) {
	path := "/" + url.PathEscape(project) + "/tree"
	if filter != "" {
		path += "?filter=" + url.QueryEscape(filter)
	}
	return s.client.Get(path)
}

func (s *projectService) Access(project string) (*GetAccessAck, error) {
	data, err := s.client.Get("/" + url.PathEscape(project) + "/access")
	if err != nil {
		return nil, err
	}
	var resp GetAccessAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing access: %w", err)
	}
	return &resp, nil
}

func (s *projectService) Audit(project string, startAt, maxResults int) (*api.TrimAuditList, error) {
	path := fmt.Sprintf("/%s/audit?startAt=%d&maxResults=%d",
		url.PathEscape(project), startAt, maxResults)
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp api.TrimAuditList
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing audit: %w", err)
	}
	return &resp, nil
}

func (s *projectService) Hide(project, reason string) error {
	path := fmt.Sprintf("/%s/hide?reason=%s",
		url.PathEscape(project), url.QueryEscape(reason))
	_, err := s.client.Put(path, nil)
	return err
}

func (s *projectService) Unhide(project, newShort, reason string) error {
	path := fmt.Sprintf("/%s/unhide?newShort=%s&reason=%s",
		url.PathEscape(project), url.QueryEscape(newShort), url.QueryEscape(reason))
	_, err := s.client.Put(path, nil)
	return err
}
