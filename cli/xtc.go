package cli

import (
	"fmt"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/execution"
	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

var xtcCmd = &cobra.Command{
	Use:   "xtc",
	Short: "Executed Test Case (XTC) operations",
}

var xtcUploadCmd = &cobra.Command{
	Use:   "upload <results.yaml>",
	Short: "Upload execution results to XTCs",
	Long: `Parse a YAML results file and upload execution results to the corresponding
XTCs in the specified folder. The YAML file should contain test results with
step-level pass/fail status matched by requirement links.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}

		folder, _ := cmd.Flags().GetString("folder")

		fm, err := fieldmap.LoadOrFetch(svc, project)
		if err != nil {
			return fmt.Errorf("loading field map: %w", err)
		}

		results, err := execution.ParseYAMLResults(args[0])
		if err != nil {
			return fmt.Errorf("parsing results: %w", err)
		}

		uploadResult, err := execution.UploadResults(svc, project, folder, results, fm)
		if err != nil {
			return err
		}

		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), uploadResult)
		}

		// Print summary
		successCount := 0
		for ref, ok := range uploadResult.Successes {
			if ok {
				successCount++
				fmt.Printf("  Updated %s\n", ref)
			}
		}

		if len(uploadResult.Issues) > 0 {
			fmt.Println("Issues:")
			for _, issue := range uploadResult.Issues {
				fmt.Printf("  %s\n", issue)
			}
		}

		fmt.Printf("\nUploaded %d/%d XTCs successfully\n", successCount, len(uploadResult.Successes))

		return nil
	},
}

var xtcStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Compute execution statistics for an XTC folder",
	Long: `Fetch all XTCs in the specified folder and compute execution statistics
including per-test status, step counts, requirement coverage, and aggregated totals.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}

		folder, _ := cmd.Flags().GetString("folder")

		fm, err := fieldmap.LoadOrFetch(svc, project)
		if err != nil {
			return fmt.Errorf("loading field map: %w", err)
		}

		stats, err := execution.ComputeStats(svc, project, folder, fm)
		if err != nil {
			return err
		}

		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), stats.ToDict())
		}

		// Print summary table
		fmt.Println("XTC Execution Statistics")
		fmt.Println(strings.Repeat("-", 60))

		// Per-test stats
		if len(stats.XTCStats) > 0 {
			headers := []string{"XTC", "Status", "Steps", "Passed", "Failed", "Not Exec"}
			var rows [][]string
			for ref, s := range stats.XTCStats {
				rows = append(rows, []string{
					ref,
					s.TestStatus,
					fmt.Sprintf("%d", s.NumSteps),
					fmt.Sprintf("%d", s.NumPassed),
					fmt.Sprintf("%d", s.NumFailed),
					fmt.Sprintf("%d", s.NumNotExecutedWithReq),
				})
			}
			if err := output.Print(getOutputFormat(), headers, rows); err != nil {
				return err
			}
		}

		// Totals
		fmt.Printf("\nTotals:\n")
		fmt.Printf("  Tests executed:     %d\n", stats.TotalTestsExecuted)
		fmt.Printf("  Tests in progress:  %d\n", stats.TotalTestsInProgress)
		fmt.Printf("  Tests not executed: %d\n", stats.TotalTestsNotExecuted)
		fmt.Printf("  Total steps:        %d\n", stats.TotalSteps)
		fmt.Printf("  Steps passed:       %d\n", stats.TotalPassed)
		fmt.Printf("  Steps failed:       %d\n", stats.TotalFailed)

		// SOFT coverage summary
		if len(stats.OverallSOFTCoverage) > 0 {
			fmt.Printf("\nRequirement Coverage:\n")
			for _, result := range []string{"passed", "failed", "pass with issue", "Not Executed"} {
				count := stats.ExecutedSOFTTotals[result]
				if count > 0 {
					fmt.Printf("  %s: %d\n", result, count)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(xtcCmd)
	xtcCmd.AddCommand(xtcUploadCmd)
	xtcCmd.AddCommand(xtcStatsCmd)

	xtcUploadCmd.Flags().StringP("folder", "f", "", "Target XTC folder (e.g., F-XTC-123)")
	_ = xtcUploadCmd.MarkFlagRequired("folder")

	xtcStatsCmd.Flags().StringP("folder", "f", "", "Target XTC folder (e.g., F-XTC-123)")
	_ = xtcStatsCmd.MarkFlagRequired("folder")
}
