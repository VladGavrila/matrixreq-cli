package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// CategoryService manages category operations.
type CategoryService interface {
	List(project string) ([]api.CategoryExtendedType, error)
	Get(project, category string) (*CategoryFull, error)
	Create(project, label, shortLabel, reason string) error
	Update(project, category, label, shortLabel, reason string, order int) error
	Delete(project, category, reason string) error
	GetSettings(project, category string) ([]api.SettingType, error)
	SetSetting(project, category, key, value string) error
}

// CategoryFull is the full category response.
type CategoryFull struct {
	FolderList []api.TrimFolder         `json:"folderList,omitempty"`
	FieldList  api.FieldListType        `json:"fieldList,omitempty"`
	Category   api.CategoryType         `json:"category,omitempty"`
	Enable     []string                 `json:"enable,omitempty"`
}

// GetProjectStructAck wraps the category list response.
// The API returns categoryList as {"categoryExtended": [...]}.
type GetProjectStructAck struct {
	CategoryList api.CategoryExtendedListWrapper `json:"categoryList,omitempty"`
}

type categoryService struct {
	client *client.Client
}

func (s *categoryService) List(project string) ([]api.CategoryExtendedType, error) {
	data, err := s.client.Get("/" + url.PathEscape(project) + "/cat")
	if err != nil {
		return nil, err
	}
	var resp GetProjectStructAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing categories: %w", err)
	}
	return resp.CategoryList.CategoryExtended, nil
}

func (s *categoryService) Get(project, category string) (*CategoryFull, error) {
	path := fmt.Sprintf("/%s/cat/%s", url.PathEscape(project), url.PathEscape(category))
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp CategoryFull
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing category: %w", err)
	}
	return &resp, nil
}

func (s *categoryService) Create(project, label, shortLabel, reason string) error {
	path := fmt.Sprintf("/%s?label=%s&shortLabel=%s&reason=%s",
		url.PathEscape(project),
		url.QueryEscape(label), url.QueryEscape(shortLabel), url.QueryEscape(reason))
	_, err := s.client.Post(path, nil)
	return err
}

func (s *categoryService) Update(project, category, label, shortLabel, reason string, order int) error {
	path := fmt.Sprintf("/%s/cat/%s?label=%s&shortLabel=%s&reason=%s&order=%d",
		url.PathEscape(project), url.PathEscape(category),
		url.QueryEscape(label), url.QueryEscape(shortLabel), url.QueryEscape(reason), order)
	_, err := s.client.Put(path, nil)
	return err
}

func (s *categoryService) Delete(project, category, reason string) error {
	path := fmt.Sprintf("/%s/cat/%s?reason=%s",
		url.PathEscape(project), url.PathEscape(category), url.QueryEscape(reason))
	_, err := s.client.Delete(path)
	return err
}

func (s *categoryService) GetSettings(project, category string) ([]api.SettingType, error) {
	path := fmt.Sprintf("/%s/cat/%s/setting", url.PathEscape(project), url.PathEscape(category))
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp GetSettingAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing settings: %w", err)
	}
	return resp.Settings, nil
}

func (s *categoryService) SetSetting(project, category, key, value string) error {
	path := fmt.Sprintf("/%s/cat/%s/setting?key=%s&value=%s",
		url.PathEscape(project), url.PathEscape(category),
		url.QueryEscape(key), url.QueryEscape(value))
	_, err := s.client.Post(path, nil)
	return err
}
