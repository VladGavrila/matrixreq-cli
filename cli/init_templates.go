package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/VladGavrila/matrixreq-cli/internal/templates"
	"github.com/spf13/cobra"
)

// langDirs maps the --lang flag to the directory name in the embedded filesystem.
var langDirs = map[string]string{
	"py":         "python",
	"python":     "python",
	"go":         "go",
	"ts":         "typescript",
	"typescript": "typescript",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold test helper templates into the current directory",
	Long: `Extract mxreq test helper templates (docstring generator and results recorder)
into the current directory for the specified language.

Templates provide:
  - Docstring generator: builds YAML docstrings for test case synchronization
  - Results recorder: records execution results for upload to Matrix XTCs

Supported languages: py (Python), go (Go), ts (TypeScript)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		lang, _ := cmd.Flags().GetString("lang")
		force, _ := cmd.Flags().GetBool("force")

		dir, ok := langDirs[lang]
		if !ok {
			return fmt.Errorf("unsupported language %q (use py, go, or ts)", lang)
		}

		srcDir := filepath.Join("files", dir)
		entries, err := fs.ReadDir(templates.TemplateFS, srcDir)
		if err != nil {
			return fmt.Errorf("reading templates: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			srcPath := filepath.Join(srcDir, entry.Name())
			dstPath := entry.Name()

			// Check if file exists
			if _, err := os.Stat(dstPath); err == nil && !force {
				fmt.Printf("  skip %s (exists, use --force to overwrite)\n", dstPath)
				continue
			}

			data, err := fs.ReadFile(templates.TemplateFS, srcPath)
			if err != nil {
				return fmt.Errorf("reading template %s: %w", entry.Name(), err)
			}

			if err := os.WriteFile(dstPath, data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", dstPath, err)
			}

			fmt.Printf("  created %s\n", dstPath)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().String("lang", "", "Language: py, go, or ts")
	_ = initCmd.MarkFlagRequired("lang")
	initCmd.Flags().Bool("force", false, "Overwrite existing files")
}
