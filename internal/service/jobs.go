package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// JobService manages async job operations.
type JobService interface {
	List(project string) ([]api.JobWithUrl, error)
	Get(project string, jobID int) (*api.JobWithUrl, error)
	Cancel(project string, jobID int, reason string) error
	Download(project string, jobID, fileNo int) (*http.Response, error)
}

type jobService struct {
	client *client.Client
}

func (s *jobService) List(project string) ([]api.JobWithUrl, error) {
	data, err := s.client.Get("/" + url.PathEscape(project) + "/job")
	if err != nil {
		return nil, err
	}
	var resp api.JobsWithUrl
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing job list: %w", err)
	}
	return resp.Jobs, nil
}

func (s *jobService) Get(project string, jobID int) (*api.JobWithUrl, error) {
	path := fmt.Sprintf("/%s/job/%d", url.PathEscape(project), jobID)
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp api.JobWithUrl
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing job: %w", err)
	}
	return &resp, nil
}

func (s *jobService) Cancel(project string, jobID int, reason string) error {
	path := fmt.Sprintf("/%s/job/%d?reason=%s",
		url.PathEscape(project), jobID, url.QueryEscape(reason))
	_, err := s.client.Delete(path)
	return err
}

func (s *jobService) Download(project string, jobID, fileNo int) (*http.Response, error) {
	path := fmt.Sprintf("/%s/job/%d/%d", url.PathEscape(project), jobID, fileNo)
	return s.client.GetRaw(path)
}
