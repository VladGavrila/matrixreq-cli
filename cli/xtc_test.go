package cli

import (
	"strings"
	"testing"

	"github.com/VladGavrila/matrixreq-cli/internal/execution"
)

func TestStrictUploadError(t *testing.T) {
	newResult := func() *execution.UploadResult {
		return &execution.UploadResult{
			Successes: map[string]bool{"XTC-1": true},
			Skipped:   map[string]string{},
			Failed:    map[string]string{},
			Missing:   []string{},
		}
	}

	tests := []struct {
		name     string
		mutate   func(*execution.UploadResult)
		strict   bool
		wantErr  bool
		wantText []string
	}{
		{"clean, strict", func(*execution.UploadResult) {}, true, false, nil},
		{"skipped, strict", func(u *execution.UploadResult) { u.Skipped["XTC-2"] = "drift" }, true, true, []string{"XTC-2", "out of sync"}},
		{"failed update, strict", func(u *execution.UploadResult) { u.Failed["XTC-3"] = "boom" }, true, true, []string{"XTC-3", "failed"}},
		{"missing test, strict", func(u *execution.UploadResult) { u.Missing = []string{"TC-9"} }, true, true, []string{"TC-9", "not found"}},
		{"all three, not strict", func(u *execution.UploadResult) {
			u.Skipped["XTC-2"] = "drift"
			u.Failed["XTC-3"] = "boom"
			u.Missing = []string{"TC-9"}
		}, false, false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newResult()
			tt.mutate(u)
			err := strictUploadError(u, tt.strict)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			for _, want := range tt.wantText {
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Errorf("error %v missing %q", err, want)
				}
			}
		})
	}
}
