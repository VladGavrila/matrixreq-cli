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
	rootCmd.AddCommand(fileCmd)
	fileCmd.AddCommand(fileListCmd)
	fileCmd.AddCommand(fileGetCmd)
	fileCmd.AddCommand(fileUploadCmd)
}

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage project files",
}

var fileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all files in a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		files, err := svc.Files.List(project)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), files)
		}
		headers := []string{"ID", "Name", "MimeType", "Path"}
		var rows [][]string
		for _, f := range files {
			rows = append(rows, []string{
				strconv.Itoa(f.FileID), f.LocalName, f.MimeType, f.FullPath,
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

var fileGetCmd = &cobra.Command{
	Use:   "get <fileNo> <key>",
	Short: "Download a file",
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
		fileNo, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid file number: %s", args[0])
		}
		outFile, _ := cmd.Flags().GetString("out")
		resp, err := svc.Files.Get(project, fileNo, args[1])
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
	fileGetCmd.Flags().String("out", "", "Output file path (default: stdout)")
}

var fileUploadCmd = &cobra.Command{
	Use:   "upload <filePath>",
	Short: "Upload a file to the project",
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
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("stat file: %w", err)
		}

		ack, err := svc.Files.Upload(project, info.Name(), f)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		fmt.Printf("Uploaded: fileId=%d, path=%s\n", ack.FileID, ack.FileFullPath)
		return nil
	},
}
