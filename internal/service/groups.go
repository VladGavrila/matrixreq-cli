package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// GroupService manages group operations.
type GroupService interface {
	List(details bool) ([]api.GroupType, error)
	Get(groupID int, details bool) (*api.GroupType, error)
	Create(groupName string) error
	Delete(groupID int) error
	Rename(groupID int, newName string) error
	AddUser(groupID int, user string) error
	RemoveUser(groupName, user string) error
	SetProjectPermission(groupID int, project string, permission int) error
}

// GetGroupListAck wraps the group list response.
type GetGroupListAck struct {
	GroupList []api.GroupType `json:"groupList,omitempty"`
}

type groupService struct {
	client *client.Client
}

func (s *groupService) List(details bool) ([]api.GroupType, error) {
	d := "0"
	if details {
		d = "1"
	}
	data, err := s.client.Get("/group?details=" + d)
	if err != nil {
		return nil, err
	}
	var resp GetGroupListAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing group list: %w", err)
	}
	return resp.GroupList, nil
}

func (s *groupService) Get(groupID int, details bool) (*api.GroupType, error) {
	d := "0"
	if details {
		d = "1"
	}
	path := fmt.Sprintf("/group/%d?details=%s", groupID, d)
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp api.GroupType
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing group: %w", err)
	}
	return &resp, nil
}

func (s *groupService) Create(groupName string) error {
	path := "/group?groupLabel=" + url.QueryEscape(groupName)
	_, err := s.client.Post(path, nil)
	return err
}

func (s *groupService) Delete(groupID int) error {
	path := fmt.Sprintf("/group/%d?confirm=yes", groupID)
	_, err := s.client.Delete(path)
	return err
}

func (s *groupService) Rename(groupID int, newName string) error {
	path := fmt.Sprintf("/group/%d/rename?newName=%s", groupID, url.QueryEscape(newName))
	_, err := s.client.Put(path, nil)
	return err
}

func (s *groupService) AddUser(groupID int, user string) error {
	path := fmt.Sprintf("/group/%d/user/%s", groupID, url.PathEscape(user))
	_, err := s.client.Put(path, nil)
	return err
}

func (s *groupService) RemoveUser(groupName, user string) error {
	path := fmt.Sprintf("/group/%s/user/%s", url.PathEscape(groupName), url.PathEscape(user))
	_, err := s.client.Delete(path)
	return err
}

func (s *groupService) SetProjectPermission(groupID int, project string, permission int) error {
	path := fmt.Sprintf("/group/%d/project/%s?permission=%d",
		groupID, url.PathEscape(project), permission)
	_, err := s.client.Post(path, nil)
	return err
}
