package cmd

import (
	"fmt"

	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Manage ticket statuses",
}

var statusListCmd = &cobra.Command{
	Use:   "list",
	Short: "List ticket statuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")

		data, err := Client.Get(fmt.Sprintf("/spaces/%s/tickets/statuses", space), nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintTable(
				internal.ToSlice(data),
				[]string{"id", "name", "state", "list_order"},
				[]string{"ID", "Name", "State", "Order"},
			)
		}
		return nil
	},
}

func init() {
	statusListCmd.Flags().Bool("json", false, "Output as JSON")

	statusCmd.AddCommand(statusListCmd)
}
