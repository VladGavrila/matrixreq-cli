package cli

import (
	"fmt"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(folderCmd)
	folderCmd.AddCommand(folderGetCmd)
	folderCmd.AddCommand(folderCreateCmd)
}

var folderCmd = &cobra.Command{
	Use:   "folder",
	Short: "Manage folders",
}

var folderGetCmd = &cobra.Command{
	Use:   "get <folderRef>",
	Short: "Get folder details (e.g., F-REQ-1)",
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
		history, _ := cmd.Flags().GetBool("history")
		folder, err := svc.Items.GetFolder(project, upperRef(args[0]), history)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), folder)
		}
		fmt.Printf("Folder: %s\n", folder.ItemRef)
		fmt.Printf("Title:  %s\n", folder.Title)
		if len(folder.ItemList) > 0 {
			fmt.Printf("Items:  %d\n", len(folder.ItemList))
			for _, item := range folder.ItemList {
				typ := "item"
				if item.IsFolder == 1 {
					typ = "folder"
				}
				fmt.Printf("  %s [%s] %s\n", item.ItemRef, typ, item.Title)
			}
		}
		return nil
	},
}

func init() {
	folderGetCmd.Flags().Bool("history", false, "Include version history")
}

var folderCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new folder",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		parent, _ := cmd.Flags().GetString("parent")
		label, _ := cmd.Flags().GetString("label")
		reason, _ := cmd.Flags().GetString("reason")

		ack, err := svc.Items.CreateFolder(project, upperRef(parent), label, reason)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		category := categoryFromRef(upperRef(parent))
		fmt.Printf("Created folder F-%s-%d\n", category, ack.Serial)
		return nil
	},
}

func init() {
	folderCreateCmd.Flags().String("parent", "", "Parent folder ref (e.g., F-REQ-1)")
	folderCreateCmd.Flags().String("label", "", "Folder label")
	folderCreateCmd.Flags().StringP("reason", "r", "", "Reason for creation")
	_ = folderCreateCmd.MarkFlagRequired("parent")
	_ = folderCreateCmd.MarkFlagRequired("label")
	_ = folderCreateCmd.MarkFlagRequired("reason")
}
