package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <profile_id>",
	Short: "Export profile data as JSON",
	Long:  "Export profile data as JSON to stdout. Perfect for piping to import or saving to a file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}

		data, err := ctx.Client.GetRaw(context.Background(), "/api/profiles/"+url.PathEscape(pid)+"/export")
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
	rootCmd.AddCommand(exportCmd)
}
