package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/api"
	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/auth"
	"github.com/openmarkers/openmarkers-cli/internal/infrastructure/config"
	"github.com/openmarkers/openmarkers-cli/internal/shared/output"
)

type cmdContext struct {
	Client    *api.Client
	Output    *output.Writer
	Config    *config.Config
	Store     *auth.Store
	IsJSON    bool
	IsTTY     bool
	Verbose   bool
	ProfileID string
	ServerURL string
}

var (
	ctx       *cmdContext
	jsonFlag  bool
	outputFmt string
	serverURL string
	verbose   bool
	noColor   bool
	profileID string
)

var rootCmd = &cobra.Command{
	Use:   "openmarkers",
	Short: "OpenMarkers CLI — biomarker and blood test tracker",
	Long:  "A CLI for the OpenMarkers biomarker tracking API. Supports JSON output for AI agents and interactive TUI for humans.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		isTTY := isatty.IsTerminal(os.Stdout.Fd())

		isJSON := jsonFlag || outputFmt == "json"
		var formatter output.Formatter
		switch {
		case isJSON || outputFmt == "json":
			formatter = output.NewJSONFormatter()
			isJSON = true
		case outputFmt == "table":
			formatter = output.NewTableFormatter()
		case outputFmt == "text":
			formatter = output.NewTextFormatter(noColor)
		default:
			if isTTY {
				formatter = output.NewTableFormatter()
			} else {
				formatter = output.NewJSONFormatter()
				isJSON = true
			}
		}

		resolvedServer := config.ResolveServer(serverURL, cfg)
		resolvedProfile := config.ResolveProfile(profileID, cfg)

		store := auth.NewStore(config.ConfigDir())
		client := api.NewClient(resolvedServer, store)
		client.Verbose = verbose
		if verbose {
			w := output.NewWriter(formatter)
			client.LogFunc = w.Verbose
		}

		ctx = &cmdContext{
			Client:    client,
			Output:    output.NewWriter(formatter),
			Config:    cfg,
			Store:     store,
			IsJSON:    isJSON,
			IsTTY:     isTTY,
			Verbose:   verbose,
			ProfileID: resolvedProfile,
			ServerURL: resolvedServer,
		}

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().StringVar(&outputFmt, "output", "", "Output format: json, text, table")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "", "Server URL (default: https://openmarkers.app)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Verbose output (debug info to stderr)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().StringVar(&profileID, "profile", "", "Default profile ID")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exitCode := 1
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			exitCode = api.ExitCodeForError(apiErr)
			if ctx != nil && ctx.IsJSON {
				_ = ctx.Output.Error(apiErr.Code, apiErr.Message)
				os.Exit(exitCode)
			}
		}

		if ctx != nil && ctx.IsJSON {
			_ = ctx.Output.Error("error", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		os.Exit(exitCode)
	}
}

func handleError(err error) error {
	if err == nil {
		return nil
	}
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		if ctx != nil && ctx.IsJSON {
			_ = ctx.Output.Error(apiErr.Code, apiErr.Message)
			os.Exit(api.ExitCodeForError(apiErr))
		}
	}
	return err
}

func requireAuth() error {
	if !ctx.Store.HasTokens() {
		err := &api.APIError{
			StatusCode: 401,
			Code:       "auth_required",
			Message:    "Not logged in. Run 'openmarkers auth login' first.",
			Err:        api.ErrAuthRequired,
		}
		return err
	}
	return nil
}

func requireProfile(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if ctx.ProfileID != "" {
		return ctx.ProfileID, nil
	}
	return "", fmt.Errorf("profile ID required: provide as argument or use --profile flag")
}
