package cmd

import (
	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
}

var userMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		asJSON, _ := cmd.Flags().GetBool("json")

		data, err := Client.Get("/user", nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintDetail(data, []string{"id", "login", "name", "email"})
		}
		return nil
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users in the space",
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")

		data, err := Client.Get(internal.APIPath("spaces", space, "users"), nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintTable(
				internal.ToSlice(data),
				[]string{"id", "login", "name", "email"},
				[]string{"ID", "Login", "Name", "Email"},
			)
		}
		return nil
	},
}

func init() {
	userMeCmd.Flags().Bool("json", false, "Output as JSON")
	userListCmd.Flags().Bool("json", false, "Output as JSON")

	userCmd.AddCommand(userMeCmd)
	userCmd.AddCommand(userListCmd)
}
