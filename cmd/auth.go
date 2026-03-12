package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in via OAuth (opens browser)",
	RunE: func(cmd *cobra.Command, args []string) error {
		oauth := &auth.OAuthConfig{
			ServerURL: ctx.ServerURL,
			Store:     ctx.Store,
		}
		if ctx.Verbose {
			oauth.LogFunc = ctx.Output.Verbose
		}
		return oauth.Login(context.Background())
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out (delete stored tokens)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ctx.Store.Delete(); err != nil {
			return fmt.Errorf("logout: %w", err)
		}
		if ctx.IsJSON {
			return ctx.Output.Output(map[string]bool{"ok": true}, nil)
		}
		fmt.Println("Logged out.")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens := ctx.Store.Load()
		if tokens == nil || tokens.AccessToken == "" {
			if ctx.IsJSON {
				return ctx.Output.Output(map[string]any{
					"authenticated": false,
				}, nil)
			}
			fmt.Println("Not logged in.")
			return nil
		}

		status := map[string]any{
			"authenticated": true,
			"server":        ctx.ServerURL,
		}

		if tokens.ExpiresAt > 0 {
			expiresAt := time.Unix(tokens.ExpiresAt, 0)
			status["expires_at"] = expiresAt.Format(time.RFC3339)
			status["expired"] = time.Now().After(expiresAt)
		}

		if ctx.IsJSON {
			return ctx.Output.Output(status, nil)
		}

		fmt.Println("Logged in.")
		fmt.Printf("Server: %s\n", ctx.ServerURL)
		if tokens.ExpiresAt > 0 {
			expiresAt := time.Unix(tokens.ExpiresAt, 0)
			if time.Now().After(expiresAt) {
				fmt.Printf("Token expired at %s (will refresh on next request)\n", expiresAt.Format(time.RFC3339))
			} else {
				fmt.Printf("Token expires at %s\n", expiresAt.Format(time.RFC3339))
			}
		}
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
