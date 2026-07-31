package cmd

import (
	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var milestoneCmd = &cobra.Command{
	Use:   "milestone",
	Short: "Manage milestones",
}

var milestoneListCmd = &cobra.Command{
	Use:   "list",
	Short: "List milestones",
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")
		showAll, _ := cmd.Flags().GetBool("all")
		completed, _ := cmd.Flags().GetBool("completed")

		var path string
		if showAll {
			path = internal.APIPath("spaces", space, "milestones", "all")
		} else if completed {
			path = internal.APIPath("spaces", space, "milestones", "completed")
		} else {
			path = internal.APIPath("spaces", space, "milestones", "upcoming")
		}

		data, err := Client.Get(path, nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintTable(
				internal.ToSlice(data),
				[]string{"id", "title", "due_date", "is_completed", "planner_type"},
				[]string{"ID", "Title", "Due Date", "Completed", "Type"},
			)
		}
		return nil
	},
}

var milestoneShowCmd = &cobra.Command{
	Use:   "show <milestone_id>",
	Short: "Show milestone details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")
		milestoneID := args[0]

		data, err := Client.Get(internal.APIPath("spaces", space, "milestones", milestoneID), nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintDetail(data, []string{
				"id", "title", "description", "due_date",
				"is_completed", "planner_type", "created_at", "updated_at",
			})
		}
		return nil
	},
}

func init() {
	milestoneListCmd.Flags().Bool("all", false, "Show all milestones")
	milestoneListCmd.Flags().Bool("completed", false, "Show completed milestones")
	milestoneListCmd.Flags().Bool("json", false, "Output as JSON")

	milestoneShowCmd.Flags().Bool("json", false, "Output as JSON")

	milestoneCmd.AddCommand(milestoneListCmd)
	milestoneCmd.AddCommand(milestoneShowCmd)
}
