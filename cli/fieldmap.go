package cli

import (
	"fmt"

	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

var fieldmapCmd = &cobra.Command{
	Use:   "fieldmap",
	Short: "Manage field ID mappings for the current project",
}

var fieldmapShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show resolved field IDs for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}

		fm, err := fieldmap.LoadOrFetch(svc, project)
		if err != nil {
			return fmt.Errorf("loading field map: %w", err)
		}

		category, _ := cmd.Flags().GetString("category")
		entries := fm.Entries()

		if getOutputFormat() == "json" {
			if category != "" {
				fields := fm.FieldsForCategory(category)
				return output.PrintItem(getOutputFormat(), fields)
			}
			// Build a nested map for JSON output
			result := make(map[string]map[string]int)
			for _, e := range entries {
				if result[e.Category] == nil {
					result[e.Category] = make(map[string]int)
				}
				result[e.Category][e.Label] = e.FieldID
			}
			return output.PrintItem(getOutputFormat(), result)
		}

		headers := []string{"Category", "Field Label", "Field ID"}
		var rows [][]string
		for _, e := range entries {
			if category != "" && e.Category != category {
				continue
			}
			rows = append(rows, []string{e.Category, e.Label, fmt.Sprintf("%d", e.FieldID)})
		}

		if len(rows) == 0 {
			if category != "" {
				fmt.Printf("No fields found for category %q in project %s\n", category, project)
			} else {
				fmt.Printf("No fields found for project %s\n", project)
			}
			return nil
		}

		return output.Print(getOutputFormat(), headers, rows)
	},
}

var fieldmapClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cached field mappings",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if all {
			if err := fieldmap.ClearAll(); err != nil {
				return fmt.Errorf("clearing all field cache: %w", err)
			}
			fmt.Println("Cleared all cached field mappings")
			return nil
		}

		project, err := requireProject()
		if err != nil {
			return err
		}

		if err := fieldmap.Clear(project); err != nil {
			return fmt.Errorf("clearing field cache: %w", err)
		}
		fmt.Printf("Cleared cached field mappings for project %s\n", project)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fieldmapCmd)
	fieldmapCmd.AddCommand(fieldmapShowCmd)
	fieldmapCmd.AddCommand(fieldmapClearCmd)

	fieldmapShowCmd.Flags().String("category", "", "Filter by category short label (e.g., TC, XTC, REQ)")
	fieldmapClearCmd.Flags().Bool("all", false, "Clear field cache for all projects")
}
