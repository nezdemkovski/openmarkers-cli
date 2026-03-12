package cmd

import (
	"context"
	"fmt"

	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/openmarkers/openmarkers-cli/internal/shared/output"
	"github.com/spf13/cobra"
)

var biomarkerCmd = &cobra.Command{
	Use:   "biomarker",
	Short: "Manage biomarkers",
}

var biomarkerCategoryFilter string

var biomarkerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List biomarkers",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		path := "/api/biomarkers"
		if biomarkerCategoryFilter != "" {
			path += "?category_id=" + biomarkerCategoryFilter
		}
		var biomarkers []models.Biomarker
		if err := ctx.Client.Get(context.Background(), path, &biomarkers); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(biomarkers, []output.Column{
			{Title: "ID", Key: "id", Width: 25},
			{Title: "Category", Key: "category_id", Width: 15},
			{Title: "Unit", Key: "unit", Width: 10},
			{Title: "Type", Key: "type", Width: 12},
		})
	},
}

var biomarkerGetCmd = &cobra.Command{
	Use:   "get <biomarker_id>",
	Short: "Get biomarker details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		var biomarkers []models.Biomarker
		if err := ctx.Client.Get(context.Background(), "/api/biomarkers", &biomarkers); err != nil {
			return handleError(err)
		}
		for _, b := range biomarkers {
			if b.ID == args[0] {
				return ctx.Output.Output(b, nil)
			}
		}
		return fmt.Errorf("biomarker '%s' not found", args[0])
	},
}

var (
	biomarkerCreateID       string
	biomarkerCreateCategory string
	biomarkerCreateUnit     string
	biomarkerCreateRefMin   float64
	biomarkerCreateRefMax   float64
	biomarkerCreateType     string
)

var biomarkerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new biomarker",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		if biomarkerCreateID == "" || biomarkerCreateCategory == "" {
			return fmt.Errorf("--id and --category are required")
		}
		body := map[string]any{
			"id":          biomarkerCreateID,
			"category_id": biomarkerCreateCategory,
		}
		if cmd.Flags().Changed("unit") {
			body["unit"] = biomarkerCreateUnit
		}
		if cmd.Flags().Changed("ref-min") {
			body["ref_min"] = biomarkerCreateRefMin
		}
		if cmd.Flags().Changed("ref-max") {
			body["ref_max"] = biomarkerCreateRefMax
		}
		if cmd.Flags().Changed("type") {
			body["type"] = biomarkerCreateType
		}
		var biomarker models.Biomarker
		if err := ctx.Client.Post(context.Background(), "/api/biomarkers", body, &biomarker); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(biomarker, nil)
	},
}

var (
	biomarkerUpdateUnit   string
	biomarkerUpdateRefMin float64
	biomarkerUpdateRefMax float64
)

var biomarkerUpdateCmd = &cobra.Command{
	Use:   "update <biomarker_id>",
	Short: "Update a biomarker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		body := map[string]any{}
		if cmd.Flags().Changed("unit") {
			body["unit"] = biomarkerUpdateUnit
		}
		if cmd.Flags().Changed("ref-min") {
			body["ref_min"] = biomarkerUpdateRefMin
		}
		if cmd.Flags().Changed("ref-max") {
			body["ref_max"] = biomarkerUpdateRefMax
		}
		if len(body) == 0 {
			return fmt.Errorf("at least one field to update is required")
		}
		var biomarker models.Biomarker
		if err := ctx.Client.Patch(context.Background(), "/api/biomarkers/"+args[0], body, &biomarker); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(biomarker, nil)
	},
}

func init() {
	biomarkerListCmd.Flags().StringVar(&biomarkerCategoryFilter, "category", "", "Filter by category")

	biomarkerCreateCmd.Flags().StringVar(&biomarkerCreateID, "id", "", "Biomarker ID")
	biomarkerCreateCmd.Flags().StringVar(&biomarkerCreateCategory, "category", "", "Category ID")
	biomarkerCreateCmd.Flags().StringVar(&biomarkerCreateUnit, "unit", "", "Unit of measurement")
	biomarkerCreateCmd.Flags().Float64Var(&biomarkerCreateRefMin, "ref-min", 0, "Reference range minimum")
	biomarkerCreateCmd.Flags().Float64Var(&biomarkerCreateRefMax, "ref-max", 0, "Reference range maximum")
	biomarkerCreateCmd.Flags().StringVar(&biomarkerCreateType, "type", "quantitative", "Type (quantitative or qualitative)")

	biomarkerUpdateCmd.Flags().StringVar(&biomarkerUpdateUnit, "unit", "", "Unit of measurement")
	biomarkerUpdateCmd.Flags().Float64Var(&biomarkerUpdateRefMin, "ref-min", 0, "Reference range minimum")
	biomarkerUpdateCmd.Flags().Float64Var(&biomarkerUpdateRefMax, "ref-max", 0, "Reference range maximum")

	biomarkerCmd.AddCommand(biomarkerListCmd)
	biomarkerCmd.AddCommand(biomarkerGetCmd)
	biomarkerCmd.AddCommand(biomarkerCreateCmd)
	biomarkerCmd.AddCommand(biomarkerUpdateCmd)
	rootCmd.AddCommand(biomarkerCmd)
}
