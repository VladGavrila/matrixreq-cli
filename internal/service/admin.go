package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// AdminService manages system administration operations.
type AdminService interface {
	Status() (*api.ServerStatus, error)
	License() ([]byte, error)
	Monitor() ([]byte, error)
	GetSettings() ([]api.SettingType, error)
	SetSetting(key, value string) error
}

// GetSettingAck wraps the settings response.
type GetSettingAck struct {
	Settings []api.SettingType `json:"settingList,omitempty"`
}

type adminService struct {
	client *client.Client
}

func (s *adminService) Status() (*api.ServerStatus, error) {
	data, err := s.client.Get("/all/status")
	if err != nil {
		return nil, err
	}
	var resp api.ServerStatus
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing status: %w", err)
	}
	return &resp, nil
}

func (s *adminService) License() ([]byte, error) {
	return s.client.Get("/all/license")
}

func (s *adminService) Monitor() ([]byte, error) {
	return s.client.Get("/all/monitor")
}

func (s *adminService) GetSettings() ([]api.SettingType, error) {
	data, err := s.client.Get("/all/setting")
	if err != nil {
		return nil, err
	}
	var resp GetSettingAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	return resp.Settings, nil
}

func (s *adminService) SetSetting(key, value string) error {
	path := fmt.Sprintf("/all/setting?key=%s&value=%s",
		url.QueryEscape(key), url.QueryEscape(value))
	_, err := s.client.Post(path, nil)
	return err
}
