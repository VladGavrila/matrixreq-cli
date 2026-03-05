package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userGetCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userUpdateCmd)
	userCmd.AddCommand(userDeleteCmd)
	userCmd.AddCommand(userRenameCmd)
	userCmd.AddCommand(userTokenCmd)
	userCmd.AddCommand(userAuditCmd)
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		details, _ := cmd.Flags().GetBool("details")
		filter, _ := cmd.Flags().GetString("filter")
		users, err := svc.Users.List(details)
		if err != nil {
			return err
		}
		if filter != "" {
			filter = strings.ToLower(filter)
			var filtered []api.UserType
			for _, u := range users {
				if strings.Contains(strings.ToLower(u.Login), filter) ||
					strings.Contains(strings.ToLower(u.Email), filter) ||
					strings.Contains(strings.ToLower(u.FirstName), filter) ||
					strings.Contains(strings.ToLower(u.LastName), filter) {
					filtered = append(filtered, u)
				}
			}
			users = filtered
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), users)
		}
		headers := []string{"ID", "Login", "Email", "First", "Last", "Status"}
		var rows [][]string
		for _, u := range users {
			rows = append(rows, []string{
				strconv.Itoa(u.ID), u.Login, u.Email, u.FirstName, u.LastName, u.UserStatus,
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

func init() {
	userListCmd.Flags().Bool("details", false, "Include detailed information")
	userListCmd.Flags().String("filter", "", "Filter users by login, email, or name")
}

var userGetCmd = &cobra.Command{
	Use:   "get <user>",
	Short: "Get user details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		user, err := svc.Users.Get(args[0])
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), user)
		}
		fmt.Printf("ID:     %d\n", user.ID)
		fmt.Printf("Login:  %s\n", user.Login)
		fmt.Printf("Email:  %s\n", user.Email)
		fmt.Printf("Name:   %s %s\n", user.FirstName, user.LastName)
		fmt.Printf("Status: %s\n", user.UserStatus)
		fmt.Printf("Admin:  %d\n", user.CustomerAdmin)
		if len(user.TokenList) > 0 {
			fmt.Println("\nTokens:")
			for _, t := range user.TokenList {
				fmt.Printf("  [%d] %s (%s) valid until %s\n", t.TokenID, t.Purpose, t.Reason, t.ValidToUserFormat)
			}
		}
		return nil
	},
}

var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		login, _ := cmd.Flags().GetString("login")
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		first, _ := cmd.Flags().GetString("first")
		last, _ := cmd.Flags().GetString("last")
		if err := svc.Users.Create(login, email, password, first, last); err != nil {
			return err
		}
		fmt.Printf("User %q created.\n", login)
		return nil
	},
}

func init() {
	userCreateCmd.Flags().String("login", "", "Login name")
	userCreateCmd.Flags().String("email", "", "Email address")
	userCreateCmd.Flags().String("password", "", "Password")
	userCreateCmd.Flags().String("first", "", "First name")
	userCreateCmd.Flags().String("last", "", "Last name")
	_ = userCreateCmd.MarkFlagRequired("login")
	_ = userCreateCmd.MarkFlagRequired("email")
	_ = userCreateCmd.MarkFlagRequired("password")
}

var userUpdateCmd = &cobra.Command{
	Use:   "update <user>",
	Short: "Update user details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		email, _ := cmd.Flags().GetString("email")
		password, _ := cmd.Flags().GetString("password")
		first, _ := cmd.Flags().GetString("first")
		last, _ := cmd.Flags().GetString("last")
		if err := svc.Users.Update(args[0], email, password, first, last); err != nil {
			return err
		}
		fmt.Printf("User %q updated.\n", args[0])
		return nil
	},
}

func init() {
	userUpdateCmd.Flags().String("email", "", "Email address")
	userUpdateCmd.Flags().String("password", "", "Password")
	userUpdateCmd.Flags().String("first", "", "First name")
	userUpdateCmd.Flags().String("last", "", "Last name")
}

var userDeleteCmd = &cobra.Command{
	Use:   "delete <user>",
	Short: "Delete a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		if err := svc.Users.Delete(args[0]); err != nil {
			return err
		}
		fmt.Printf("User %q deleted.\n", args[0])
		return nil
	},
}

var userRenameCmd = &cobra.Command{
	Use:   "rename <user> <newLogin>",
	Short: "Rename a user's login",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		if err := svc.Users.Rename(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("User %q renamed to %q.\n", args[0], args[1])
		return nil
	},
}

var userTokenCmd = &cobra.Command{
	Use:   "token <user>",
	Short: "Create an API token for a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		purpose, _ := cmd.Flags().GetString("purpose")
		reason, _ := cmd.Flags().GetString("reason")
		validity, _ := cmd.Flags().GetInt("validity")
		token, err := svc.Users.CreateToken(args[0], purpose, reason, validity)
		if err != nil {
			return err
		}
		fmt.Printf("Token created: %s\n", token)
		return nil
	},
}

func init() {
	userTokenCmd.Flags().String("purpose", "", "Token purpose")
	userTokenCmd.Flags().String("reason", "", "Reason for creation")
	userTokenCmd.Flags().Int("validity", 365, "Validity in days")
	_ = userTokenCmd.MarkFlagRequired("purpose")
}

var userAuditCmd = &cobra.Command{
	Use:   "audit <user>",
	Short: "Show user audit log",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		startAt, _ := cmd.Flags().GetInt("start")
		maxResults, _ := cmd.Flags().GetInt("max")
		audit, err := svc.Users.Audit(args[0], startAt, maxResults)
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
	userAuditCmd.Flags().Int("start", 0, "Start at offset")
	userAuditCmd.Flags().Int("max", 50, "Maximum results")
}
