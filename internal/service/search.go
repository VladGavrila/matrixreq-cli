package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// SearchService manages search operations.
type SearchService interface {
	Search(project string, needle *api.TrimNeedle) ([]api.TrimNeedleItem, error)
	SearchMinimal(project, search, filter string) ([]byte, error)
}

// TrimNeedleResponse wraps the search result.
type TrimNeedleResponse struct {
	NeedleResponse []api.TrimNeedleItem `json:"needles,omitempty"`
}

type searchService struct {
	client *client.Client
}

func (s *searchService) Search(project string, needle *api.TrimNeedle) ([]api.TrimNeedleItem, error) {
	path := fmt.Sprintf("/%s/needle", url.PathEscape(project))
	data, err := s.client.Post(path, needle)
	if err != nil {
		return nil, err
	}
	var resp TrimNeedleResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing search results: %w", err)
	}
	return resp.NeedleResponse, nil
}

func (s *searchService) SearchMinimal(project, search, filter string) ([]byte, error) {
	path := fmt.Sprintf("/%s/needleminimal?search=%s",
		url.PathEscape(project), url.QueryEscape(search))
	if filter != "" {
		path += "&filter=" + url.QueryEscape(filter)
	}
	return s.client.Get(path)
}
