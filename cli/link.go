package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(linkCmd)
	linkCmd.AddCommand(linkCreateCmd)
	linkCmd.AddCommand(linkDeleteCmd)
}

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Manage traceability links",
}

var linkCreateCmd = &cobra.Command{
	Use:   "create <upItem> <downItem>",
	Short: "Create a link between two items",
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
		reason, _ := cmd.Flags().GetString("reason")
		if err := svc.Items.CreateLink(project, args[0], args[1], reason); err != nil {
			return err
		}
		fmt.Printf("Link created: %s -> %s\n", args[0], args[1])
		return nil
	},
}

var linkDeleteCmd = &cobra.Command{
	Use:   "delete <upItem> <downItem>",
	Short: "Delete a link between two items",
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
		reason, _ := cmd.Flags().GetString("reason")
		if err := svc.Items.DeleteLink(project, args[0], args[1], reason); err != nil {
			return err
		}
		fmt.Printf("Link deleted: %s -> %s\n", args[0], args[1])
		return nil
	},
}

func init() {
	linkCreateCmd.Flags().StringP("reason", "r", "", "Reason for link creation")
	_ = linkCreateCmd.MarkFlagRequired("reason")
	linkDeleteCmd.Flags().StringP("reason", "r", "", "Reason for link deletion")
	_ = linkDeleteCmd.MarkFlagRequired("reason")
}
