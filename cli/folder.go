package cli

import (
	"fmt"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(folderCmd)
	folderCmd.AddCommand(folderGetCmd)
	folderCmd.AddCommand(folderCreateCmd)
	folderCmd.AddCommand(folderUpdateCmd)
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

var folderUpdateCmd = &cobra.Command{
	Use:   "update <folderRef>",
	Short: "Update a folder (title, fields)",
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
		title, _ := cmd.Flags().GetString("title")
		reason, _ := cmd.Flags().GetString("reason")
		if reason == "" {
			reason = "Updated by mxreq CLI"
		}
		description, _ := cmd.Flags().GetString("description")
		fieldFlags, _ := cmd.Flags().GetStringArray("field")

		ref := upperRef(args[0])

		req := &api.UpdateItemRequest{
			Title:  title,
			Reason: reason,
		}

		if description != "" || len(fieldFlags) > 0 {
			fm, err := fieldmap.LoadOrFetch(svc, project)
			if err != nil {
				return fmt.Errorf("loading field map: %w", err)
			}
			if description != "" {
				id, err := fm.DescriptionField("FOLDER")
				if err != nil {
					return err
				}
				req.Fields = append(req.Fields, api.FieldValSetType{ID: id, Value: description})
			}
			for _, f := range fieldFlags {
				parts := strings.SplitN(f, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid --field format %q, expected fieldName=value", f)
				}
				id, err := fm.Resolve("FOLDER", parts[0])
				if err != nil {
					return err
				}
				req.Fields = append(req.Fields, api.FieldValSetType{ID: id, Value: parts[1]})
			}
		}

		item, err := svc.Items.Update(project, ref, req)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), item)
		}
		fmt.Printf("Updated %s: %s (v%d)\n", item.ItemRef, item.Title, item.MaxVersion)
		return nil
	},
}

func init() {
	folderUpdateCmd.Flags().String("title", "", "New folder title")
	folderUpdateCmd.Flags().StringP("description", "d", "", "Folder description (auto-resolves the field)")
	folderUpdateCmd.Flags().StringArrayP("field", "f", nil, "Set field value (format: fieldName=value, repeatable)")
	folderUpdateCmd.Flags().StringP("reason", "r", "", "Reason for update (default: Updated by mxreq CLI)")
}
