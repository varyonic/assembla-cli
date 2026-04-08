package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var ticketCmd = &cobra.Command{
	Use:   "ticket",
	Short: "Manage tickets",
}

var ticketListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tickets",
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()

		status, _ := cmd.Flags().GetString("status")
		assignee, _ := cmd.Flags().GetString("assignee")
		milestone, _ := cmd.Flags().GetString("milestone")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")
		asJSON, _ := cmd.Flags().GetBool("json")

		params := map[string]string{
			"page":     strconv.Itoa(page),
			"per_page": strconv.Itoa(perPage),
		}
		if status != "" {
			params["ticket_status"] = status
		}
		if assignee != "" {
			params["assigned_to_id"] = assignee
		}
		if milestone != "" {
			params["milestone_id"] = milestone
		}

		data, err := Client.Get(fmt.Sprintf("/spaces/%s/tickets", space), params)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintTable(
				internal.ToSlice(data),
				[]string{"number", "summary", "status", "assigned_to_id", "priority"},
				[]string{"#", "Summary", "Status", "Assignee", "Priority"},
			)
		}
		return nil
	},
}

var ticketShowCmd = &cobra.Command{
	Use:   "show <number>",
	Short: "Show ticket details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")

		number := args[0]
		data, err := Client.Get(fmt.Sprintf("/spaces/%s/tickets/%s", space, number), nil)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			internal.PrintDetail(data, []string{
				"number", "summary", "description", "status",
				"priority", "assigned_to_id", "milestone_id",
				"created_on", "updated_at",
			})
		}
		return nil
	},
}

var ticketCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new ticket",
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()

		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetInt("priority")
		assignee, _ := cmd.Flags().GetString("assignee")
		milestone, _ := cmd.Flags().GetString("milestone")
		asJSON, _ := cmd.Flags().GetBool("json")

		ticket := map[string]interface{}{
			"summary":     title,
			"description": description,
		}
		if status != "" {
			ticket["status"] = status
		}
		if cmd.Flags().Changed("priority") {
			ticket["priority"] = priority
		}
		if assignee != "" {
			ticket["assigned_to_id"] = assignee
		}
		if milestone != "" {
			ticket["milestone_id"] = milestone
		}

		payload := map[string]interface{}{
			"ticket": ticket,
		}

		data, err := Client.Post(fmt.Sprintf("/spaces/%s/tickets", space), payload)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				fmt.Printf("Created ticket #%v: %v\n", dataMap["number"], dataMap["summary"])
			}
		}
		return nil
	},
}

var ticketUpdateCmd = &cobra.Command{
	Use:   "update <number>",
	Short: "Update an existing ticket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		asJSON, _ := cmd.Flags().GetBool("json")
		number := args[0]

		ticket := map[string]interface{}{}

		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			ticket["summary"] = v
		}
		if cmd.Flags().Changed("description") {
			v, _ := cmd.Flags().GetString("description")
			ticket["description"] = v
		}
		if cmd.Flags().Changed("status") {
			v, _ := cmd.Flags().GetString("status")
			ticket["status"] = v
		}
		if cmd.Flags().Changed("priority") {
			v, _ := cmd.Flags().GetInt("priority")
			ticket["priority"] = v
		}
		if cmd.Flags().Changed("assignee") {
			v, _ := cmd.Flags().GetString("assignee")
			ticket["assigned_to_id"] = v
		}
		if cmd.Flags().Changed("milestone") {
			v, _ := cmd.Flags().GetString("milestone")
			ticket["milestone_id"] = v
		}

		if len(ticket) == 0 {
			fmt.Fprintln(os.Stderr, "Nothing to update.")
			os.Exit(1)
		}

		payload := map[string]interface{}{
			"ticket": ticket,
		}

		data, err := Client.Put(fmt.Sprintf("/spaces/%s/tickets/%s", space, number), payload)
		if err != nil {
			return err
		}

		if asJSON {
			internal.PrintJSON(data)
		} else {
			fmt.Printf("Updated ticket #%s\n", number)
		}
		return nil
	},
}

var ticketMoveCmd = &cobra.Command{
	Use:   "move <number> <status>",
	Short: "Move a ticket to a new status",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		space := requireSpace()
		number := args[0]
		status := args[1]

		payload := map[string]interface{}{
			"ticket": map[string]interface{}{
				"status": status,
			},
		}

		_, err := Client.Put(fmt.Sprintf("/spaces/%s/tickets/%s", space, number), payload)
		if err != nil {
			return err
		}

		fmt.Printf("Ticket #%s moved to \"%s\"\n", number, status)
		return nil
	},
}

func requireSpace() string {
	if Space == "" {
		fmt.Fprintln(os.Stderr, "Error: No space specified. Use --space flag or set in config.")
		os.Exit(1)
	}
	return Space
}

func init() {
	// ticket list flags
	ticketListCmd.Flags().StringP("status", "s", "", "Filter by status")
	ticketListCmd.Flags().StringP("assignee", "a", "", "Filter by assignee ID")
	ticketListCmd.Flags().StringP("milestone", "m", "", "Filter by milestone ID")
	ticketListCmd.Flags().IntP("page", "p", 1, "Page number")
	ticketListCmd.Flags().IntP("per-page", "n", 25, "Results per page")
	ticketListCmd.Flags().Bool("json", false, "Output as JSON")

	// ticket show flags
	ticketShowCmd.Flags().Bool("json", false, "Output as JSON")

	// ticket create flags
	ticketCreateCmd.Flags().StringP("title", "t", "", "Ticket title (required)")
	ticketCreateCmd.MarkFlagRequired("title")
	ticketCreateCmd.Flags().StringP("description", "d", "", "Ticket description")
	ticketCreateCmd.Flags().StringP("status", "s", "", "Ticket status")
	ticketCreateCmd.Flags().IntP("priority", "p", 0, "Ticket priority")
	ticketCreateCmd.Flags().StringP("assignee", "a", "", "Assignee ID")
	ticketCreateCmd.Flags().StringP("milestone", "m", "", "Milestone ID")
	ticketCreateCmd.Flags().Bool("json", false, "Output as JSON")

	// ticket update flags
	ticketUpdateCmd.Flags().StringP("title", "t", "", "Ticket title")
	ticketUpdateCmd.Flags().StringP("description", "d", "", "Ticket description")
	ticketUpdateCmd.Flags().StringP("status", "s", "", "Ticket status")
	ticketUpdateCmd.Flags().IntP("priority", "p", 0, "Ticket priority")
	ticketUpdateCmd.Flags().StringP("assignee", "a", "", "Assignee ID")
	ticketUpdateCmd.Flags().StringP("milestone", "m", "", "Milestone ID")
	ticketUpdateCmd.Flags().Bool("json", false, "Output as JSON")

	ticketCmd.AddCommand(ticketListCmd)
	ticketCmd.AddCommand(ticketShowCmd)
	ticketCmd.AddCommand(ticketCreateCmd)
	ticketCmd.AddCommand(ticketUpdateCmd)
	ticketCmd.AddCommand(ticketMoveCmd)
}
