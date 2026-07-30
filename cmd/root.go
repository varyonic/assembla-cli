package cmd

import (
	"fmt"
	"os"

	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	spaceOverride string
	apiKeyFlag    string
	apiSecretFlag string

	// Shared state for subcommands
	Client *internal.AssemblaClient
	Space  string
)

// noAuthCommands are commands that do not require authentication.
var noAuthCommands = map[string]bool{
	"auth": true,
}

var rootCmd = &cobra.Command{
	Use:   "assembla",
	Short: "Assembla CLI - manage tickets, spaces, and more",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip auth check for auth subcommands
		if parent := cmd.Parent(); parent != nil && noAuthCommands[parent.Name()] {
			return nil
		}
		if noAuthCommands[cmd.Name()] {
			return nil
		}

		config := internal.LoadConfig()

		key := apiKeyFlag
		if key == "" {
			if v, ok := config["api_key"].(string); ok {
				key = v
			}
		}
		secret := apiSecretFlag
		if secret == "" {
			if v, ok := config["api_secret"].(string); ok {
				secret = v
			}
		}

		if key == "" || secret == "" {
			fmt.Fprintln(os.Stderr, "Error: Assembla credentials required.")
			fmt.Fprintln(os.Stderr, "Set ASSEMBLA_API_KEY and ASSEMBLA_API_SECRET env vars,")
			fmt.Fprintln(os.Stderr, "or create a .assembla.yml file with api_key and api_secret.")
			fmt.Fprintln(os.Stderr, "Or run: assembla auth login")
			os.Exit(1)
		}

		rawURL := internal.DefaultAPIURL
		if v, ok := config["api_url"].(string); ok {
			rawURL = v
		}
		urlSource, _ := config["_api_url_source"].(string)

		// Validated before the client exists, so credentials cannot reach an
		// unvetted host.
		apiURL, err := internal.ResolveAPIURL(rawURL, urlSource)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		Client = internal.NewAssemblaClient(key, secret, apiURL)

		Space = spaceOverride
		if Space == "" {
			if v, ok := config["space"].(string); ok {
				Space = v
			}
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&spaceOverride, "space", "", "Assembla space (overrides config)")
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Assembla API key")
	rootCmd.PersistentFlags().StringVar(&apiSecretFlag, "api-secret", "", "Assembla API secret")

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(ticketCmd)
	rootCmd.AddCommand(commentCmd)
	rootCmd.AddCommand(spaceCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(milestoneCmd)
	rootCmd.AddCommand(userCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
