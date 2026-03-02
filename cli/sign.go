package cli

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/VladGavrila/matrixreq-cli/internal/api"
	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.AddCommand(signCreateCmd)
}

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Manage document signing",
}

var signCreateCmd = &cobra.Command{
	Use:   "create <item>",
	Short: "Sign an item (e.g., SIGN-1)",
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
		password, _ := cmd.Flags().GetString("password")
		acceptComments, _ := cmd.Flags().GetString("accept-comments")

		path := fmt.Sprintf("/%s/sign/%s?password=%s",
			url.PathEscape(project), url.PathEscape(args[0]), url.QueryEscape(password))
		if acceptComments != "" {
			path += "&acceptComments=" + url.QueryEscape(acceptComments)
		}
		data, err := svc.Client.Post(path, nil)
		if err != nil {
			return err
		}
		var ack api.SignItemAck
		if err := json.Unmarshal(data, &ack); err != nil {
			return fmt.Errorf("parsing sign response: %w", err)
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), ack)
		}
		fmt.Printf("Sign result: ok=%v, result=%s\n", ack.OK, ack.Result)
		return nil
	},
}

func init() {
	signCreateCmd.Flags().String("password", "", "Signature password")
	signCreateCmd.Flags().String("accept-comments", "", "Accept comments")
	_ = signCreateCmd.MarkFlagRequired("password")
}
