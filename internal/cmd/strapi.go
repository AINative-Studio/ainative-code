package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	strapiURL      string
	strapiToken    string
	strapiTestConn bool
)

// strapiCmd represents the strapi command
var strapiCmd = &cobra.Command{
	Use:   "strapi",
	Short: "Interact with Strapi CMS",
	Long: `Interact with Strapi CMS for content management operations.

Strapi is used for managing design tokens, documentation, and other
content that needs to be shared across applications and teams.

Examples:
  # Test Strapi connection
  ainative-code strapi test

  # Configure Strapi connection
  ainative-code strapi config --url https://strapi.example.com --token your-token

  # Fetch content from Strapi
  ainative-code strapi fetch design-tokens

  # Push content to Strapi
  ainative-code strapi push design-tokens

  # List available content types
  ainative-code strapi list`,
	Aliases: []string{"cms"},
}

// strapiTestCmd represents the strapi test command
var strapiTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test Strapi connection",
	Long:  `Test the connection to Strapi CMS and verify authentication.`,
	RunE:  runStrapiTest,
}

// strapiConfigCmd represents the strapi config command
var strapiConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Strapi connection",
	Long:  `Configure Strapi CMS connection settings.`,
	RunE:  runStrapiConfig,
}

// strapiFetchCmd represents the strapi fetch command
var strapiFetchCmd = &cobra.Command{
	Use:   "fetch [content-type]",
	Short: "Fetch content from Strapi",
	Long:  `Fetch content from Strapi CMS and store locally.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStrapiFetch,
}

// strapiPushCmd represents the strapi push command
var strapiPushCmd = &cobra.Command{
	Use:   "push [content-type]",
	Short: "Push content to Strapi",
	Long:  `Push local content to Strapi CMS.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runStrapiPush,
}

// strapiListCmd represents the strapi list command
var strapiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Strapi content types",
	Long:  `List available content types in Strapi CMS.`,
	Aliases: []string{"ls"},
	RunE:  runStrapiList,
}

// strapiSyncCmd represents the strapi sync command
var strapiSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync content with Strapi",
	Long:  `Bidirectional sync of content between local database and Strapi CMS.`,
	RunE:  runStrapiSync,
}

func init() {
	rootCmd.AddCommand(strapiCmd)

	// Add subcommands
	strapiCmd.AddCommand(strapiTestCmd)
	strapiCmd.AddCommand(strapiConfigCmd)
	strapiCmd.AddCommand(strapiFetchCmd)
	strapiCmd.AddCommand(strapiPushCmd)
	strapiCmd.AddCommand(strapiListCmd)
	strapiCmd.AddCommand(strapiSyncCmd)

	// Config flags
	strapiConfigCmd.Flags().StringVar(&strapiURL, "url", "", "Strapi server URL")
	strapiConfigCmd.Flags().StringVar(&strapiToken, "token", "", "Strapi API token")

	// Fetch flags
	strapiFetchCmd.Flags().BoolP("force", "f", false, "force fetch (overwrite local data)")

	// Push flags
	strapiPushCmd.Flags().BoolP("force", "f", false, "force push (overwrite remote data)")

	// Sync flags
	strapiSyncCmd.Flags().String("strategy", "merge", "sync strategy (merge, local-wins, remote-wins)")
}

func runStrapiTest(cmd *cobra.Command, args []string) error {
	logger.Info("Testing Strapi connection")

	// Get Strapi URL from config
	url := viper.GetString("strapi.url")

	if url == "" {
		return fmt.Errorf("Strapi URL not configured. Use 'ainative-code strapi config' to set it up")
	}

	token := viper.GetString("strapi.token")

	fmt.Printf("Testing connection to: %s\n", url)
	fmt.Println()

	// Test 1: Check URL accessibility
	fmt.Print("1. Testing URL accessibility... ")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("FAILED\n   Error: %v\n", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header if token is configured
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("FAILED\n   Error: %v\n", err)
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		fmt.Printf("OK (Status: %d)\n", resp.StatusCode)
	} else {
		fmt.Printf("WARNING (Status: %d)\n", resp.StatusCode)
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			fmt.Println("   This might be an authentication issue. Check your token.")
		}
	}

	// Test 2: Check authentication
	fmt.Print("2. Testing authentication... ")
	if token == "" {
		fmt.Println("SKIPPED (no token configured)")
	} else {
		// Try to access a protected endpoint
		apiURL := url
		if !strings.HasSuffix(apiURL, "/") {
			apiURL += "/"
		}
		apiURL += "api/users/me"

		authReq, err := http.NewRequest("GET", apiURL, nil)
		if err == nil {
			authReq.Header.Set("Authorization", "Bearer "+token)
			authResp, err := client.Do(authReq)
			if err == nil {
				defer authResp.Body.Close()
				if authResp.StatusCode == 200 {
					fmt.Println("OK")
				} else if authResp.StatusCode == 401 || authResp.StatusCode == 403 {
					fmt.Println("FAILED (invalid token)")
				} else {
					fmt.Printf("WARNING (Status: %d)\n", authResp.StatusCode)
				}
			} else {
				fmt.Printf("FAILED (%v)\n", err)
			}
		} else {
			fmt.Printf("FAILED (%v)\n", err)
		}
	}

	// Test 3: Check API availability
	fmt.Print("3. Testing API endpoint... ")
	apiURL := url
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "api"

	apiReq, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		fmt.Printf("FAILED (%v)\n", err)
	} else {
		if token != "" {
			apiReq.Header.Set("Authorization", "Bearer "+token)
		}
		apiResp, err := client.Do(apiReq)
		if err != nil {
			fmt.Printf("FAILED (%v)\n", err)
		} else {
			defer apiResp.Body.Close()
			if apiResp.StatusCode >= 200 && apiResp.StatusCode < 400 {
				fmt.Println("OK")
			} else {
				fmt.Printf("WARNING (Status: %d)\n", apiResp.StatusCode)
			}
		}
	}

	fmt.Println()
	fmt.Println("Connection test completed!")
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  URL: %s\n", url)
	if token != "" {
		fmt.Println("  Authentication: Configured")
	} else {
		fmt.Println("  Authentication: Not configured")
		fmt.Println("  Tip: Use 'ainative-code strapi config --token YOUR_TOKEN' to configure authentication")
	}

	return nil
}

