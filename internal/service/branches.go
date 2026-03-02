package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// BranchService manages branching and merging operations.
type BranchService interface {
	Create(project, label, shortLabel, tagToCreate string, keepPermissions, keepContent int) error
	Clone(project, label, shortLabel string, keepHistory, keepContent, keepPermissions int) error
	Merge(mainProject, branchProject, reason string, params *api.MergeParam) error
	Info(project string) ([]byte, error)
	History(project string) ([]api.MergeHistory, error)
}

type branchService struct {
	client *client.Client
}

func (s *branchService) Create(project, label, shortLabel, tagToCreate string, keepPermissions, keepContent int) error {
	path := fmt.Sprintf("/%s/branch?label=%s&shortLabel=%s&branch=1&keepPermissions=%d&keepContent=%d",
		url.PathEscape(project),
		url.QueryEscape(label), url.QueryEscape(shortLabel),
		keepPermissions, keepContent)
	if tagToCreate != "" {
		path += "&tagToCreate=" + url.QueryEscape(tagToCreate)
	}
	_, err := s.client.Post(path, nil)
	return err
}

func (s *branchService) Clone(project, label, shortLabel string, keepHistory, keepContent, keepPermissions int) error {
	path := fmt.Sprintf("/%s/clone?label=%s&shortLabel=%s&keepHistory=%d&keepContent=%d&keepPermissions=%d",
		url.PathEscape(project),
		url.QueryEscape(label), url.QueryEscape(shortLabel),
		keepHistory, keepContent, keepPermissions)
	_, err := s.client.Post(path, nil)
	return err
}

func (s *branchService) Merge(mainProject, branchProject, reason string, params *api.MergeParam) error {
	path := fmt.Sprintf("/%s/merge/%s?reason=%s",
		url.PathEscape(mainProject), url.PathEscape(branchProject), url.QueryEscape(reason))
	_, err := s.client.Post(path, params)
	return err
}

func (s *branchService) Info(project string) ([]byte, error) {
	return s.client.Get("/" + url.PathEscape(project) + "/mergeinfo")
}

func (s *branchService) History(project string) ([]api.MergeHistory, error) {
	data, err := s.client.Get("/" + url.PathEscape(project) + "/mergehistory")
	if err != nil {
		return nil, err
	}
	var resp []api.MergeHistory
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing merge history: %w", err)
	}
	return resp, nil
}
