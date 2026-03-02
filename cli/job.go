package cli

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(jobCmd)
	jobCmd.AddCommand(jobListCmd)
	jobCmd.AddCommand(jobGetCmd)
	jobCmd.AddCommand(jobCancelCmd)
	jobCmd.AddCommand(jobDownloadCmd)
}

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage async jobs",
}

var jobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all jobs in a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		jobs, err := svc.Jobs.List(project)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), jobs)
		}
		headers := []string{"ID", "Status", "Progress", "Name"}
		var rows [][]string
		for _, j := range jobs {
			rows = append(rows, []string{
				strconv.Itoa(j.JobID), j.Status, strconv.Itoa(j.Progress), j.VisibleName,
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

var jobGetCmd = &cobra.Command{
	Use:   "get <jobID>",
	Short: "Get job status",
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
		jobID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid job ID: %s", args[0])
		}
		job, err := svc.Jobs.Get(project, jobID)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), job)
		}
		fmt.Printf("Job ID:   %d\n", job.JobID)
		fmt.Printf("Status:   %s\n", job.Status)
		fmt.Printf("Progress: %d%%\n", job.Progress)
		if job.JobFileURL != "" {
			fmt.Printf("File URL: %s\n", job.JobFileURL)
		}
		return nil
	},
}

var jobCancelCmd = &cobra.Command{
	Use:   "cancel <jobID>",
	Short: "Cancel a job",
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
		jobID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid job ID: %s", args[0])
		}
		reason, _ := cmd.Flags().GetString("reason")
		if err := svc.Jobs.Cancel(project, jobID, reason); err != nil {
			return err
		}
		fmt.Printf("Job %d cancelled.\n", jobID)
		return nil
	},
}

func init() {
	jobCancelCmd.Flags().StringP("reason", "r", "", "Reason for cancellation")
	_ = jobCancelCmd.MarkFlagRequired("reason")
}

var jobDownloadCmd = &cobra.Command{
	Use:   "download <jobID> <fileNo>",
	Short: "Download a job file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		jobID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid job ID: %s", args[0])
		}
		fileNo, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid file number: %s", args[1])
		}
		outFile, _ := cmd.Flags().GetString("out")
		resp, err := svc.Jobs.Download(project, jobID, fileNo)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var w io.Writer = os.Stdout
		if outFile != "" {
			f, err := os.Create(outFile)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer f.Close()
			w = f
		}
		n, err := io.Copy(w, resp.Body)
		if err != nil {
			return fmt.Errorf("writing file: %w", err)
		}
		if outFile != "" {
			fmt.Fprintf(os.Stderr, "Downloaded %d bytes to %s\n", n, outFile)
		}
		return nil
	},
}

func init() {
	jobDownloadCmd.Flags().String("out", "", "Output file path (default: stdout)")
}
