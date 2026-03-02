package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectTreeCmd)
	projectCmd.AddCommand(projectAccessCmd)
	projectCmd.AddCommand(projectAuditCmd)
	projectCmd.AddCommand(projectHideCmd)
	projectCmd.AddCommand(projectUnhideCmd)
}

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"proj"},
	Short:   "Manage projects",
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		projects, err := svc.Projects.List()
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), projects)
		}
		headers := []string{"ID", "Short", "Label", "QMS", "Access"}
		var rows [][]string
		for _, p := range projects {
			qms := ""
			if p.QMSProject {
				qms = "yes"
			}
			rows = append(rows, []string{
				strconv.Itoa(p.ID), p.ShortLabel, p.Label, qms, p.AccessType,
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get [project]",
	Short: "Get project details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project := ""
		if len(args) > 0 {
			project = args[0]
		} else {
			project, err = requireProject()
			if err != nil {
				return err
			}
		}
		info, err := svc.Projects.Get(project)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), info)
		}
		fmt.Printf("Project: %s (%s)\n", info.Label, info.ShortLabel)
		if info.Access != nil {
			fmt.Printf("Access:  readWrite=%d, visitorOnly=%v\n", info.Access.ReadWrite, info.Access.VisitorOnly)
		}
		fmt.Printf("Categories: %d\n", len(info.CategoryList.CategoryExtended))
		for _, cat := range info.CategoryList.CategoryExtended {
			fmt.Printf("  - %s (%s) [%d fields]\n", cat.Category.Label, cat.Category.ShortLabel, len(cat.FieldList.Field))
		}
		return nil
	},
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <label> <shortLabel>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		if err := svc.Projects.Create(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Project %q (%s) created.\n", args[0], args[1])
		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <project>",
	Short: "Delete a project permanently",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		if err := svc.Projects.Delete(args[0], "yes"); err != nil {
			return err
		}
		fmt.Printf("Project %q deleted.\n", args[0])
		return nil
	},
}

var projectTreeCmd = &cobra.Command{
	Use:   "tree [project]",
	Short: "Show project tree structure",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project := ""
		if len(args) > 0 {
			project = args[0]
		} else {
			project, err = requireProject()
			if err != nil {
				return err
			}
		}
		filter, _ := cmd.Flags().GetString("filter")
		data, err := svc.Projects.Tree(project, filter)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			fmt.Println(string(data))
			return nil
		}
		// Pretty-print the JSON tree
		var tree any
		if err := json.Unmarshal(data, &tree); err != nil {
			fmt.Println(string(data))
			return nil
		}
		out, _ := json.MarshalIndent(tree, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

func init() {
	projectTreeCmd.Flags().String("filter", "", "Filter categories (e.g., REQ,SPEC)")
}

var projectAccessCmd = &cobra.Command{
	Use:   "access [project]",
	Short: "Show project access (users and groups)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project := ""
		if len(args) > 0 {
			project = args[0]
		} else {
			project, err = requireProject()
			if err != nil {
				return err
			}
		}
		access, err := svc.Projects.Access(project)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), access)
		}
		fmt.Println("Group Permissions:")
		for _, g := range access.GroupPermission {
			fmt.Printf("  %s (id=%d): permission=%d, members=%d\n",
				g.GroupName, g.GroupID, g.Permission, len(g.Membership))
		}
		fmt.Println("\nUser Permissions:")
		for _, u := range access.UserPermission {
			fmt.Printf("  %s (%s): permission=%d\n", u.Login, u.Email, u.Permission)
		}
		return nil
	},
}

var projectAuditCmd = &cobra.Command{
	Use:   "audit [project]",
	Short: "Show project audit log",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		project := ""
		if len(args) > 0 {
			project = args[0]
		} else {
			project, err = requireProject()
			if err != nil {
				return err
			}
		}
		startAt, _ := cmd.Flags().GetInt("start")
		maxResults, _ := cmd.Flags().GetInt("max")
		audit, err := svc.Projects.Audit(project, startAt, maxResults)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), audit)
		}
		headers := []string{"ID", "Action", "User", "Date", "Item", "Reason"}
		var rows [][]string
		for _, a := range audit.Audit {
			rows = append(rows, []string{
				strconv.Itoa(a.AuditID), a.Action, a.UserLogin, a.DateUser, a.ItemRef(), a.Reason,
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

func init() {
	projectAuditCmd.Flags().Int("start", 0, "Start at offset")
	projectAuditCmd.Flags().Int("max", 50, "Maximum results")
}

var projectHideCmd = &cobra.Command{
	Use:   "hide <project>",
	Short: "Hide a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		reason, _ := cmd.Flags().GetString("reason")
		if err := svc.Projects.Hide(args[0], reason); err != nil {
			return err
		}
		fmt.Printf("Project %q hidden.\n", args[0])
		return nil
	},
}

func init() {
	projectHideCmd.Flags().StringP("reason", "r", "", "Reason for hiding")
	_ = projectHideCmd.MarkFlagRequired("reason")
}

var projectUnhideCmd = &cobra.Command{
	Use:   "unhide <project> <newShortLabel>",
	Short: "Unhide a project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		reason, _ := cmd.Flags().GetString("reason")
		if err := svc.Projects.Unhide(args[0], args[1], reason); err != nil {
			return err
		}
		fmt.Printf("Project %q unhidden as %q.\n", args[0], args[1])
		return nil
	},
}

func init() {
	projectUnhideCmd.Flags().StringP("reason", "r", "", "Reason for unhiding")
	_ = projectUnhideCmd.MarkFlagRequired("reason")
}
