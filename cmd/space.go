package cmd

import (
	"fmt"

	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var spaceCmd = &cobra.Command{
	Use:   "space",
	Short: "Manage spaces",
}

var spaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available spaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")

		data, err := Client.Get("/spaces", nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintTable(
				internal.ToSlice(data),
				[]string{"id", "wiki_name", "name", "status"},
				[]string{"ID", "Wiki Name", "Name", "Status"},
			)
		}
		return nil
	},
}

var spaceShowCmd = &cobra.Command{
	Use:   "show [space_id]",
	Short: "Show space details",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")

		spaceID := Space
		if len(args) > 0 {
			spaceID = args[0]
		}
		if spaceID == "" {
			return fmt.Errorf("no space specified; use --space flag or provide space_id argument")
		}

		data, err := Client.Get(fmt.Sprintf("/spaces/%s", spaceID), nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintDetail(data, []string{
				"id", "wiki_name", "name", "status",
				"created_at", "updated_at", "description",
			})
		}
		return nil
	},
}

func init() {
	spaceListCmd.Flags().Bool("json", false, "Output as JSON")
	spaceShowCmd.Flags().Bool("json", false, "Output as JSON")

	spaceCmd.AddCommand(spaceListCmd)
	spaceCmd.AddCommand(spaceShowCmd)
}
