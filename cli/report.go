package cli

import (
	"fmt"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.AddCommand(reportGenerateCmd)
	reportCmd.AddCommand(reportSignedCmd)
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports",
}

var reportGenerateCmd = &cobra.Command{
	Use:   "generate <report>",
	Short: "Generate a report (returns job ID)",
	Long:  "Generate a report. <report> can be REPORT-n, DOC-n, SIGN-n, or a report name.",
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
		format, _ := cmd.Flags().GetString("format")
		signed, _ := cmd.Flags().GetBool("signed")
		ack, err := svc.Reports.Generate(project, args[0], format, signed)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		fmt.Printf("Report job created: jobId=%d\n", ack.JobID)
		fmt.Printf("Use 'mxreq job get %d' to check status.\n", ack.JobID)
		return nil
	},
}

func init() {
	reportGenerateCmd.Flags().String("format", "docx", "Output format (docx, pdf, etc.)")
	reportGenerateCmd.Flags().Bool("signed", false, "Generate as signed report")
}

var reportSignedCmd = &cobra.Command{
	Use:   "signed <signItem>",
	Short: "Generate a signed report (e.g., SIGN-1)",
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
		format, _ := cmd.Flags().GetString("format")
		data, err := svc.Reports.GenerateSigned(project, args[0], format)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	reportSignedCmd.Flags().String("format", "docx", "Output format")
}
