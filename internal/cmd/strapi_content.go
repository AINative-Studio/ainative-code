package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/strapi"
	"github.com/AINative-studio/ainative-code/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	contentType   string
	contentData   string
	contentID     int
	contentFilter string
	contentSchema string
)

// strapiContentCmd represents the strapi content command
var strapiContentCmd = &cobra.Command{
	Use:   "content",
	Short: "Manage generic Strapi content",
	Long: `Manage generic content entries in Strapi CMS.

This command allows you to work with any content type in Strapi,
not just blog posts. You can create, list, update, and delete
entries from any content type.

Examples:
  # Create a content type
  ainative-code strapi content create-type \
    --name articles \
    --schema '{"title":"string","body":"text"}'

  # Create a content entry
  ainative-code strapi content create \
    --type articles \
    --data '{"title":"My Article","body":"Content here"}'

  # List content entries
  ainative-code strapi content list \
    --type articles \
    --filter status=published

  # Update a content entry
  ainative-code strapi content update \
    --type articles \
    --id 123 \
    --data '{"title":"Updated Title"}'

  # Delete a content entry
  ainative-code strapi content delete \
    --type articles \
    --id 123`,
}

// strapiContentCreateTypeCmd creates a new content type
var strapiContentCreateTypeCmd = &cobra.Command{
	Use:   "create-type",
	Short: "Create a new content type",
	Long:  `Create a new content type in Strapi CMS.`,
	RunE:  runStrapiContentCreateType,
}

// strapiContentCreateCmd creates a new content entry
var strapiContentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a content entry",
	Long:  `Create a new entry in a specified content type.`,
	RunE:  runStrapiContentCreate,
}

// strapiContentListCmd lists content entries
var strapiContentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List content entries",
	Long:  `List all entries from a specified content type.`,
	RunE:  runStrapiContentList,
}

// strapiContentUpdateCmd updates a content entry
var strapiContentUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a content entry",
	Long:  `Update an existing entry in a specified content type.`,
	RunE:  runStrapiContentUpdate,
}

// strapiContentDeleteCmd deletes a content entry
var strapiContentDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a content entry",
	Long:  `Delete an entry from a specified content type.`,
	RunE:  runStrapiContentDelete,
}

func init() {
	strapiCmd.AddCommand(strapiContentCmd)

	// Add subcommands
	strapiContentCmd.AddCommand(strapiContentCreateTypeCmd)
	strapiContentCmd.AddCommand(strapiContentCreateCmd)
	strapiContentCmd.AddCommand(strapiContentListCmd)
	strapiContentCmd.AddCommand(strapiContentUpdateCmd)
	strapiContentCmd.AddCommand(strapiContentDeleteCmd)

	// Create type flags
	strapiContentCreateTypeCmd.Flags().StringVarP(&contentType, "name", "n", "", "content type name (required)")
	strapiContentCreateTypeCmd.Flags().StringVarP(&contentSchema, "schema", "s", "", "JSON schema for the content type (required)")
	strapiContentCreateTypeCmd.MarkFlagRequired("name")
	strapiContentCreateTypeCmd.MarkFlagRequired("schema")

	// Create flags
	strapiContentCreateCmd.Flags().StringVarP(&contentType, "type", "t", "", "content type (required)")
	strapiContentCreateCmd.Flags().StringVarP(&contentData, "data", "d", "", "JSON data for the entry (required)")
	strapiContentCreateCmd.Flags().BoolP("json", "j", false, "output as JSON")
	strapiContentCreateCmd.MarkFlagRequired("type")
	strapiContentCreateCmd.MarkFlagRequired("data")

	// List flags
	strapiContentListCmd.Flags().StringVarP(&contentType, "type", "t", "", "content type (required)")
	strapiContentListCmd.Flags().StringVarP(&contentFilter, "filter", "f", "", "filter query (e.g., status=published)")
	strapiContentListCmd.Flags().IntP("limit", "l", 25, "limit number of results")
	strapiContentListCmd.Flags().IntP("page", "p", 1, "page number")
	strapiContentListCmd.Flags().BoolP("json", "j", false, "output as JSON")
	strapiContentListCmd.MarkFlagRequired("type")

	// Update flags
	strapiContentUpdateCmd.Flags().StringVarP(&contentType, "type", "t", "", "content type (required)")
	strapiContentUpdateCmd.Flags().IntVarP(&contentID, "id", "i", 0, "content entry ID (required)")
	strapiContentUpdateCmd.Flags().StringVarP(&contentData, "data", "d", "", "JSON data for the update (required)")
	strapiContentUpdateCmd.Flags().BoolP("json", "j", false, "output as JSON")
	strapiContentUpdateCmd.MarkFlagRequired("type")
	strapiContentUpdateCmd.MarkFlagRequired("id")
	strapiContentUpdateCmd.MarkFlagRequired("data")

	// Delete flags
	strapiContentDeleteCmd.Flags().StringVarP(&contentType, "type", "t", "", "content type (required)")
	strapiContentDeleteCmd.Flags().IntVarP(&contentID, "id", "i", 0, "content entry ID (required)")
	strapiContentDeleteCmd.Flags().BoolP("yes", "y", false, "skip confirmation")
	strapiContentDeleteCmd.MarkFlagRequired("type")
	strapiContentDeleteCmd.MarkFlagRequired("id")
}

