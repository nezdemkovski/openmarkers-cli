package cmd

import (
	"github.com/openmarkers/openmarkers-cli/internal/presentation"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		return presentation.Run(ctx.Client)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
