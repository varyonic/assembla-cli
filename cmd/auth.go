package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/eugene-software/assembla-cli/internal"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Assembla",
	RunE: func(cmd *cobra.Command, args []string) error {
		scope, _ := cmd.Flags().GetString("scope")
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Get your API key/secret at: https://www.assembla.com/user/edit/manage_clients")
		fmt.Println()

		fmt.Print("API Key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)

		fmt.Print("API Secret: ")
		apiSecret, _ := reader.ReadString('\n')
		apiSecret = strings.TrimSpace(apiSecret)

		fmt.Print("\nVerifying credentials... ")

		// Verify credentials
		req, _ := http.NewRequest("GET", "https://api.assembla.com/v1/user", nil)
		req.Header.Set("X-Api-Key", apiKey)
		req.Header.Set("X-Api-Secret", apiSecret)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("FAILED")
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fmt.Println("FAILED")
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Error %d: %s\n", resp.StatusCode, string(body))
			os.Exit(1)
		}

		var user map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &user)

		name := user["name"]
		if name == nil || name == "" {
			name = user["login"]
		}
		fmt.Printf("OK (logged in as %v)\n", name)

		// Fetch spaces
		var spaceID string
		req2, _ := http.NewRequest("GET", "https://api.assembla.com/v1/spaces", nil)
		req2.Header.Set("X-Api-Key", apiKey)
		req2.Header.Set("X-Api-Secret", apiSecret)

		resp2, err := http.DefaultClient.Do(req2)
		if err == nil {
			defer resp2.Body.Close()
			if resp2.StatusCode >= 200 && resp2.StatusCode < 300 {
				body2, _ := io.ReadAll(resp2.Body)
				var spaces []map[string]interface{}
				if json.Unmarshal(body2, &spaces) == nil && len(spaces) > 0 {
					fmt.Printf("\nAvailable spaces (%d):\n", len(spaces))
					for i, s := range spaces {
						fmt.Printf("  %d. %v (%v)\n", i+1, s["name"], s["wiki_name"])
					}

					fmt.Print("\nDefault space (number or wiki_name, Enter to skip): ")
					choice, _ := reader.ReadString('\n')
					choice = strings.TrimSpace(choice)

					if choice != "" {
						if num, err := strconv.Atoi(choice); err == nil && num >= 1 && num <= len(spaces) {
							spaceID = fmt.Sprintf("%v", spaces[num-1]["wiki_name"])
						} else {
							spaceID = choice
						}
					}
				}
			}
		}

		data := map[string]interface{}{
			"api_key":    apiKey,
			"api_secret": apiSecret,
		}
		if spaceID != "" {
			data["space"] = spaceID
		}

		var path string
		if scope == "project" {
			path, err = internal.SaveProjectConfig(data, "")
		} else {
			path, err = internal.SaveGlobalConfig(data)
		}
		if err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("\nCredentials saved to %s\n", path)
		if spaceID != "" {
			fmt.Printf("Default space: %s\n", spaceID)
		}

		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		scope, _ := cmd.Flags().GetString("scope")

		if scope == "project" {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			projectFile := cwd + "/.assembla.yml"
			if _, err := os.Stat(projectFile); err == nil {
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("Remove %s? [y/N]: ", projectFile)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer == "y" || answer == "yes" {
					os.Remove(projectFile)
					fmt.Printf("Removed %s\n", projectFile)
				}
			} else {
				fmt.Println("No project config found in current directory.")
			}
		} else {
			if _, err := os.Stat(internal.GlobalConfigFile); err == nil {
				os.Remove(internal.GlobalConfigFile)
				fmt.Printf("Removed %s\n", internal.GlobalConfigFile)
			} else {
				fmt.Println("No global config found.")
			}
		}

		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		config := internal.LoadConfig()

		apiKey, _ := config["api_key"].(string)
		if apiKey == "" {
			fmt.Println("Not authenticated. Run: assembla auth login")
			return nil
		}

		maskedKey := "****"
		if len(apiKey) > 8 {
			maskedKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
		}

		space := "(not set)"
		if v, ok := config["space"].(string); ok && v != "" {
			space = v
		}

		apiURL, _ := config["api_url"].(string)

		fmt.Printf("API Key:  %s\n", maskedKey)
		fmt.Printf("Space:    %s\n", space)
		fmt.Printf("API URL:  %s\n", apiURL)

		if _, err := os.Stat(internal.GlobalConfigFile); err == nil {
			fmt.Printf("\nGlobal config:  %s\n", internal.GlobalConfigFile)
		}
		if projectConfig, ok := config["_project_config"].(string); ok {
			fmt.Printf("Project config: %s\n", projectConfig)
		}

		fmt.Print("\nVerifying... ")

		apiSecret, _ := config["api_secret"].(string)
		if apiSecret == "" {
			fmt.Println("MISSING API SECRET")
			return nil
		}

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/user", apiURL), nil)
		req.Header.Set("X-Api-Key", apiKey)
		req.Header.Set("X-Api-Secret", apiSecret)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("FAILED (%v)\n", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			body, _ := io.ReadAll(resp.Body)
			var user map[string]interface{}
			json.Unmarshal(body, &user)
			name := user["name"]
			if name == nil || name == "" {
				name = user["login"]
			}
			fmt.Printf("OK (logged in as %v)\n", name)
		} else {
			fmt.Printf("FAILED (%d)\n", resp.StatusCode)
		}

		return nil
	},
}

func init() {
	authLoginCmd.Flags().String("scope", "global", "Config scope: global or project")
	authLogoutCmd.Flags().String("scope", "global", "Config scope: global or project")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
}
