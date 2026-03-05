package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// UserService manages user operations.
type UserService interface {
	List(details bool) ([]api.UserType, error)
	Get(user string) (*api.UserType, error)
	Create(login, email, password, first, last string) error
	Update(user, email, password, first, last string) error
	Delete(user string) error
	Rename(user, newLogin string) error
	SetStatus(user, status string) error
	CreateToken(user, purpose, reason string, validity int) (string, error)
	DeleteToken(user string, tokenID int) error
	Audit(user string, startAt, maxResults int) (*api.TrimAuditList, error)
	AddToProject(user, project string, permission int) error
	UpdateProjectPermission(user, project string, permission int) error
}

// GetUserListAck wraps the user list response.
type GetUserListAck struct {
	UserList []api.UserType `json:"user,omitempty"`
}

type userService struct {
	client *client.Client
}

func (s *userService) List(details bool) ([]api.UserType, error) {
	d := "0"
	if details {
		d = "1"
	}
	data, err := s.client.Get("/user?details=" + d)
	if err != nil {
		return nil, err
	}
	var resp GetUserListAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing user list: %w", err)
	}
	return resp.UserList, nil
}

func (s *userService) Get(user string) (*api.UserType, error) {
	data, err := s.client.Get("/user/" + url.PathEscape(user))
	if err != nil {
		return nil, err
	}
	var resp api.UserType
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing user: %w", err)
	}
	return &resp, nil
}

func (s *userService) Create(login, email, password, first, last string) error {
	path := fmt.Sprintf("/user?login=%s&email=%s&password=%s&first=%s&last=%s&json=1",
		url.QueryEscape(login), url.QueryEscape(email), url.QueryEscape(password),
		url.QueryEscape(first), url.QueryEscape(last))
	_, err := s.client.Post(path, nil)
	return err
}

func (s *userService) Update(user, email, password, first, last string) error {
	path := fmt.Sprintf("/user/%s?email=%s&password=%s&first=%s&last=%s&json=1",
		url.PathEscape(user), url.QueryEscape(email), url.QueryEscape(password),
		url.QueryEscape(first), url.QueryEscape(last))
	_, err := s.client.Put(path, nil)
	return err
}

func (s *userService) Delete(user string) error {
	path := fmt.Sprintf("/user/%s?confirm=yes", url.PathEscape(user))
	_, err := s.client.Delete(path)
	return err
}

func (s *userService) Rename(user, newLogin string) error {
	path := fmt.Sprintf("/user/%s/rename?newLogin=%s",
		url.PathEscape(user), url.QueryEscape(newLogin))
	_, err := s.client.Put(path, nil)
	return err
}

func (s *userService) SetStatus(user, status string) error {
	path := fmt.Sprintf("/user/%s/status?status=%s",
		url.PathEscape(user), url.QueryEscape(status))
	_, err := s.client.Put(path, nil)
	return err
}

func (s *userService) CreateToken(user, purpose, reason string, validity int) (string, error) {
	path := fmt.Sprintf("/user/%s/token?purpose=%s&reason=%s&validity=%d",
		url.PathEscape(user), url.QueryEscape(purpose), url.QueryEscape(reason), validity)
	data, err := s.client.Post(path, nil)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *userService) DeleteToken(user string, tokenID int) error {
	path := fmt.Sprintf("/user/%s/token?tokenId=%d", url.PathEscape(user), tokenID)
	_, err := s.client.Delete(path)
	return err
}

func (s *userService) Audit(user string, startAt, maxResults int) (*api.TrimAuditList, error) {
	path := fmt.Sprintf("/user/%s/audit?startAt=%d&maxResults=%d",
		url.PathEscape(user), startAt, maxResults)
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

func (s *userService) AddToProject(user, project string, permission int) error {
	path := fmt.Sprintf("/user/%s/%s?permission=%d",
		url.PathEscape(user), url.PathEscape(project), permission)
	_, err := s.client.Post(path, nil)
	return err
}

func (s *userService) UpdateProjectPermission(user, project string, permission int) error {
	path := fmt.Sprintf("/user/%s/%s?permission=%d",
		url.PathEscape(user), url.PathEscape(project), permission)
	_, err := s.client.Put(path, nil)
	return err
}
