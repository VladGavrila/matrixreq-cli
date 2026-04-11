package service

import (
	"encoding/json"
	"fmt"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// JiraService manages Jira-issue links on Matrix items via the /rest/2/wfgw/
// endpoint. This is separate from Matrix's internal uplinks/downlinks, which
// live on ItemService.
type JiraService interface {
	// GetLinks returns the Jira issues linked to a Matrix item.
	// Returns an empty slice (not nil) when there are no links.
	GetLinks(project, itemRef string) ([]api.JiraLinkedIssue, error)

	// CreateLinks links one or more Jira issues to a single Matrix item.
	// Idempotent: linking the same pair twice is a no-op on the server.
	CreateLinks(project, itemRef string, externals []api.JiraExternalItem) error

	// BreakLinks unlinks one or more Jira issues from a single Matrix item.
	BreakLinks(project, itemRef string, externals []api.JiraExternalItem, pluginID int) error
}

type jiraService struct {
	client *client.Client
}

func (s *jiraService) GetLinks(project, itemRef string) ([]api.JiraLinkedIssue, error) {
	payload := api.JiraGetIssuesPayload{
		Action: "GetIssues",
		MatrixItem: api.JiraMatrixItem{
			Project:    project,
			MatrixItem: itemRef,
		},
	}
	data, err := s.client.WfgwGet(payload)
	if err != nil {
		return nil, err
	}
	var entries []api.JiraGetIssuesResponseEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing wfgw GetIssues response: %w", err)
	}
	if len(entries) == 0 {
		return []api.JiraLinkedIssue{}, nil
	}
	if entries[0].Links == nil {
		return []api.JiraLinkedIssue{}, nil
	}
	return entries[0].Links, nil
}

func (s *jiraService) CreateLinks(project, itemRef string, externals []api.JiraExternalItem) error {
	payload := api.JiraCreateLinksPayload{
		Action: "CreateLinks",
		MatrixItem: api.JiraMatrixItem{
			Project:    project,
			MatrixItem: itemRef,
		},
		ExternalItems: externals,
	}
	_, err := s.client.WfgwPostForm(payload)
	return err
}

func (s *jiraService) BreakLinks(project, itemRef string, externals []api.JiraExternalItem, pluginID int) error {
	// BreakLinks requires matrixItemIds on each external — default to [itemRef].
	normalized := make([]api.JiraExternalItem, len(externals))
	for i, ex := range externals {
		if len(ex.MatrixItemIDs) == 0 {
			ex.MatrixItemIDs = []string{itemRef}
		}
		normalized[i] = ex
	}
	body := api.JiraBreakLinksBody{
		PluginID: pluginID,
		Action:   "BreakLinks",
		MatrixItem: api.JiraMatrixItem{
			Project:    project,
			MatrixItem: itemRef,
		},
		ExternalItems: normalized,
	}
	_, err := s.client.WfgwDeleteJSON(body)
	return err
}
