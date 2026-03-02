package cli

import (
	"encoding/json"
	"fmt"

	"github.com/VladGavrila/matrixreq-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(adminCmd)
	adminCmd.AddCommand(adminStatusCmd)
	adminCmd.AddCommand(adminLicenseCmd)
	adminCmd.AddCommand(adminMonitorCmd)
	adminCmd.AddCommand(adminSettingsCmd)
}

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "System administration",
}

var adminStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get instance status",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		status, err := svc.Admin.Status()
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			return output.PrintItem(getOutputFormat(), status)
		}
		fmt.Printf("Version:    %s\n", status.Version)
		fmt.Printf("Public URL: %s\n", status.PublicURL)
		if status.ExceptionStatus != nil {
			fmt.Printf("Exceptions: %d since start, %d in last hour\n",
				status.ExceptionStatus.NbExceptionsStillStart,
				len(status.ExceptionStatus.LastHourExceptions))
		}
		return nil
	},
}

var adminLicenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Show license status",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		data, err := svc.Admin.License()
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			fmt.Println(string(data))
			return nil
		}
		var license any
		if err := json.Unmarshal(data, &license); err != nil {
			fmt.Println(string(data))
			return nil
		}
		out, _ := json.MarshalIndent(license, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var adminMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Show monitoring information",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		data, err := svc.Admin.Monitor()
		if err != nil {
			return err
		}
		if getOutputFormat() == "json" {
			fmt.Println(string(data))
			return nil
		}
		var monitor any
		if err := json.Unmarshal(data, &monitor); err != nil {
			fmt.Println(string(data))
			return nil
		}
		out, _ := json.MarshalIndent(monitor, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var adminSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Show or set global settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newService()
		if err != nil {
			return err
		}
		key, _ := cmd.Flags().GetString("set-key")
		value, _ := cmd.Flags().GetString("set-value")

		if key != "" {
			if err := svc.Admin.SetSetting(key, value); err != nil {
				return err
			}
			fmt.Printf("Setting %q updated.\n", key)
			return nil
		}

		settings, err := svc.Admin.GetSettings()
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

func init() {
	adminSettingsCmd.Flags().String("set-key", "", "Setting key to update")
	adminSettingsCmd.Flags().String("set-value", "", "Setting value to set")
}
