package cli

import (
	"fmt"
	"strconv"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(todoCmd)
	todoCmd.AddCommand(todoListCmd)
	todoCmd.AddCommand(todoCreateCmd)
	todoCmd.AddCommand(todoDoneCmd)
}

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "Manage todos",
}

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List todos",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		includeDone, _ := cmd.Flags().GetBool("done")
		includeFuture, _ := cmd.Flags().GetBool("future")
		all, _ := cmd.Flags().GetBool("all")

		var todos []struct {
			TodoID  int
			Login   string
			ItemRef string
			Action  string
			Created string
			Project string
		}

		if all {
			list, err := svc.Todos.ListAll(includeDone, includeFuture)
			if err != nil {
				return err
			}
			if getOutputFormat() == "json" {
				return output.PrintItem(getOutputFormat(), list)
			}
			for _, t := range list {
				action := ""
				if t.Action != nil {
					action = t.Action.Text
				}
				todos = append(todos, struct {
					TodoID  int
					Login   string
					ItemRef string
					Action  string
					Created string
					Project string
				}{t.TodoID, t.Login, t.ItemRef, action, t.CreatedAtUserFormat, t.ProjectShort})
			}
		} else {
			project, err := requireProject()
			if err != nil {
				return err
			}
			list, err := svc.Todos.List(project, includeDone, includeFuture)
			if err != nil {
				return err
			}
			if getOutputFormat() == "json" {
				return output.PrintItem(getOutputFormat(), list)
			}
			for _, t := range list {
				action := ""
				if t.Action != nil {
					action = t.Action.Text
				}
				todos = append(todos, struct {
					TodoID  int
					Login   string
					ItemRef string
					Action  string
					Created string
					Project string
				}{t.TodoID, t.Login, t.ItemRef, action, t.CreatedAtUserFormat, t.ProjectShort})
			}
		}

		headers := []string{"ID", "User", "Item", "Action", "Created", "Project"}
		var rows [][]string
		for _, t := range todos {
			rows = append(rows, []string{
				strconv.Itoa(t.TodoID), t.Login, t.ItemRef, t.Action, t.Created, t.Project,
			})
		}
		return output.Print(getOutputFormat(), headers, rows)
	},
}

func init() {
	todoListCmd.Flags().Bool("done", false, "Include completed todos")
	todoListCmd.Flags().Bool("future", false, "Include future todos")
	todoListCmd.Flags().Bool("all", false, "List across all projects")
}

var todoCreateCmd = &cobra.Command{
	Use:   "create <item>",
	Short: "Create a todo on an item",
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
		text, _ := cmd.Flags().GetString("text")
		todoType, _ := cmd.Flags().GetString("type")
		fieldID, _ := cmd.Flags().GetInt("field-id")
		logins, _ := cmd.Flags().GetString("logins")
		if err := svc.Todos.Create(project, upperRef(args[0]), text, todoType, fieldID, logins); err != nil {
			return err
		}
		fmt.Printf("Todo created on %s.\n", upperRef(args[0]))
		return nil
	},
}

func init() {
	todoCreateCmd.Flags().String("text", "", "Todo text")
	todoCreateCmd.Flags().String("type", "", "Todo type")
	todoCreateCmd.Flags().Int("field-id", 0, "Field ID")
	todoCreateCmd.Flags().String("logins", "", "Comma-separated user logins")
	_ = todoCreateCmd.MarkFlagRequired("text")
}

var todoDoneCmd = &cobra.Command{
	Use:   "done <todoID>",
	Short: "Mark a todo as done",
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
		todoID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid todo ID: %s", args[0])
		}
		hardDelete, _ := cmd.Flags().GetBool("hard-delete")
		if err := svc.Todos.Done(project, todoID, hardDelete); err != nil {
			return err
		}
		fmt.Printf("Todo %d marked as done.\n", todoID)
		return nil
	},
}

func init() {
	todoDoneCmd.Flags().Bool("hard-delete", false, "Permanently delete instead of marking done")
}
