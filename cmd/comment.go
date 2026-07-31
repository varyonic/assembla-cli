package cmd

import (
	"fmt"

	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage ticket comments",
}

var commentListCmd = &cobra.Command{
	Use:   "list <ticket_number>",
	Short: "List comments on a ticket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")
		ticketNumber := args[0]

		data, err := Client.Get(internal.APIPath("spaces", space, "tickets", ticketNumber, "ticket_comments"), nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintTable(
				internal.ToSlice(data),
				[]string{"id", "author_name", "comment", "created_on"},
				[]string{"ID", "Author", "Comment", "Created"},
			)
		}
		return nil
	},
}

var commentAddCmd = &cobra.Command{
	Use:   "add <ticket_number> <body>",
	Short: "Add a comment to a ticket",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")
		ticketNumber := args[0]
		body := args[1]

		payload := map[string]interface{}{
			"ticket_comment": map[string]interface{}{
				"comment": body,
			},
		}

		data, err := Client.Post(internal.APIPath("spaces", space, "tickets", ticketNumber, "ticket_comments"), payload)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			fmt.Printf("Comment added to ticket #%s\n", ticketNumber)
		}
		return nil
	},
}

func init() {
	commentListCmd.Flags().Bool("json", false, "Output as JSON")
	commentAddCmd.Flags().Bool("json", false, "Output as JSON")

	commentCmd.AddCommand(commentListCmd)
	commentCmd.AddCommand(commentAddCmd)
}
