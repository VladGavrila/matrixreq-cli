package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fieldCmd)
	fieldCmd.AddCommand(fieldGetCmd)
	fieldCmd.AddCommand(fieldUpdateCmd)
	fieldCmd.AddCommand(fieldDeleteCmd)
	fieldCmd.AddCommand(fieldAddCmd)
}

var fieldCmd = &cobra.Command{
	Use:   "field",
	Short: "Manage fields",
}

var fieldGetCmd = &cobra.Command{
	Use:   "get <itemRef> <fieldName>",
	Short: "Get a field value from an item",
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
		value, err := svc.Fields.Get(project, upperRef(args[0]), args[1])
		if err != nil {
			return err
		}
		fmt.Println(value)
		return nil
	},
}

var fieldUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a field definition",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		fieldID, _ := cmd.Flags().GetInt("field-id")
		label, _ := cmd.Flags().GetString("label")
		param, _ := cmd.Flags().GetString("param")
		reason, _ := cmd.Flags().GetString("reason")
		order, _ := cmd.Flags().GetInt("order")
		if err := svc.Fields.Update(project, fieldID, label, param, reason, order); err != nil {
			return err
		}
		fmt.Printf("Field %d updated.\n", fieldID)
		return nil
	},
}

func init() {
	fieldUpdateCmd.Flags().Int("field-id", 0, "Field ID")
	fieldUpdateCmd.Flags().String("label", "", "New label")
	fieldUpdateCmd.Flags().String("param", "", "New parameter")
	fieldUpdateCmd.Flags().StringP("reason", "r", "", "Reason for update")
	fieldUpdateCmd.Flags().Int("order", 0, "New order")
	_ = fieldUpdateCmd.MarkFlagRequired("field-id")
	_ = fieldUpdateCmd.MarkFlagRequired("reason")
}

var fieldDeleteCmd = &cobra.Command{
	Use:   "delete <category>",
	Short: "Delete a field from a category",
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
		fieldID, _ := cmd.Flags().GetInt("field-id")
		reason, _ := cmd.Flags().GetString("reason")
		if err := svc.Fields.Delete(project, upperRef(args[0]), fieldID, reason); err != nil {
			return err
		}
		fmt.Printf("Field %d deleted from %s.\n", fieldID, upperRef(args[0]))
		return nil
	},
}

func init() {
	fieldDeleteCmd.Flags().Int("field-id", 0, "Field ID")
	fieldDeleteCmd.Flags().StringP("reason", "r", "", "Reason for deletion")
	_ = fieldDeleteCmd.MarkFlagRequired("field-id")
	_ = fieldDeleteCmd.MarkFlagRequired("reason")
}

var fieldAddCmd = &cobra.Command{
	Use:   "add <category>",
	Short: "Add a field to a category",
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
		label, _ := cmd.Flags().GetString("label")
		fieldType, _ := cmd.Flags().GetString("type")
		param, _ := cmd.Flags().GetString("param")
		reason, _ := cmd.Flags().GetString("reason")
		if err := svc.Fields.AddToCategory(project, upperRef(args[0]), label, fieldType, param, reason); err != nil {
			return err
		}
		fmt.Printf("Field %q added to %s.\n", label, upperRef(args[0]))
		return nil
	},
}

func init() {
	fieldAddCmd.Flags().String("label", "", "Field label")
	fieldAddCmd.Flags().String("type", "", "Field type")
	fieldAddCmd.Flags().String("param", "", "Field parameter")
	fieldAddCmd.Flags().StringP("reason", "r", "", "Reason for addition")
	_ = fieldAddCmd.MarkFlagRequired("label")
	_ = fieldAddCmd.MarkFlagRequired("type")
	_ = fieldAddCmd.MarkFlagRequired("param")
	_ = fieldAddCmd.MarkFlagRequired("reason")
}
