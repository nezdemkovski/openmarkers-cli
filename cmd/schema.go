package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Get biomarker schema (public, no auth required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := ctx.Client.GetRaw(context.Background(), "/schema.json")
		if err != nil {
			return handleError(err)
		}
		var parsed any
		if err := json.Unmarshal(data, &parsed); err != nil {
			os.Stdout.Write(data)
			return nil
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(parsed); err != nil {
			return fmt.Errorf("encode: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(schemaCmd)
}
