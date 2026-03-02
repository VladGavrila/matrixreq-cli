package cli

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import items into a project from an XML file",
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
		reason, _ := cmd.Flags().GetString("reason")

		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return fmt.Errorf("stat file: %w", err)
		}

		path := fmt.Sprintf("/%s/import?reason=%s",
			url.PathEscape(project), url.QueryEscape(reason))
		data, err := svc.Client.PostForm(path, nil, info.Name(), f)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	importCmd.Flags().StringP("reason", "r", "", "Reason for import")
	_ = importCmd.MarkFlagRequired("reason")
}
