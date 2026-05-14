package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/client"
)

// TestItemUpdateSendsFormBody pins the wire format for item update: parameters
// must travel in an application/x-www-form-urlencoded body, not the URL query
// string. Sending a large Test Case Steps JSON via the query string is what
// causes HTTP 414 on real-world XTC payloads.
func TestItemUpdateSendsFormBody(t *testing.T) {
	largeSteps := `[{"RequirementLink":"REQ-1","result":"p","comment":"` + strings.Repeat("ok ", 4000) + `"}]`

	var got struct {
		method      string
		path        string
		query       string
		contentType string
		formFields  map[string]string
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.RawQuery
		got.contentType = r.Header.Get("Content-Type")
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		got.formFields = make(map[string]string, len(r.PostForm))
		for k, v := range r.PostForm {
			if len(v) > 0 {
				got.formFields[k] = v[0]
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	svc := New(client.New(srv.URL, "tok", false))

	req := &api.UpdateItemRequest{
		Title:  "Title",
		Reason: "smoke",
		Fields: []api.FieldValSetType{
			{ID: 12706, Value: largeSteps},
			{ID: 12705, Value: "p"},
		},
	}
	if _, err := svc.Items.Update("PROJ", "XTC-1", req); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got.method != http.MethodPut {
		t.Errorf("method = %q, want PUT", got.method)
	}
	if got.path != "/rest/1/PROJ/item/XTC-1" {
		t.Errorf("path = %q, want /rest/1/PROJ/item/XTC-1", got.path)
	}
	if got.query != "" {
		t.Errorf("query must be empty (payload belongs in body), got %q", got.query)
	}
	if got.contentType != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q, want application/x-www-form-urlencoded", got.contentType)
	}
	if got.formFields["reason"] != "smoke" {
		t.Errorf("reason = %q, want %q", got.formFields["reason"], "smoke")
	}
	if got.formFields["title"] != "Title" {
		t.Errorf("title = %q, want %q", got.formFields["title"], "Title")
	}
	if got.formFields["fx12705"] != "p" {
		t.Errorf("fx12705 = %q, want %q", got.formFields["fx12705"], "p")
	}
	if got.formFields["fx12706"] != largeSteps {
		t.Errorf("fx12706 roundtrip mismatch (len got=%d want=%d)", len(got.formFields["fx12706"]), len(largeSteps))
	}
}

// TestItemCreateSendsFormBody pins the same contract for create.
func TestItemCreateSendsFormBody(t *testing.T) {
	var got struct {
		method      string
		path        string
		query       string
		contentType string
		formFields  map[string]string
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.RawQuery
		got.contentType = r.Header.Get("Content-Type")
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		got.formFields = make(map[string]string, len(r.PostForm))
		for k, v := range r.PostForm {
			if len(v) > 0 {
				got.formFields[k] = v[0]
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"itemId":1,"serial":1}`))
	}))
	defer srv.Close()

	svc := New(client.New(srv.URL, "tok", false))

	req := &api.CreateItemRequest{
		Title:  "New Item",
		Folder: "F-REQ-1",
		Reason: "smoke",
		Fields: []api.FieldValSetType{
			{ID: 100, Value: "hello"},
		},
	}
	if _, err := svc.Items.Create("PROJ", req); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.method != http.MethodPost {
		t.Errorf("method = %q, want POST", got.method)
	}
	if got.path != "/rest/1/PROJ/item" {
		t.Errorf("path = %q, want /rest/1/PROJ/item", got.path)
	}
	if got.query != "" {
		t.Errorf("query must be empty, got %q", got.query)
	}
	if got.contentType != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q, want application/x-www-form-urlencoded", got.contentType)
	}
	if got.formFields["title"] != "New Item" {
		t.Errorf("title = %q", got.formFields["title"])
	}
	if got.formFields["folder"] != "F-REQ-1" {
		t.Errorf("folder = %q", got.formFields["folder"])
	}
	if got.formFields["reason"] != "smoke" {
		t.Errorf("reason = %q", got.formFields["reason"])
	}
	if got.formFields["fx100"] != "hello" {
		t.Errorf("fx100 = %q", got.formFields["fx100"])
	}
}
