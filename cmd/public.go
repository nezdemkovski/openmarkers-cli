package cmd

import (
	"context"

	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/openmarkers/openmarkers-cli/internal/shared/output"
	"github.com/spf13/cobra"
)

var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Browse public profiles (no auth required)",
}

var publicListCmd = &cobra.Command{
	Use:   "list",
	Short: "List public profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		var profiles []models.PublicProfile
		if err := ctx.Client.Get(context.Background(), "/api/public", &profiles); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(profiles, []output.Column{
			{Title: "Name", Key: "name", Width: 25},
			{Title: "Handle", Key: "handle", Width: 25},
		})
	},
}

var publicGetCmd = &cobra.Command{
	Use:   "get <handle>",
	Short: "Get a public profile by handle",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var data models.ProfileData
		if err := ctx.Client.Get(context.Background(), "/api/public/"+args[0], &data); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(data, nil)
	},
}

func init() {
	publicCmd.AddCommand(publicListCmd)
	publicCmd.AddCommand(publicGetCmd)
	rootCmd.AddCommand(publicCmd)
}
