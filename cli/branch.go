package cli

import (
	"fmt"
	"strconv"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(branchCmd)
	branchCmd.AddCommand(branchCreateCmd)
	branchCmd.AddCommand(branchCloneCmd)
	branchCmd.AddCommand(branchMergeCmd)
	branchCmd.AddCommand(branchInfoCmd)
	branchCmd.AddCommand(branchHistoryCmd)
}

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Manage project branches",
}

var branchCreateCmd = &cobra.Command{
	Use:   "create <label> <shortLabel>",
	Short: "Create a branch of the current project",
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
		tag, _ := cmd.Flags().GetString("tag")
		keepPerms, _ := cmd.Flags().GetInt("keep-permissions")
		keepContent, _ := cmd.Flags().GetInt("keep-content")
		if err := svc.Branches.Create(project, args[0], args[1], tag, keepPerms, keepContent); err != nil {
			return err
		}
		fmt.Printf("Branch %q (%s) created from %s.\n", args[0], args[1], project)
		return nil
	},
}

func init() {
	branchCreateCmd.Flags().String("tag", "", "Tag to create for the branch point")
	branchCreateCmd.Flags().Int("keep-permissions", 1, "Keep permissions (0=no, 1=yes)")
	branchCreateCmd.Flags().Int("keep-content", 1, "Keep content (0=no, 1=yes)")
}

var branchCloneCmd = &cobra.Command{
	Use:   "clone <label> <shortLabel>",
	Short: "Clone the current project",
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
		keepHistory, _ := cmd.Flags().GetInt("keep-history")
		keepContent, _ := cmd.Flags().GetInt("keep-content")
		keepPerms, _ := cmd.Flags().GetInt("keep-permissions")
		if err := svc.Branches.Clone(project, args[0], args[1], keepHistory, keepContent, keepPerms); err != nil {
			return err
		}
		fmt.Printf("Project %q (%s) cloned from %s.\n", args[0], args[1], project)
		return nil
	},
}

func init() {
	branchCloneCmd.Flags().Int("keep-history", 1, "Keep history (0=no, 1=yes)")
	branchCloneCmd.Flags().Int("keep-content", 1, "Keep content (0=no, 1=yes)")
	branchCloneCmd.Flags().Int("keep-permissions", 1, "Keep permissions (0=no, 1=yes)")
}

var branchMergeCmd = &cobra.Command{
	Use:   "merge <branchProject>",
	Short: "Merge a branch into the main project",
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
		pushOrPull, _ := cmd.Flags().GetString("direction")
		params := &api.MergeParam{
			PushOrPull: pushOrPull,
		}
		if err := svc.Branches.Merge(project, args[0], reason, params); err != nil {
			return err
		}
		fmt.Printf("Merged %s into %s.\n", args[0], project)
		return nil
	},
}

func init() {
	branchMergeCmd.Flags().StringP("reason", "r", "", "Reason for merge")
	branchMergeCmd.Flags().String("direction", "push", "Merge direction (push or pull)")
	_ = branchMergeCmd.MarkFlagRequired("reason")
}

var branchInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show branch/merge info for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		data, err := svc.Branches.Info(project)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

var branchHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show merge history for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project, err := requireProject()
		if err != nil {
			return err
		}
		history, err := svc.Branches.History(project)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), history)
		}
		headers := []string{"ID", "Date", "User", "Direction", "Details"}
		var rows [][]string
		for _, h := range history {
			rows = append(rows, []string{
				strconv.Itoa(h.MergeID), h.Date, h.User, h.PushOrPull, h.Details,
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}
