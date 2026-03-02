package cli

import (
	"fmt"
	"strconv"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(categoryCmd)
	categoryCmd.AddCommand(categoryListCmd)
	categoryCmd.AddCommand(categoryGetCmd)
	categoryCmd.AddCommand(categoryCreateCmd)
	categoryCmd.AddCommand(categoryUpdateCmd)
	categoryCmd.AddCommand(categoryDeleteCmd)
	categoryCmd.AddCommand(categorySettingsCmd)
}

var categoryCmd = &cobra.Command{
	Use:     "category",
	Aliases: []string{"cat"},
	Short:   "Manage categories",
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories in a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		cats, err := svc.Categories.List(project)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), cats)
		}
		headers := []string{"ID", "Short", "Label", "Fields"}
		var rows [][]string
		for _, c := range cats {
			rows = append(rows, []string{
				strconv.Itoa(c.Category.ID),
				c.Category.ShortLabel,
				c.Category.Label,
				strconv.Itoa(len(c.FieldList.Field)),
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

var categoryGetCmd = &cobra.Command{
	Use:   "get <category>",
	Short: "Get category details",
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
		cat, err := svc.Categories.Get(project, args[0])
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), cat)
		}
		fmt.Printf("Category: %s (%s)\n", cat.Category.Label, cat.Category.ShortLabel)
		fmt.Printf("Fields:   %d\n", len(cat.FieldList.Field))
		for _, f := range cat.FieldList.Field {
			fmt.Printf("  [%d] %s (%s) order=%d\n", f.ID, f.Label, f.FieldType, f.Order)
		}
		fmt.Printf("Folders:  %d\n", len(cat.FolderList))
		return nil
	},
}

var categoryCreateCmd = &cobra.Command{
	Use:   "create <label> <shortLabel>",
	Short: "Create a new category",
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
		if err := svc.Categories.Create(project, args[0], args[1], reason); err != nil {
			return err
		}
		fmt.Printf("Category %q (%s) created.\n", args[0], args[1])
		return nil
	},
}

func init() {
	categoryCreateCmd.Flags().StringP("reason", "r", "", "Reason for creation")
	_ = categoryCreateCmd.MarkFlagRequired("reason")
}

var categoryUpdateCmd = &cobra.Command{
	Use:   "update <category>",
	Short: "Update a category",
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
		shortLabel, _ := cmd.Flags().GetString("short")
		reason, _ := cmd.Flags().GetString("reason")
		order, _ := cmd.Flags().GetInt("order")
		if err := svc.Categories.Update(project, args[0], label, shortLabel, reason, order); err != nil {
			return err
		}
		fmt.Printf("Category %q updated.\n", args[0])
		return nil
	},
}

func init() {
	categoryUpdateCmd.Flags().String("label", "", "New label")
	categoryUpdateCmd.Flags().String("short", "", "New short label")
	categoryUpdateCmd.Flags().StringP("reason", "r", "", "Reason for update")
	categoryUpdateCmd.Flags().Int("order", 0, "New order")
	_ = categoryUpdateCmd.MarkFlagRequired("reason")
}

var categoryDeleteCmd = &cobra.Command{
	Use:   "delete <category>",
	Short: "Delete a category",
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
		if err := svc.Categories.Delete(project, args[0], reason); err != nil {
			return err
		}
		fmt.Printf("Category %q deleted.\n", args[0])
		return nil
	},
}

func init() {
	categoryDeleteCmd.Flags().StringP("reason", "r", "", "Reason for deletion")
	_ = categoryDeleteCmd.MarkFlagRequired("reason")
}

var categorySettingsCmd = &cobra.Command{
	Use:   "settings <category>",
	Short: "Show category settings",
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
		settings, err := svc.Categories.GetSettings(project, args[0])
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), settings)
		}
		headers := []string{"Key", "Value", "Secret"}
		var rows [][]string
		for _, s := range settings {
			secret := ""
			if s.Secret {
				secret = "yes"
			}
			rows = append(rows, []string{s.Key, s.Value, secret})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}
