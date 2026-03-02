package service

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// ReportService manages report generation.
type ReportService interface {
	Generate(project, report, format string, isSignedReport bool) (*api.CreateReportJobAck, error)
	GenerateSigned(project, signItem, format string) ([]byte, error)
}

type reportService struct {
	client *client.Client
}

func (s *reportService) Generate(project, report, format string, isSignedReport bool) (*api.CreateReportJobAck, error) {
	signed := "no"
	if isSignedReport {
		signed = "yes"
	}
	path := fmt.Sprintf("/%s/report/%s?format=%s&isSignedReport=%s&includeSignatures=yes&newTitle=&copyFields=",
		url.PathEscape(project), url.PathEscape(report),
		url.QueryEscape(format), signed)
	data, err := s.client.Post(path, nil)
	if err != nil {
		return nil, err
	}
	var resp api.CreateReportJobAck
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing report response: %w", err)
	}
	return &resp, nil
}

func (s *reportService) GenerateSigned(project, signItem, format string) ([]byte, error) {
	path := fmt.Sprintf("/%s/signedreport/%s",
		url.PathEscape(project), url.PathEscape(signItem))
	if format != "" {
		path += "?format=" + url.QueryEscape(format)
	}
	return s.client.Post(path, nil)
}
