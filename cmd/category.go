package cmd

import (
	"context"

	"github.com/openmarkers/openmarkers-cli/internal/shared/output"
	"github.com/spf13/cobra"
)

var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage categories",
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		var categories []string
		if err := ctx.Client.Get(context.Background(), "/api/categories", &categories); err != nil {
			return handleError(err)
		}
		rows := make([]map[string]string, len(categories))
		for i, c := range categories {
			rows[i] = map[string]string{"id": c}
		}
		return ctx.Output.Output(rows, []output.Column{
			{Title: "Category ID", Key: "id", Width: 30},
		})
	},
}

func init() {
	categoryCmd.AddCommand(categoryListCmd)
	rootCmd.AddCommand(categoryCmd)
}
