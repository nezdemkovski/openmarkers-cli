package cmd

import (
	"context"

	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/openmarkers/openmarkers-cli/internal/shared/output"
	"github.com/spf13/cobra"
)

var timelineCmd = &cobra.Command{
	Use:   "timeline [profile_id]",
	Short: "Get timeline of test dates",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}
		var dates []string
		if err := ctx.Client.Get(context.Background(), "/api/profiles/"+pid+"/timeline", &dates); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(dates, []output.Column{
			{Title: "Date", Width: 12},
		})
	},
}

var snapshotDate string

var snapshotCmd = &cobra.Command{
	Use:   "snapshot [profile_id]",
	Short: "Get biomarker values on a specific date",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}
		var entries []models.SnapshotEntry
		path := "/api/profiles/" + pid + "/snapshot?date=" + snapshotDate
		if err := ctx.Client.Get(context.Background(), path, &entries); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(entries, []output.Column{
			{Title: "Category", Key: "categoryId", Width: 15},
			{Title: "Biomarker", Key: "biomarkerId", Width: 25},
			{Title: "Value", Key: "value", Width: 10},
			{Title: "Unit", Key: "unit", Width: 10},
			{Title: "Out of Range", Key: "outOfRange", Width: 12},
		})
	},
}

var (
	trendsBiomarker string
	trendsCategory  string
)

var trendsCmd = &cobra.Command{
	Use:   "trends [profile_id]",
	Short: "Analyze biomarker trends",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}
		path := "/api/profiles/" + pid + "/trends"
		sep := "?"
		if trendsBiomarker != "" {
			path += sep + "biomarker_id=" + trendsBiomarker
			sep = "&"
		}
		if trendsCategory != "" {
			path += sep + "category_id=" + trendsCategory
		}
		var entries []models.TrendEntry
		if err := ctx.Client.Get(context.Background(), path, &entries); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(entries, []output.Column{
			{Title: "Biomarker", Key: "biomarkerId", Width: 25},
			{Title: "Direction", Key: "direction", Width: 10},
			{Title: "Change", Key: "overallChange", Width: 10},
			{Title: "Warning", Key: "trendWarning", Width: 8},
			{Title: "Latest", Key: "latestValue", Width: 10},
			{Title: "Date", Key: "latestDate", Width: 12},
		})
	},
}

var (
	compareDate1 string
	compareDate2 string
)

var compareCmd = &cobra.Command{
	Use:   "compare [profile_id]",
	Short: "Compare biomarker values between two dates",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}
		path := "/api/profiles/" + pid + "/compare?date1=" + compareDate1 + "&date2=" + compareDate2
		var entries []models.CompareEntry
		if err := ctx.Client.Get(context.Background(), path, &entries); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(entries, []output.Column{
			{Title: "Biomarker", Key: "biomarkerId", Width: 25},
			{Title: "V1", Key: "v1", Width: 10},
			{Title: "V2", Key: "v2", Width: 10},
			{Title: "Delta", Key: "delta", Width: 10},
			{Title: "Delta %", Key: "deltaPct", Width: 8},
		})
	},
}

var correlationsCmd = &cobra.Command{
	Use:   "correlations [profile_id]",
	Short: "Get biomarker correlations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}
		var entries []models.CorrelationGroup
		if err := ctx.Client.Get(context.Background(), "/api/profiles/"+pid+"/correlations", &entries); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(entries, []output.Column{
			{Title: "ID", Key: "id", Width: 25},
			{Title: "Matched Biomarkers", Key: "matched", Width: 50},
		})
	},
}

var bioageCmd = &cobra.Command{
	Use:   "bioage [profile_id]",
	Short: "Calculate biological age",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}
		var entries []models.BioAgeEntry
		if err := ctx.Client.Get(context.Background(), "/api/profiles/"+pid+"/biological-age", &entries); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(entries, nil)
	},
}

var analysisLang string

var analysisCmd = &cobra.Command{
	Use:   "analysis [profile_id]",
	Short: "Generate AI analysis prompt",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		pid, err := requireProfile(args)
		if err != nil {
			return err
		}
		path := "/api/profiles/" + pid + "/analysis-prompt"
		if analysisLang != "" {
			path += "?lang=" + analysisLang
		}
		var result models.AnalysisPrompt
		if err := ctx.Client.Get(context.Background(), path, &result); err != nil {
			return handleError(err)
		}
		if ctx.IsJSON {
			return ctx.Output.Output(result, nil)
		}
		return ctx.Output.Output(result.Prompt, nil)
	},
}

func init() {
	snapshotCmd.Flags().StringVar(&snapshotDate, "date", "", "Date (YYYY-MM-DD, required)")
	_ = snapshotCmd.MarkFlagRequired("date")

	trendsCmd.Flags().StringVar(&trendsBiomarker, "biomarker", "", "Filter by biomarker")
	trendsCmd.Flags().StringVar(&trendsCategory, "category", "", "Filter by category")

	compareCmd.Flags().StringVar(&compareDate1, "date1", "", "First date (required)")
	compareCmd.Flags().StringVar(&compareDate2, "date2", "", "Second date (required)")
	_ = compareCmd.MarkFlagRequired("date1")
	_ = compareCmd.MarkFlagRequired("date2")

	analysisCmd.Flags().StringVar(&analysisLang, "lang", "", "Language (en, cs, ru, is)")

	rootCmd.AddCommand(timelineCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(trendsCmd)
	rootCmd.AddCommand(compareCmd)
	rootCmd.AddCommand(correlationsCmd)
	rootCmd.AddCommand(bioageCmd)
	rootCmd.AddCommand(analysisCmd)
}