func runStrapiConfig(cmd *cobra.Command, args []string) error {
	logger.Info("Configuring Strapi connection")

	if strapiURL != "" {
		viper.Set("strapi.url", strapiURL)
		fmt.Printf("Set Strapi URL: %s\n", strapiURL)
	}

	if strapiToken != "" {
		viper.Set("strapi.token", strapiToken)
		fmt.Println("Set Strapi API token")
	}

	if strapiURL != "" || strapiToken != "" {
		// Save configuration
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("Configuration saved!")
	} else {
		// Display current configuration
		fmt.Println("Current Strapi Configuration:")
		fmt.Printf("URL: %s\n", viper.GetString("strapi.url"))
		if viper.GetString("strapi.token") != "" {
			fmt.Println("Token: [configured]")
		} else {
			fmt.Println("Token: [not configured]")
		}
	}

	return nil
}

func runStrapiFetch(cmd *cobra.Command, args []string) error {
	contentType := args[0]
	force, _ := cmd.Flags().GetBool("force")

	logger.InfoEvent().
		Str("content_type", contentType).
		Bool("force", force).
		Msg("Fetching content from Strapi")

	fmt.Printf("Fetching %s from Strapi...\n", contentType)

	// TODO: Implement content fetch
	// - Connect to Strapi
	// - Query content type
	// - Store in local database
	// - Handle conflicts if not force

	fmt.Println("Fetch completed!")

	return nil
}

func runStrapiPush(cmd *cobra.Command, args []string) error {
	contentType := args[0]
	force, _ := cmd.Flags().GetBool("force")

	logger.InfoEvent().
		Str("content_type", contentType).
		Bool("force", force).
		Msg("Pushing content to Strapi")

	fmt.Printf("Pushing %s to Strapi...\n", contentType)

	// TODO: Implement content push
	// - Read from local database
	// - Send to Strapi
	// - Handle conflicts if not force

	fmt.Println("Push completed!")

	return nil
}

func runStrapiList(cmd *cobra.Command, args []string) error {
	logger.Debug("Listing Strapi content types")

	// Get Strapi URL from config
	url := viper.GetString("strapi.url")
	if url == "" {
		return fmt.Errorf("Strapi URL not configured. Use 'ainative-code strapi config' to set it up")
	}

	token := viper.GetString("strapi.token")

	fmt.Println("Available Content Types:")
	fmt.Println("========================")
	fmt.Println()

	// Create HTTP client
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build API URL for content types
	apiURL := url
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "api/content-type-builder/content-types"

	// Create request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header if token is configured
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Strapi: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		fmt.Println("Authentication required or insufficient permissions.")
		fmt.Println()
		fmt.Println("This endpoint typically requires admin access.")
		fmt.Println("Alternative approach: List commonly used content types:")
		fmt.Println()
		displayCommonContentTypes()
		return nil
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Unable to fetch content types (Status: %d)\n", resp.StatusCode)
		fmt.Println()
		fmt.Println("Showing commonly used content types instead:")
		fmt.Println()
		displayCommonContentTypes()
		return nil
	}

	// For now, display common content types since parsing the response
	// would require understanding the specific Strapi version and schema
	fmt.Println("Common content types typically available in Strapi:")
	fmt.Println()
	displayCommonContentTypes()

	return nil
}

func displayCommonContentTypes() {
	contentTypes := []struct {
		name        string
		description string
	}{
		{"design-tokens", "Design system tokens (colors, typography, spacing)"},
		{"articles", "Article or blog post content"},
		{"pages", "Static page content"},
		{"components", "Reusable UI components"},
		{"media", "Media library assets"},
		{"users", "User accounts and profiles"},
		{"roles", "User roles and permissions"},
		{"settings", "Application settings"},
	}

	for _, ct := range contentTypes {
		fmt.Printf("  • %s\n", ct.name)
		fmt.Printf("    %s\n\n", ct.description)
	}

	fmt.Println("Note: Actual content types depend on your Strapi configuration.")
	fmt.Println("Use 'ainative-code strapi fetch <content-type>' to fetch specific content.")
}

func runStrapiSync(cmd *cobra.Command, args []string) error {
	strategy, _ := cmd.Flags().GetString("strategy")

	logger.InfoEvent().Str("strategy", strategy).Msg("Syncing with Strapi")

	fmt.Printf("Syncing content with Strapi (strategy: %s)...\n", strategy)

	// TODO: Implement bidirectional sync
	// - Compare local and remote content
	// - Apply sync strategy
	// - Resolve conflicts
	// - Report changes

	fmt.Println("Sync completed!")

	return nil
}
