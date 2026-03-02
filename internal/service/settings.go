package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// SettingsService manages project-level settings.
type SettingsService interface {
	Get(project string) ([]api.SettingType, error)
	Set(project, key, value string) error
	GetSchema(project string) ([]byte, error)
}

// GetProjectSettingAck wraps the project settings response.
type GetProjectSettingAck struct {
	Settings []api.SettingType `json:"settingList,omitempty"`
}

type settingsService struct {
	client *client.Client
}

func (s *settingsService) Get(project string) ([]api.SettingType, error) {
	data, err := s.client.Get("/" + url.PathEscape(project) + "/setting")
	if err != nil {
		return nil, err
	}
	var resp GetProjectSettingAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	return resp.Settings, nil
}

func (s *settingsService) Set(project, key, value string) error {
	path := fmt.Sprintf("/%s/setting?key=%s&value=%s",
		url.PathEscape(project), url.QueryEscape(key), url.QueryEscape(value))
	_, err := s.client.Post(path, nil)
	return err
}

func (s *settingsService) GetSchema(project string) ([]byte, error) {
	return s.client.Get("/" + url.PathEscape(project) + "/schema")
}
