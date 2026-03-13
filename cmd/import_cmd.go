package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/spf13/cobra"
)

var importConfirm bool

var importCmd = &cobra.Command{
	Use:   "import [file]",
	Short: "Import profile data from JSON file or stdin",
	Long:  "Import profile data. Reads from file argument or stdin. Use --confirm to skip the duplicate check prompt in non-interactive mode.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		var data []byte
		var err error
		if len(args) > 0 && args[0] != "-" {
			data, err = os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
		} else {
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
		}

		if len(data) == 0 {
			return fmt.Errorf("no input data")
		}

		var importData any
		if err := json.Unmarshal(data, &importData); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}

		if !importConfirm {
			var checkResult models.ImportCheck
			if err := ctx.Client.Post(context.Background(), "/api/import/check", importData, &checkResult); err != nil {
				return handleError(err)
			}
			if checkResult.Exists {
				userName := "unknown"
				if checkResult.User != nil {
					userName = checkResult.User.Name
				}
				if ctx.IsJSON {
					return ctx.Output.Output(map[string]any{
						"warning":  "profile_exists",
						"message":  fmt.Sprintf("Profile '%s' already exists", userName),
						"existing": checkResult.User,
					}, nil)
				}
				return fmt.Errorf("profile already exists. Use --confirm to import anyway")
			}
		}

		var result models.ImportResult
		if err := ctx.Client.Post(context.Background(), "/api/import", importData, &result); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(result, nil)
	},
}

func init() {
	importCmd.Flags().BoolVar(&importConfirm, "confirm", false, "Skip duplicate check and import directly")
	rootCmd.AddCommand(importCmd)
}
