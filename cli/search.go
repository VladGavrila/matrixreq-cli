package cli

import (
	"fmt"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for items in a project",
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
		filter, _ := cmd.Flags().GetString("filter")
		minimal, _ := cmd.Flags().GetBool("minimal")

		if minimal {
			data, err := svc.Search.SearchMinimal(project, args[0], filter)
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		needle := &api.TrimNeedle{
			Search: args[0],
			Filter: filter,
		}
		results, err := svc.Search.Search(project, needle)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), results)
		}
		headers := []string{"Ref", "Title", "Labels", "Modified"}
		var rows [][]string
		for _, r := range results {
			rows = append(rows, []string{r.ItemOrFolderRef, r.Title, r.Labels, r.LastModDate})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

func init() {
	searchCmd.Flags().String("filter", "", "Filter by categories (e.g., REQ,SPEC)")
	searchCmd.Flags().Bool("minimal", false, "Return only item IDs")
}
