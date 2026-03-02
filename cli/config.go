package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/VladGavrila/matrixreq-cli/internal/client"
	"github.com/VladGavrila/matrixreq-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage mxreq configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Matrix instance URL: ")
		url, _ := reader.ReadString('\n')
		url = strings.TrimSpace(url)
		if url == "" {
			return fmt.Errorf("URL is required")
		}

		// Validate URL by calling GET /all/status
		fmt.Print("Validating URL... ")
		c := client.New(url, "", flagDebug)
		if _, err := c.Get("/all/status"); err != nil {
			fmt.Println("warning: could not reach instance (continuing anyway)")
		} else {
			fmt.Println("OK")
		}

		fmt.Print("API token: ")
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("token is required")
		}

		fmt.Print("Default project (optional): ")
		project, _ := reader.ReadString('\n')
		project = strings.TrimSpace(project)

		cfg := &config.Config{
			URL:            url,
			Token:          token,
			DefaultProject: project,
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		path, _ := config.ConfigPath()
		fmt.Printf("Configuration saved to %s\n", path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		path, _ := config.ConfigPath()
		fmt.Printf("Config file: %s\n", path)
		fmt.Printf("URL:         %s\n", cfg.URL)
		if len(cfg.Token) > 8 {
			fmt.Printf("Token:       %s...%s\n", cfg.Token[:4], cfg.Token[len(cfg.Token)-4:])
		} else if cfg.Token != "" {
			fmt.Printf("Token:       ***\n")
		} else {
			fmt.Printf("Token:       (not set)\n")
		}
		fmt.Printf("Project:     %s\n", cfg.DefaultProject)
		return nil
	},
}
