package cli

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export <itemList>",
	Short: "Export items (returns job ID)",
	Long:  "Export items from a project. <itemList> is a comma-separated list of item refs.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/%s/export?itemList=%s",
			url.PathEscape(project), url.QueryEscape(args[0]))
		data, err := svc.Client.Get(path)
		if err != nil {
			return err
		}
		var ack api.ExportItemsAck
		if err := json.Unmarshal(data, &ack); err != nil {
			return fmt.Errorf("parsing export response: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		fmt.Printf("Export job created: jobId=%d\n", ack.JobID)
		fmt.Printf("Use 'mxreq job get %d' to check status.\n", ack.JobID)
		return nil
	},
}
