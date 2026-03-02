package cli

import (
	"fmt"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(itemCmd)
	itemCmd.AddCommand(itemGetCmd)
	itemCmd.AddCommand(itemCreateCmd)
	itemCmd.AddCommand(itemUpdateCmd)
	itemCmd.AddCommand(itemDeleteCmd)
	itemCmd.AddCommand(itemRestoreCmd)
	itemCmd.AddCommand(itemCopyCmd)
	itemCmd.AddCommand(itemMoveCmd)
	itemCmd.AddCommand(itemTouchCmd)
}

var itemCmd = &cobra.Command{
	Use:   "item",
	Short: "Manage items",
}

var itemGetCmd = &cobra.Command{
	Use:   "get <itemRef>",
	Short: "Get item details (e.g., REQ-1, F-REQ-1)",
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
		item, err := svc.Items.Get(project, args[0], history)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), item)
		}
		fmt.Printf("Item:     %s\n", item.ItemRef)
		fmt.Printf("Title:    %s\n", item.Title)
		fmt.Printf("Version:  %d\n", item.MaxVersion)
		fmt.Printf("Modified: %s\n", item.ModDateUserFormat)
		if item.FieldValList != nil {
			fmt.Println("\nFields:")
			for _, f := range item.FieldValList.FieldVal {
				val := f.Value
				if len(val) > 80 {
					val = val[:80] + "..."
				}
				fmt.Printf("  [%d] %s (%s): %s\n", f.ID, f.FieldName, f.FieldType, val)
			}
		}
		if len(item.UpLinkList) > 0 {
			fmt.Println("\nUp Links:")
			for _, l := range item.UpLinkList {
				fmt.Printf("  -> %s: %s\n", l.ItemRef, l.Title)
			}
		}
		if len(item.DownLinkList) > 0 {
			fmt.Println("\nDown Links:")
			for _, l := range item.DownLinkList {
				fmt.Printf("  <- %s: %s\n", l.ItemRef, l.Title)
			}
		}
		return nil
	},
}

func init() {
	itemGetCmd.Flags().Bool("history", false, "Include version history")
}

var itemCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new item",
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
		folder, _ := cmd.Flags().GetString("folder")
		reason, _ := cmd.Flags().GetString("reason")

		req := &api.CreateItemRequest{
			Title:  title,
			Folder: folder,
			Reason: reason,
		}
		ack, err := svc.Items.Create(project, req)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		fmt.Printf("Created item ID=%d, serial=%d\n", ack.ItemID, ack.Serial)
		return nil
	},
}

func init() {
	itemCreateCmd.Flags().String("title", "", "Item title")
	itemCreateCmd.Flags().String("folder", "", "Parent folder (e.g., F-REQ-1)")
	itemCreateCmd.Flags().StringP("reason", "r", "", "Reason for creation")
	_ = itemCreateCmd.MarkFlagRequired("title")
	_ = itemCreateCmd.MarkFlagRequired("folder")
	_ = itemCreateCmd.MarkFlagRequired("reason")
}

var itemUpdateCmd = &cobra.Command{
	Use:   "update <itemRef>",
	Short: "Update an item",
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

		req := &api.UpdateItemRequest{
			Title:  title,
			Reason: reason,
		}
		item, err := svc.Items.Update(project, args[0], req)
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
	itemUpdateCmd.Flags().String("title", "", "New title")
	itemUpdateCmd.Flags().StringP("reason", "r", "", "Reason for update")
	_ = itemUpdateCmd.MarkFlagRequired("reason")
}

var itemDeleteCmd = &cobra.Command{
	Use:   "delete <itemRef>",
	Short: "Delete an item",
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
		if err := svc.Items.Delete(project, args[0], reason); err != nil {
			return err
		}
		fmt.Printf("Deleted %s.\n", args[0])
		return nil
	},
}

func init() {
	itemDeleteCmd.Flags().StringP("reason", "r", "", "Reason for deletion")
	_ = itemDeleteCmd.MarkFlagRequired("reason")
}

var itemRestoreCmd = &cobra.Command{
	Use:   "restore <itemRef>",
	Short: "Restore a deleted item",
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
		ack, err := svc.Items.Restore(project, args[0], reason)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		fmt.Printf("Restored %s to parent=%s, order=%d\n", args[0], ack.NewParent, ack.NewOrder)
		return nil
	},
}

func init() {
	itemRestoreCmd.Flags().StringP("reason", "r", "", "Reason for restore")
	_ = itemRestoreCmd.MarkFlagRequired("reason")
}

var itemCopyCmd = &cobra.Command{
	Use:   "copy <itemOrFolder> <targetFolder>",
	Short: "Copy item(s) to another folder",
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
		copyLabels, _ := cmd.Flags().GetInt("copy-labels")
		ack, err := svc.Items.Copy(project, args[0], args[1], reason, copyLabels)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		fmt.Printf("Copied %d items/folders: %v\n", len(ack.ItemsAndFoldersCreated), ack.ItemsAndFoldersCreated)
		return nil
	},
}

func init() {
	itemCopyCmd.Flags().StringP("reason", "r", "", "Reason for copy")
	itemCopyCmd.Flags().Int("copy-labels", 1, "Copy labels (0=no, 1=yes)")
	_ = itemCopyCmd.MarkFlagRequired("reason")
}

var itemMoveCmd = &cobra.Command{
	Use:   "move <targetFolder> <items>",
	Short: "Move items into a folder (items as comma-separated refs)",
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
		if err := svc.Items.Move(project, args[0], args[1], reason); err != nil {
			return err
		}
		fmt.Printf("Moved items to %s.\n", args[0])
		return nil
	},
}

func init() {
	itemMoveCmd.Flags().StringP("reason", "r", "", "Reason for move")
	_ = itemMoveCmd.MarkFlagRequired("reason")
}

var itemTouchCmd = &cobra.Command{
	Use:   "touch <itemRef>",
	Short: "Touch an item (update modification date)",
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
		if err := svc.Items.Touch(project, args[0], reason); err != nil {
			return err
		}
		fmt.Printf("Touched %s.\n", args[0])
		return nil
	},
}

func init() {
	itemTouchCmd.Flags().StringP("reason", "r", "", "Reason for touch")
	_ = itemTouchCmd.MarkFlagRequired("reason")
}
