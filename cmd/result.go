package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/openmarkers/openmarkers-cli/internal/shared/output"
	"github.com/spf13/cobra"
)

var resultCmd = &cobra.Command{
	Use:   "result",
	Short: "Manage test results",
}

var (
	resultProfile   string
	resultCategory  string
	resultBiomarker string
	resultDateFrom  string
	resultDateTo    string
)

var resultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List results",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid := resultProfile
		if pid == "" {
			pid = ctx.ProfileID
		}
		if pid == "" {
			return fmt.Errorf("--profile is required")
		}
		path := "/api/results?profile_id=" + url.QueryEscape(pid)
		if resultCategory != "" {
			path += "&category_id=" + url.QueryEscape(resultCategory)
		}
		if resultBiomarker != "" {
			path += "&biomarker_id=" + url.QueryEscape(resultBiomarker)
		}
		if resultDateFrom != "" {
			path += "&date_from=" + url.QueryEscape(resultDateFrom)
		}
		if resultDateTo != "" {
			path += "&date_to=" + url.QueryEscape(resultDateTo)
		}
		var results []models.Result
		if err := ctx.Client.Get(context.Background(), path, &results); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(results, []output.Column{
			{Title: "ID", Key: "id", Width: 6},
			{Title: "Biomarker", Key: "biomarker_id", Width: 25},
			{Title: "Date", Key: "date", Width: 12},
			{Title: "Value", Key: "value", Width: 10},
		})
	},
}

var (
	resultAddProfile   string
	resultAddBiomarker string
	resultAddDate      string
	resultAddValue     string
)

var resultAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a result",
	Long:  "Add a result. Supports stdin JSON: echo '{\"profile_id\":1,\"biomarker_id\":\"glucose\",\"date\":\"2024-01-01\",\"value\":90}' | openmarkers result add",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		var body any

		stat, err := os.Stdin.Stat()
		if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			if len(data) > 0 {
				var parsed any
				if err := json.Unmarshal(data, &parsed); err != nil {
					return fmt.Errorf("invalid JSON from stdin: %w", err)
				}
				body = parsed
			}
		}

		if body == nil {
			pid := resultAddProfile
			if pid == "" {
				pid = ctx.ProfileID
			}
			if pid == "" {
				return fmt.Errorf("--profile is required")
			}
			if resultAddBiomarker == "" || resultAddDate == "" || resultAddValue == "" {
				return fmt.Errorf("--biomarker, --date, and --value are required")
			}
			body = map[string]any{
				"profile_id":  pid,
				"biomarker_id": resultAddBiomarker,
				"date":         resultAddDate,
				"value":        resultAddValue,
			}
		}

		var result models.Result
		if err := ctx.Client.Post(context.Background(), "/api/results", body, &result); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(result, nil)
	},
}

var (
	batchAddProfile string
	batchAddDate    string
	batchAddFile    string
)

var resultBatchAddCmd = &cobra.Command{
	Use:   "batch-add",
	Short: "Add multiple results at once",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}

		var body any

		var data []byte
		var err error
		if batchAddFile != "" && batchAddFile != "-" {
			data, err = os.ReadFile(batchAddFile)
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
		} else {
			stat, err := os.Stdin.Stat()
			if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
				data, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
			}
		}

		if data != nil && len(data) > 0 {
			if err := json.Unmarshal(data, &body); err != nil {
				return fmt.Errorf("invalid JSON: %w", err)
			}
		} else {
			return fmt.Errorf("provide input via --file or stdin")
		}

		if entries, ok := body.([]any); ok {
			pid := batchAddProfile
			if pid == "" {
				pid = ctx.ProfileID
			}
			if pid == "" {
				return fmt.Errorf("--profile is required when providing entries array")
			}
			if batchAddDate == "" {
				return fmt.Errorf("--date is required when providing entries array")
			}
			body = map[string]any{
				"profile_id": pid,
				"date":       batchAddDate,
				"entries":    entries,
			}
		}

		var result models.BatchResult
		if err := ctx.Client.Post(context.Background(), "/api/batch-results", body, &result); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(result, nil)
	},
}

var (
	resultUpdateDate  string
	resultUpdateValue string
)

var resultUpdateCmd = &cobra.Command{
	Use:   "update <result_id>",
	Short: "Update a result",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		body := map[string]any{}
		if cmd.Flags().Changed("date") {
			body["date"] = resultUpdateDate
		}
		if cmd.Flags().Changed("value") {
			body["value"] = resultUpdateValue
		}
		if len(body) == 0 {
			return fmt.Errorf("at least one field to update is required")
		}
		var result models.Result
		if err := ctx.Client.Patch(context.Background(), "/api/results/"+url.PathEscape(args[0]), body, &result); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(result, nil)
	},
}

var resultDeleteCmd = &cobra.Command{
	Use:   "delete <result_id>",
	Short: "Delete a result",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		var result map[string]any
		if err := ctx.Client.Delete(context.Background(), "/api/results/"+url.PathEscape(args[0]), &result); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(result, nil)
	},
}

func init() {
	resultListCmd.Flags().StringVar(&resultProfile, "profile", "", "Profile ID (required)")
	resultListCmd.Flags().StringVar(&resultCategory, "category", "", "Filter by category")
	resultListCmd.Flags().StringVar(&resultBiomarker, "biomarker", "", "Filter by biomarker")
	resultListCmd.Flags().StringVar(&resultDateFrom, "date-from", "", "Filter from date (YYYY-MM-DD)")
	resultListCmd.Flags().StringVar(&resultDateTo, "date-to", "", "Filter to date (YYYY-MM-DD)")

	resultAddCmd.Flags().StringVar(&resultAddProfile, "profile", "", "Profile ID")
	resultAddCmd.Flags().StringVar(&resultAddBiomarker, "biomarker", "", "Biomarker ID")
	resultAddCmd.Flags().StringVar(&resultAddDate, "date", "", "Date (YYYY-MM-DD)")
	resultAddCmd.Flags().StringVar(&resultAddValue, "value", "", "Value")

	resultBatchAddCmd.Flags().StringVar(&batchAddProfile, "profile", "", "Profile ID")
	resultBatchAddCmd.Flags().StringVar(&batchAddDate, "date", "", "Date (YYYY-MM-DD)")
	resultBatchAddCmd.Flags().StringVar(&batchAddFile, "file", "", "JSON file with entries")

	resultUpdateCmd.Flags().StringVar(&resultUpdateDate, "date", "", "Date (YYYY-MM-DD)")
	resultUpdateCmd.Flags().StringVar(&resultUpdateValue, "value", "", "Value")

	resultCmd.AddCommand(resultListCmd)
	resultCmd.AddCommand(resultAddCmd)
	resultCmd.AddCommand(resultBatchAddCmd)
	resultCmd.AddCommand(resultUpdateCmd)
	resultCmd.AddCommand(resultDeleteCmd)
	rootCmd.AddCommand(resultCmd)
}
