package service

import (
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// FieldService manages field operations.
type FieldService interface {
	Get(project, itemRef, fieldName string) (string, error)
	Update(project string, fieldID int, label, fieldParam, reason string, order int) error
	Delete(project, category string, fieldID int, reason string) error
	AddToCategory(project, category, label, fieldType, fieldParam, reason string) error
}

type fieldService struct {
	client *client.Client
}

func (s *fieldService) Get(project, itemRef, fieldName string) (string, error) {
	path := fmt.Sprintf("/%s/field/%s?field=%s",
		url.PathEscape(project), url.PathEscape(itemRef), url.QueryEscape(fieldName))
	data, err := s.client.Get(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *fieldService) Update(project string, fieldID int, label, fieldParam, reason string, order int) error {
	path := fmt.Sprintf("/%s/field?field=%d&label=%s&fieldParam=%s&reason=%s&order=%d",
		url.PathEscape(project), fieldID,
		url.QueryEscape(label), url.QueryEscape(fieldParam), url.QueryEscape(reason), order)
	_, err := s.client.Put(path, nil)
	return err
}

func (s *fieldService) Delete(project, category string, fieldID int, reason string) error {
	path := fmt.Sprintf("/%s/field/%s?field=%d&reason=%s",
		url.PathEscape(project), url.PathEscape(category), fieldID, url.QueryEscape(reason))
	_, err := s.client.Delete(path)
	return err
}

func (s *fieldService) AddToCategory(project, category, label, fieldType, fieldParam, reason string) error {
	path := fmt.Sprintf("/%s/cat?label=%s&category=%s&fieldType=%s&fieldParam=%s&reason=%s",
		url.PathEscape(project),
		url.QueryEscape(label), url.QueryEscape(category),
		url.QueryEscape(fieldType), url.QueryEscape(fieldParam), url.QueryEscape(reason))
	_, err := s.client.Post(path, nil)
	return err
}
