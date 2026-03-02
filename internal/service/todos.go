package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// TodoService manages todo operations.
type TodoService interface {
	List(project string, includeDone, includeFuture bool) ([]api.Todo, error)
	ListAll(includeDone, includeFuture bool) ([]api.Todo, error)
	Create(project, item, text, todoType string, fieldID int, logins string) error
	Done(project string, todoID int, hardDelete bool) error
}

// GetTodosAck wraps the todo list response.
type GetTodosAck struct {
	Todos []api.Todo `json:"todos,omitempty"`
}

type todoService struct {
	client *client.Client
}

func (s *todoService) List(project string, includeDone, includeFuture bool) ([]api.Todo, error) {
	path := fmt.Sprintf("/%s/todo?includeDone=%s&includeFuture=%s",
		url.PathEscape(project), boolToStr(includeDone), boolToStr(includeFuture))
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp GetTodosAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing todos: %w", err)
	}
	return resp.Todos, nil
}

func (s *todoService) ListAll(includeDone, includeFuture bool) ([]api.Todo, error) {
	path := fmt.Sprintf("/all/todo?includeDone=%s&includeFuture=%s",
		boolToStr(includeDone), boolToStr(includeFuture))
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp GetTodosAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing todos: %w", err)
	}
	return resp.Todos, nil
}

func (s *todoService) Create(project, item, text, todoType string, fieldID int, logins string) error {
	path := fmt.Sprintf("/%s/todo/%s?text=%s",
		url.PathEscape(project), url.PathEscape(item), url.QueryEscape(text))
	if todoType != "" {
		path += "&todoType=" + url.QueryEscape(todoType)
	}
	if fieldID > 0 {
		path += fmt.Sprintf("&fieldId=%d", fieldID)
	}
	if logins != "" {
		path += "&logins=" + url.QueryEscape(logins)
	}
	_, err := s.client.Post(path, nil)
	return err
}

func (s *todoService) Done(project string, todoID int, hardDelete bool) error {
	hd := "no"
	if hardDelete {
		hd = "yes"
	}
	path := fmt.Sprintf("/%s/todo/%d?hardDelete=%s",
		url.PathEscape(project), todoID, hd)
	_, err := s.client.Delete(path)
	return err
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