func runStrapiContentCreateType(cmd *cobra.Command, args []string) error {
	logger.InfoEvent().
		Str("name", contentType).
		Msg("Creating Strapi content type")

	// Parse schema
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(contentSchema), &schema); err != nil {
		return fmt.Errorf("invalid schema JSON: %w", err)
	}

	fmt.Printf("Creating content type: %s\n", contentType)
	fmt.Println("\nNote: Content type creation requires admin API access.")
	fmt.Println("This feature will create the content type schema in your Strapi instance.")

	// TODO: Implement content type creation via admin API
	// This requires admin authentication and the Content-Type Builder API
	fmt.Println("\n✓ Content type creation - Implementation in progress!")
	fmt.Println("For now, please create content types via the Strapi admin panel.")

	return nil
}

func runStrapiContentCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Parse data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(contentData), &data); err != nil {
		return fmt.Errorf("invalid data JSON: %w", err)
	}

	// Initialize Strapi client
	strapiClient, err := initStrapiClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Strapi client: %w", err)
	}

	logger.InfoEvent().
		Str("content_type", contentType).
		Msg("Creating content entry")

	// Create content
	entry, err := strapiClient.CreateContent(ctx, contentType, data)
	if err != nil {
		return fmt.Errorf("failed to create content: %w", err)
	}

	// Output result
	if jsonOutput {
		output, _ := json.MarshalIndent(entry, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println("\n✓ Content entry created successfully!")
		fmt.Printf("\nID: %d\n", entry.ID)
		fmt.Printf("Type: %s\n", contentType)
		fmt.Println("\nAttributes:")
		displayAttributes(entry.Attributes)
	}

	return nil
}

func runStrapiContentList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	jsonOutput, _ := cmd.Flags().GetBool("json")
	limit, _ := cmd.Flags().GetInt("limit")
	page, _ := cmd.Flags().GetInt("page")

	// Initialize Strapi client
	strapiClient, err := initStrapiClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Strapi client: %w", err)
	}

	// Parse filters
	filters := make(map[string]interface{})
	if contentFilter != "" {
		parts := strings.Split(contentFilter, "=")
		if len(parts) == 2 {
			filters[parts[0]] = parts[1]
		}
	}

	// Build list options
	opts := &strapi.ListOptions{
		Page:     page,
		PageSize: limit,
		Filters:  filters,
	}

	logger.InfoEvent().
		Str("content_type", contentType).
		Int("limit", limit).
		Int("page", page).
		Msg("Listing content entries")

	// List content
	entries, meta, err := strapiClient.ListContent(ctx, contentType, opts)
	if err != nil {
		return fmt.Errorf("failed to list content: %w", err)
	}

	// Output result
	if jsonOutput {
		output, _ := json.MarshalIndent(map[string]interface{}{
			"data": entries,
			"meta": meta,
		}, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("\nContent Type: %s\n", contentType)
		fmt.Printf("Total Entries: %d\n", meta.Pagination.Total)
		fmt.Printf("Page: %d of %d\n\n", meta.Pagination.Page, meta.Pagination.PageCount)

		if len(entries) == 0 {
			fmt.Println("No entries found.")
			return nil
		}

		displayContentEntriesTable(entries)
	}

	return nil
}

func runStrapiContentUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Parse data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(contentData), &data); err != nil {
		return fmt.Errorf("invalid data JSON: %w", err)
	}

	// Initialize Strapi client
	strapiClient, err := initStrapiClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Strapi client: %w", err)
	}

	logger.InfoEvent().
		Str("content_type", contentType).
		Int("id", contentID).
		Msg("Updating content entry")

	// Update content
	entry, err := strapiClient.UpdateContent(ctx, contentType, contentID, data)
	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}

	// Output result
	if jsonOutput {
		output, _ := json.MarshalIndent(entry, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Println("\n✓ Content entry updated successfully!")
		fmt.Printf("\nID: %d\n", entry.ID)
		fmt.Printf("Type: %s\n", contentType)
		fmt.Println("\nAttributes:")
		displayAttributes(entry.Attributes)
	}

	return nil
}

func runStrapiContentDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	// Confirm deletion unless --yes flag is provided
	if !skipConfirm {
		fmt.Printf("Are you sure you want to delete entry %d from %s? (y/N): ", contentID, contentType)
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Initialize Strapi client
	strapiClient, err := initStrapiClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Strapi client: %w", err)
	}

	logger.InfoEvent().
		Str("content_type", contentType).
		Int("id", contentID).
		Msg("Deleting content entry")

	// Delete content
	if err := strapiClient.DeleteContent(ctx, contentType, contentID); err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}

	fmt.Printf("\n✓ Content entry %d deleted from %s\n", contentID, contentType)

	return nil
}

// Helper functions

func initStrapiClient() (*strapi.Client, error) {
	// Get configuration
	baseURL := viper.GetString("strapi.base_url")
	if baseURL == "" {
		baseURL = viper.GetString("ainative.base_url")
		if baseURL == "" {
			return nil, fmt.Errorf("Strapi base URL not configured. Use 'ainative-code config set strapi.base_url <url>'")
		}
	}

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(baseURL),
	)

	// Create Strapi client
	strapiClient := strapi.New(
		strapi.WithAPIClient(apiClient),
		strapi.WithBaseURL(baseURL),
	)

	return strapiClient, nil
}

func displayAttributes(attrs map[string]interface{}) {
	for key, value := range attrs {
		fmt.Printf("  %s: ", key)
		switch v := value.(type) {
		case string:
			if len(v) > 100 {
				fmt.Printf("%s...\n", v[:97])
			} else {
				fmt.Printf("%s\n", v)
			}
		case map[string]interface{}, []interface{}:
			jsonVal, _ := json.MarshalIndent(v, "  ", "  ")
			fmt.Printf("%s\n", string(jsonVal))
		default:
			fmt.Printf("%v\n", v)
		}
	}
}

func displayContentEntriesTable(entries []*strapi.ContentEntry) {
	// Print header
	fmt.Printf("%-8s | %s\n", "ID", "Attributes")
	fmt.Println(strings.Repeat("-", 80))

	// Print rows
	for _, entry := range entries {
		// Get first few attributes for preview
		preview := getAttributesPreview(entry.Attributes)
		fmt.Printf("%-8d | %s\n", entry.ID, preview)
	}
}

func getAttributesPreview(attrs map[string]interface{}) string {
	if len(attrs) == 0 {
		return "-"
	}

	parts := make([]string, 0, 3)
	count := 0
	for key, value := range attrs {
		if count >= 3 {
			break
		}
		var valueStr string
		switch v := value.(type) {
		case string:
			if len(v) > 30 {
				valueStr = v[:27] + "..."
			} else {
				valueStr = v
			}
		default:
			valueStr = fmt.Sprintf("%v", v)
			if len(valueStr) > 30 {
				valueStr = valueStr[:27] + "..."
			}
		}
		parts = append(parts, fmt.Sprintf("%s: %s", key, valueStr))
		count++
	}

	return strings.Join(parts, ", ")
}

// Helper to convert string to int
func parseIntFlag(s string) (int, error) {
	return strconv.Atoi(s)
}
