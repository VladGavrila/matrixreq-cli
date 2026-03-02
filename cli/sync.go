package cli

import (
	"fmt"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/fieldmap"
	"github.com/VladGavrila/matrixreq-cli/internal/itemsync"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync <file>",
	Short: "Sync items from a file to Matrix",
	Long: `Parse items from a source file (.py, .go, .ts) or YAML definition file (.yaml)
and create/update them in Matrix.

Source files: Finds test functions with YAML docstrings. NEW_ functions create items,
TC_### functions update existing items.

YAML files: Reads an items list. Items without item_ref are created, items with
item_ref are updated.`,
	Args: cobra.ExactArgs(1),
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		createOnly, _ := cmd.Flags().GetBool("create")
		updateOnly, _ := cmd.Flags().GetBool("update")
		yes, _ := cmd.Flags().GetBool("yes")

		opts := itemsync.SyncOptions{
			DryRun:     dryRun,
			CreateOnly: createOnly,
			UpdateOnly: updateOnly,
			Yes:        yes,
		}

		result, err := itemsync.Sync(svc, project, args[0], fm, opts)
		if err != nil {
			return err
		}

		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), result)
		}

		// Print summary
		if len(result.Created) > 0 {
			fmt.Println("Created:")
			for _, e := range result.Created {
				fmt.Printf("  %s → %s (%s)\n", e.Title, e.ItemRef, e.Folder)
			}
		}
		if len(result.Updated) > 0 {
			fmt.Println("Updated:")
			for _, e := range result.Updated {
				fmt.Printf("  %s (%s)\n", e.ItemRef, e.Title)
			}
		}
		if len(result.Skipped) > 0 {
			fmt.Printf("Skipped: %d items\n", len(result.Skipped))
		}
		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for _, e := range result.Errors {
				ref := e.ItemRef
				if ref == "" {
					ref = e.Title
				}
				fmt.Printf("  %s: %s\n", ref, e.Error)
			}
		}

		total := len(result.Created) + len(result.Updated)
		if total == 0 && len(result.Errors) == 0 {
			fmt.Println("No items to sync")
		}

		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff <file>",
	Short: "Compare local item definitions against server state",
	Long: `Parse items from a source file or YAML definition file and compare each
item's local definition against its current state on the server.

This is a read-only operation that shows what would change if you ran sync.`,
	Args: cobra.ExactArgs(1),
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

		diffs, err := itemsync.Diff(svc, project, args[0], fm)
		if err != nil {
			return err
		}

		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), diffs)
		}

		if len(diffs) == 0 {
			fmt.Println("No items found in file")
			return nil
		}

		allEqual := true
		for _, d := range diffs {
			if d.IsEqual {
				fmt.Printf("  %s: no changes\n", d.ItemRef)
				continue
			}
			allEqual = false
			fmt.Printf("  %s:\n", d.ItemRef)
			for field, vals := range d.Differences {
				fmt.Printf("    %s:\n", field)
				fmt.Printf("      local:  %s\n", vals[0])
				fmt.Printf("      server: %s\n", vals[1])
			}
			if d.LabelDiff != nil {
				if len(d.LabelDiff.Added) > 0 {
					fmt.Printf("    labels added:   %s\n", strings.Join(d.LabelDiff.Added, ", "))
				}
				if len(d.LabelDiff.Removed) > 0 {
					fmt.Printf("    labels removed: %s\n", strings.Join(d.LabelDiff.Removed, ", "))
				}
			}
			if d.LinkDiff != nil {
				if len(d.LinkDiff.Added) > 0 {
					fmt.Printf("    links added:   %s\n", strings.Join(d.LinkDiff.Added, ", "))
				}
				if len(d.LinkDiff.Removed) > 0 {
					fmt.Printf("    links removed: %s\n", strings.Join(d.LinkDiff.Removed, ", "))
				}
			}
		}

		if allEqual {
			fmt.Println("All items are in sync")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(diffCmd)

	syncCmd.Flags().Bool("dry-run", false, "Show what would happen without making changes")
	syncCmd.Flags().Bool("create", false, "Only create new items")
	syncCmd.Flags().Bool("update", false, "Only update existing items")
	syncCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
}
