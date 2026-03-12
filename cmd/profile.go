package cmd

import (
	"context"
	"fmt"

	"github.com/openmarkers/openmarkers-cli/internal/shared/models"
	"github.com/openmarkers/openmarkers-cli/internal/shared/output"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		var profiles []models.Profile
		if err := ctx.Client.Get(context.Background(), "/api/profiles", &profiles); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(profiles, []output.Column{
			{Title: "ID", Key: "id", Width: 4},
			{Title: "Name", Key: "name", Width: 20},
			{Title: "DOB", Key: "dateOfBirth", Width: 12},
			{Title: "Sex", Key: "sex", Width: 4},
			{Title: "Public", Key: "isPublic", Width: 6},
		})
	},
}

var profileGetCmd = &cobra.Command{
	Use:   "get <profile_id>",
	Short: "Get profile details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		var data models.ProfileData
		if err := ctx.Client.Get(context.Background(), "/api/profiles/"+args[0], &data); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(data, nil)
	},
}

var (
	profileName string
	profileDOB  string
	profileSex  string
)

var profileCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		if profileName == "" {
			return fmt.Errorf("--name is required")
		}
		body := map[string]string{
			"name":          profileName,
			"date_of_birth": profileDOB,
			"sex":           profileSex,
		}
		var profile models.Profile
		if err := ctx.Client.Post(context.Background(), "/api/profiles", body, &profile); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(profile, nil)
	},
}

var (
	profileUpdatePublic bool
	profileUpdateHandle string
)

var profileUpdateCmd = &cobra.Command{
	Use:   "update <profile_id>",
	Short: "Update a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		body := map[string]any{}
		if cmd.Flags().Changed("name") {
			body["name"] = profileName
		}
		if cmd.Flags().Changed("dob") {
			body["date_of_birth"] = profileDOB
		}
		if cmd.Flags().Changed("sex") {
			body["sex"] = profileSex
		}
		if cmd.Flags().Changed("public") {
			body["is_public"] = profileUpdatePublic
		}
		if cmd.Flags().Changed("handle") {
			body["public_handle"] = profileUpdateHandle
		}
		if len(body) == 0 {
			return fmt.Errorf("at least one field to update is required")
		}
		var profile models.Profile
		if err := ctx.Client.Patch(context.Background(), "/api/profiles/"+args[0], body, &profile); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(profile, nil)
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete <profile_id>",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireAuth(); err != nil {
			return err
		}
		var result map[string]any
		if err := ctx.Client.Delete(context.Background(), "/api/profiles/"+args[0], &result); err != nil {
			return handleError(err)
		}
		return ctx.Output.Output(result, nil)
	},
}

func init() {
	profileCreateCmd.Flags().StringVar(&profileName, "name", "", "Profile name")
	profileCreateCmd.Flags().StringVar(&profileDOB, "dob", "", "Date of birth (YYYY-MM-DD)")
	profileCreateCmd.Flags().StringVar(&profileSex, "sex", "", "Sex (M or F)")

	profileUpdateCmd.Flags().StringVar(&profileName, "name", "", "Profile name")
	profileUpdateCmd.Flags().StringVar(&profileDOB, "dob", "", "Date of birth (YYYY-MM-DD)")
	profileUpdateCmd.Flags().StringVar(&profileSex, "sex", "", "Sex (M or F)")
	profileUpdateCmd.Flags().BoolVar(&profileUpdatePublic, "public", false, "Make profile public")
	profileUpdateCmd.Flags().StringVar(&profileUpdateHandle, "handle", "", "Public handle")

	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileGetCmd)
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileUpdateCmd)
	profileCmd.AddCommand(profileDeleteCmd)
	rootCmd.AddCommand(profileCmd)
}
