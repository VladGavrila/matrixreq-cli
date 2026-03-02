package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// ItemService manages item operations.
type ItemService interface {
	Get(project, itemRef string, history bool) (*api.TrimItem, error)
	GetFolder(project, folderRef string, history bool) (*api.TrimFolder, error)
	Create(project string, req *api.CreateItemRequest) (*api.AddItemAck, error)
	Update(project, itemRef string, req *api.UpdateItemRequest) (*api.TrimItem, error)
	Delete(project, itemRef, reason string) error
	Restore(project, itemRef, reason string) (*api.UndeleteAnswer, error)
	Copy(project, itemOrFolder, targetFolder, reason string, copyLabels int) (*api.CopyItemAck, error)
	Move(project, folder, items, reason string) error
	CreateFolder(project, parent, label, reason string) (*api.AddItemAck, error)
	Touch(project, itemRef, reason string) error
	CreateLink(project, upItem, downItem, reason string) error
	DeleteLink(project, upItem, downItem, reason string) error
}

type itemService struct {
	client *client.Client
}

func (s *itemService) Get(project, itemRef string, history bool) (*api.TrimItem, error) {
	path := fmt.Sprintf("/%s/item/%s", url.PathEscape(project), url.PathEscape(itemRef))
	if history {
		path += "?history=1"
	}
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp api.TrimItem
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing item: %w", err)
	}
	return &resp, nil
}

func (s *itemService) GetFolder(project, folderRef string, history bool) (*api.TrimFolder, error) {
	path := fmt.Sprintf("/%s/item/%s", url.PathEscape(project), url.PathEscape(folderRef))
	if history {
		path += "?history=1"
	}
	data, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var resp api.TrimFolder
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing folder: %w", err)
	}
	return &resp, nil
}

func (s *itemService) Create(project string, req *api.CreateItemRequest) (*api.AddItemAck, error) {
	data, err := s.client.Post("/"+url.PathEscape(project)+"/item", req)
	if err != nil {
		return nil, err
	}
	var resp api.AddItemAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create item response: %w", err)
	}
	return &resp, nil
}

func (s *itemService) Update(project, itemRef string, req *api.UpdateItemRequest) (*api.TrimItem, error) {
	path := fmt.Sprintf("/%s/item/%s", url.PathEscape(project), url.PathEscape(itemRef))
	data, err := s.client.Put(path, req)
	if err != nil {
		return nil, err
	}
	var resp api.TrimItem
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing update item response: %w", err)
	}
	return &resp, nil
}

func (s *itemService) Delete(project, itemRef, reason string) error {
	path := fmt.Sprintf("/%s/item/%s?confirm=yes&reason=%s",
		url.PathEscape(project), url.PathEscape(itemRef), url.QueryEscape(reason))
	_, err := s.client.Delete(path)
	return err
}

func (s *itemService) Restore(project, itemRef, reason string) (*api.UndeleteAnswer, error) {
	path := fmt.Sprintf("/%s/item/%s?reason=%s",
		url.PathEscape(project), url.PathEscape(itemRef), url.QueryEscape(reason))
	data, err := s.client.Post(path, nil)
	if err != nil {
		return nil, err
	}
	var resp api.UndeleteAnswer
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing restore response: %w", err)
	}
	return &resp, nil
}

func (s *itemService) Copy(project, itemOrFolder, targetFolder, reason string, copyLabels int) (*api.CopyItemAck, error) {
	path := fmt.Sprintf("/%s/copy/%s?targetFolder=%s&reason=%s&copyLabels=%d",
		url.PathEscape(project), url.PathEscape(itemOrFolder),
		url.QueryEscape(targetFolder), url.QueryEscape(reason), copyLabels)
	data, err := s.client.Post(path, nil)
	if err != nil {
		return nil, err
	}
	var resp api.CopyItemAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing copy response: %w", err)
	}
	return &resp, nil
}

func (s *itemService) Move(project, folder, items, reason string) error {
	path := fmt.Sprintf("/%s/movein/%s?items=%s&reason=%s",
		url.PathEscape(project), url.PathEscape(folder),
		url.QueryEscape(items), url.QueryEscape(reason))
	_, err := s.client.Post(path, nil)
	return err
}

func (s *itemService) CreateFolder(project, parent, label, reason string) (*api.AddItemAck, error) {
	path := fmt.Sprintf("/%s/folder?parent=%s&label=%s&reason=%s",
		url.PathEscape(project),
		url.QueryEscape(parent), url.QueryEscape(label), url.QueryEscape(reason))
	data, err := s.client.Post(path, nil)
	if err != nil {
		return nil, err
	}
	var resp api.AddItemAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create folder response: %w", err)
	}
	return &resp, nil
}

func (s *itemService) Touch(project, itemRef, reason string) error {
	path := fmt.Sprintf("/%s/touch/%s?reason=%s",
		url.PathEscape(project), url.PathEscape(itemRef), url.QueryEscape(reason))
	_, err := s.client.Put(path, nil)
	return err
}

func (s *itemService) CreateLink(project, upItem, downItem, reason string) error {
	path := fmt.Sprintf("/%s/itemlink/%s/%s?reason=%s",
		url.PathEscape(project), url.PathEscape(upItem),
		url.PathEscape(downItem), url.QueryEscape(reason))
	_, err := s.client.Post(path, nil)
	return err
}

func (s *itemService) DeleteLink(project, upItem, downItem, reason string) error {
	path := fmt.Sprintf("/%s/itemlink/%s/%s?reason=%s",
		url.PathEscape(project), url.PathEscape(upItem),
		url.PathEscape(downItem), url.QueryEscape(reason))
	_, err := s.client.Delete(path)
	return err
}
