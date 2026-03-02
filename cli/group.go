package cli

import (
	"fmt"
	"strconv"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupGetCmd)
	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	groupCmd.AddCommand(groupAddUserCmd)
	groupCmd.AddCommand(groupRemoveUserCmd)
}

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		details, _ := cmd.Flags().GetBool("details")
		groups, err := svc.Groups.List(details)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), groups)
		}
		headers := []string{"ID", "Name", "Members"}
		var rows [][]string
		for _, g := range groups {
			rows = append(rows, []string{
				strconv.Itoa(g.GroupID), g.GroupName, strconv.Itoa(len(g.Membership)),
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

func init() {
	groupListCmd.Flags().Bool("details", false, "Include member details")
}

var groupGetCmd = &cobra.Command{
	Use:   "get <groupID>",
	Short: "Get group details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		groupID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid group ID: %s", args[0])
		}
		group, err := svc.Groups.Get(groupID, true)
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), group)
		}
		fmt.Printf("Group ID: %d\n", group.GroupID)
		fmt.Printf("Name:     %s\n", group.GroupName)
		if len(group.Membership) > 0 {
			fmt.Println("Members:")
			for _, m := range group.Membership {
				fmt.Printf("  %s (%s)\n", m.Login, m.Email)
			}
		}
		return nil
	},
}

var groupCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		if err := svc.Groups.Create(args[0]); err != nil {
			return err
		}
		fmt.Printf("Group %q created.\n", args[0])
		return nil
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete <groupID>",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		groupID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid group ID: %s", args[0])
		}
		if err := svc.Groups.Delete(groupID); err != nil {
			return err
		}
		fmt.Printf("Group %d deleted.\n", groupID)
		return nil
	},
}

var groupAddUserCmd = &cobra.Command{
	Use:   "add-user <groupID> <user>",
	Short: "Add a user to a group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		groupID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid group ID: %s", args[0])
		}
		if err := svc.Groups.AddUser(groupID, args[1]); err != nil {
			return err
		}
		fmt.Printf("User %q added to group %d.\n", args[1], groupID)
		return nil
	},
}

var groupRemoveUserCmd = &cobra.Command{
	Use:   "remove-user <groupName> <user>",
	Short: "Remove a user from a group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		if err := svc.Groups.RemoveUser(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("User %q removed from group %q.\n", args[1], args[0])
		return nil
	},
}
