package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	// Table create flags
	tableName   string
	tableSchema string

	// Table insert flags
	insertTable string
	insertData  string

	// Table query flags
	queryTable  string
	queryFilter string
	queryLimit  int
	queryOffset int
	querySort   string

	// Table update flags
	updateTable string
	updateID    string
	updateData  string

	// Table delete flags
	deleteTable string
	deleteID    string

	// Table output flags
	tableOutputJSON bool
)

// zerodbTableCmd represents the zerodb table command
var zerodbTableCmd = &cobra.Command{
	Use:   "table",
	Short: "Manage ZeroDB NoSQL tables",
	Long: `Manage ZeroDB NoSQL tables including creation, querying, and data manipulation.

ZeroDB NoSQL provides a flexible document database with MongoDB-style query capabilities.

Examples:
  # Create a table
  ainative-code zerodb table create --name users --schema '{"type":"object","properties":{"name":{"type":"string"},"email":{"type":"string"},"age":{"type":"number"}}}'

  # Insert a document
  ainative-code zerodb table insert --table users --data '{"name":"John Doe","email":"john@example.com","age":30}'

  # Query documents
  ainative-code zerodb table query --table users --filter '{"age":{"$gte":18}}'

  # Update a document
  ainative-code zerodb table update --table users --id abc123 --data '{"age":31}'

  # Delete a document
  ainative-code zerodb table delete --table users --id abc123

  # List all tables
  ainative-code zerodb table list`,
	Aliases: []string{"tables", "tbl"},
}

// zerodbTableCreateCmd represents the zerodb table create command
var zerodbTableCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new NoSQL table",
	Long: `Create a new NoSQL table with a specified schema.

The schema should be a JSON object defining the structure of documents in the table.
The schema follows JSON Schema specification.

Examples:
  # Create a simple table
  ainative-code zerodb table create --name users --schema '{"type":"object","properties":{"name":{"type":"string"},"age":{"type":"number"}}}'

  # Create a table with required fields
  ainative-code zerodb table create --name products --schema '{"type":"object","properties":{"name":{"type":"string"},"price":{"type":"number"}},"required":["name","price"]}'`,
	RunE: runTableCreate,
}

// zerodbTableInsertCmd represents the zerodb table insert command
var zerodbTableInsertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert a document into a table",
	Long: `Insert a new document into the specified NoSQL table.

The data should be a JSON object matching the table's schema.

Examples:
  # Insert a user document
  ainative-code zerodb table insert --table users --data '{"name":"Jane Smith","email":"jane@example.com","age":25}'

  # Insert with nested data
  ainative-code zerodb table insert --table products --data '{"name":"Widget","price":19.99,"metadata":{"category":"tools","brand":"ACME"}}'`,
	RunE: runTableInsert,
}

// zerodbTableQueryCmd represents the zerodb table query command
var zerodbTableQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query documents from a table",
	Long: `Query documents from a NoSQL table using MongoDB-style filters.

Supports comparison operators ($gt, $gte, $lt, $lte, $ne, $eq),
logical operators ($and, $or, $not), array operators ($in, $nin),
and existence checks ($exists).

Examples:
  # Query all documents
  ainative-code zerodb table query --table users

  # Query with simple filter
  ainative-code zerodb table query --table users --filter '{"age":30}'

  # Query with comparison operators
  ainative-code zerodb table query --table users --filter '{"age":{"$gte":18,"$lt":65}}'

  # Query with logical operators
  ainative-code zerodb table query --table users --filter '{"$and":[{"age":{"$gte":18}},{"status":"active"}]}'

  # Query with limit and offset
  ainative-code zerodb table query --table users --limit 10 --offset 20

  # Query with sorting
  ainative-code zerodb table query --table users --sort 'age:desc,name:asc'`,
	RunE: runTableQuery,
}

// zerodbTableUpdateCmd represents the zerodb table update command
var zerodbTableUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a document in a table",
	Long: `Update an existing document in the specified NoSQL table.

The data should be a JSON object with the fields to update.
Only specified fields will be updated; other fields remain unchanged.

Examples:
  # Update a single field
  ainative-code zerodb table update --table users --id abc123 --data '{"age":31}'

  # Update multiple fields
  ainative-code zerodb table update --table users --id abc123 --data '{"age":31,"email":"newemail@example.com"}'`,
	RunE: runTableUpdate,
}

// zerodbTableDeleteCmd represents the zerodb table delete command
var zerodbTableDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a document from a table",
	Long: `Delete a document from the specified NoSQL table by its ID.

This operation is permanent and cannot be undone.

Examples:
  # Delete a document
  ainative-code zerodb table delete --table users --id abc123`,
	RunE: runTableDelete,
}

// zerodbTableListCmd represents the zerodb table list command
var zerodbTableListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all NoSQL tables",
	Long: `List all NoSQL tables in the current project.

Displays table name, ID, and creation date.

Examples:
  # List all tables
  ainative-code zerodb table list

  # List with JSON output
  ainative-code zerodb table list --json`,
	RunE: runTableList,
}

func init() {
	zerodbCmd.AddCommand(zerodbTableCmd)

	// Add table subcommands
	zerodbTableCmd.AddCommand(zerodbTableCreateCmd)
	zerodbTableCmd.AddCommand(zerodbTableInsertCmd)
	zerodbTableCmd.AddCommand(zerodbTableQueryCmd)
	zerodbTableCmd.AddCommand(zerodbTableUpdateCmd)
	zerodbTableCmd.AddCommand(zerodbTableDeleteCmd)
	zerodbTableCmd.AddCommand(zerodbTableListCmd)

	// Table create flags
	zerodbTableCreateCmd.Flags().StringVar(&tableName, "name", "", "table name (required)")
	zerodbTableCreateCmd.Flags().StringVar(&tableSchema, "schema", "", "table schema as JSON (required)")
	zerodbTableCreateCmd.MarkFlagRequired("name")
	zerodbTableCreateCmd.MarkFlagRequired("schema")

	// Table insert flags
	zerodbTableInsertCmd.Flags().StringVar(&insertTable, "table", "", "table name (required)")
	zerodbTableInsertCmd.Flags().StringVar(&insertData, "data", "", "document data as JSON (required)")
	zerodbTableInsertCmd.MarkFlagRequired("table")
	zerodbTableInsertCmd.MarkFlagRequired("data")

	// Table query flags
	zerodbTableQueryCmd.Flags().StringVar(&queryTable, "table", "", "table name (required)")
	zerodbTableQueryCmd.Flags().StringVar(&queryFilter, "filter", "", "query filter as JSON")
	zerodbTableQueryCmd.Flags().IntVar(&queryLimit, "limit", 100, "maximum number of documents to return")
	zerodbTableQueryCmd.Flags().IntVar(&queryOffset, "offset", 0, "number of documents to skip")
	zerodbTableQueryCmd.Flags().StringVar(&querySort, "sort", "", "sort fields (e.g., 'age:desc,name:asc')")
	zerodbTableQueryCmd.MarkFlagRequired("table")

	// Table update flags
	zerodbTableUpdateCmd.Flags().StringVar(&updateTable, "table", "", "table name (required)")
	zerodbTableUpdateCmd.Flags().StringVar(&updateID, "id", "", "document ID (required)")
	zerodbTableUpdateCmd.Flags().StringVar(&updateData, "data", "", "update data as JSON (required)")
	zerodbTableUpdateCmd.MarkFlagRequired("table")
	zerodbTableUpdateCmd.MarkFlagRequired("id")
	zerodbTableUpdateCmd.MarkFlagRequired("data")

	// Table delete flags
	zerodbTableDeleteCmd.Flags().StringVar(&deleteTable, "table", "", "table name (required)")
	zerodbTableDeleteCmd.Flags().StringVar(&deleteID, "id", "", "document ID (required)")
	zerodbTableDeleteCmd.MarkFlagRequired("table")
	zerodbTableDeleteCmd.MarkFlagRequired("id")

	// Table output flags - register --json flag for all table commands
	zerodbTableCmd.PersistentFlags().BoolVar(&tableOutputJSON, "json", false, "output in JSON format")
}

func runTableCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse schema JSON
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(tableSchema), &schema); err != nil {
		return fmt.Errorf("invalid schema JSON: %w", err)
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Create table
	table, err := zdbClient.CreateTable(ctx, tableName, schema)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Output result
	if tableOutputJSON {
		return zerodbOutputJSON(table)
	}

	fmt.Printf("Table created successfully!\n")
	fmt.Printf("  ID:         %s\n", table.ID)
	fmt.Printf("  Name:       %s\n", table.Name)
	fmt.Printf("  Created At: %s\n", table.CreatedAt.Format(time.RFC3339))

	return nil
}

func runTableInsert(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse data JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(insertData), &data); err != nil {
		return fmt.Errorf("invalid data JSON: %w", err)
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Insert document
	id, doc, err := zdbClient.Insert(ctx, insertTable, data)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	// Output result
	if tableOutputJSON {
		return zerodbOutputJSON(doc)
	}

	fmt.Printf("Document inserted successfully!\n")
	fmt.Printf("  ID:         %s\n", id)
	fmt.Printf("  Table:      %s\n", insertTable)
	fmt.Printf("  Created At: %s\n", doc.CreatedAt.Format(time.RFC3339))

	return nil
}

func runTableQuery(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse filter JSON
	var filter zerodb.QueryFilter
	if queryFilter != "" {
		if err := json.Unmarshal([]byte(queryFilter), &filter); err != nil {
			return fmt.Errorf("invalid filter JSON: %w", err)
		}
	}

	// Parse sort options
	sortMap := make(map[string]int)
	if querySort != "" {
		for _, sortField := range strings.Split(querySort, ",") {
			parts := strings.Split(strings.TrimSpace(sortField), ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid sort format: %s (expected field:asc or field:desc)", sortField)
			}
			field := strings.TrimSpace(parts[0])
			order := strings.TrimSpace(parts[1])
			if order == "asc" {
				sortMap[field] = 1
			} else if order == "desc" {
				sortMap[field] = -1
			} else {
				return fmt.Errorf("invalid sort order: %s (expected asc or desc)", order)
			}
		}
	}

	options := zerodb.QueryOptions{
		Limit:  queryLimit,
		Offset: queryOffset,
		Sort:   sortMap,
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Query documents
	docs, err := zdbClient.Query(ctx, queryTable, filter, options)
	if err != nil {
		return fmt.Errorf("failed to query documents: %w", err)
	}

	// Output result
	if tableOutputJSON {
		return zerodbOutputJSON(docs)
	}

	if len(docs) == 0 {
		fmt.Println("No documents found.")
		return nil
	}

	fmt.Printf("Found %d document(s):\n\n", len(docs))
	for i, doc := range docs {
		fmt.Printf("Document %d:\n", i+1)
		fmt.Printf("  ID:         %s\n", doc.ID)
		fmt.Printf("  Created At: %s\n", doc.CreatedAt.Format(time.RFC3339))
		fmt.Printf("  Updated At: %s\n", doc.UpdatedAt.Format(time.RFC3339))
		fmt.Printf("  Data:\n")

		// Pretty print data
		dataJSON, _ := json.MarshalIndent(doc.Data, "    ", "  ")
		fmt.Printf("    %s\n", string(dataJSON))
		fmt.Println()
	}

	return nil
}

func runTableUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse data JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(updateData), &data); err != nil {
		return fmt.Errorf("invalid data JSON: %w", err)
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Update document
	doc, err := zdbClient.Update(ctx, updateTable, updateID, data)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	// Output result
	if tableOutputJSON {
		return zerodbOutputJSON(doc)
	}

	fmt.Printf("Document updated successfully!\n")
	fmt.Printf("  ID:         %s\n", doc.ID)
	fmt.Printf("  Table:      %s\n", updateTable)
	fmt.Printf("  Updated At: %s\n", doc.UpdatedAt.Format(time.RFC3339))

	return nil
}

func runTableDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Delete document
	if err := zdbClient.Delete(ctx, deleteTable, deleteID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	// Output result
	if tableOutputJSON {
		return zerodbOutputJSON(map[string]interface{}{
			"success": true,
			"id":      deleteID,
			"table":   deleteTable,
		})
	}

	fmt.Printf("Document deleted successfully!\n")
	fmt.Printf("  ID:    %s\n", deleteID)
	fmt.Printf("  Table: %s\n", deleteTable)

	return nil
}

func runTableList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// List tables
	tables, err := zdbClient.ListTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	// Output result
	if tableOutputJSON {
		return zerodbOutputJSON(tables)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found.")
		return nil
	}

	// Create table writer for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tID\tCREATED AT")
	fmt.Fprintln(w, "----\t--\t----------")

	for _, table := range tables {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			table.Name,
			table.ID,
			table.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	w.Flush()

	fmt.Printf("\nTotal: %d table(s)\n", len(tables))

	return nil
}

// createZeroDBClient creates a ZeroDB client with configuration from viper.
func createZeroDBClient() (*zerodb.Client, error) {
	// Get configuration
	baseURL := viper.GetString("services.zerodb.base_url")
	if baseURL == "" {
		baseURL = "https://api.ainative.studio"
	}

	projectID := viper.GetString("services.zerodb.project_id")
	if projectID == "" {
		return nil, fmt.Errorf("services.zerodb.project_id not configured (set in config file or AINATIVE_CODE_SERVICES_ZERODB_PROJECT_ID env var)")
	}

	apiKey := viper.GetString("services.zerodb.api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("services.zerodb.api_key not configured (set in config file or AINATIVE_CODE_SERVICES_ZERODB_API_KEY env var)")
	}

	// Create HTTP client with API key header injector
	httpClient := client.New(
		client.WithBaseURL(baseURL),
		client.WithTimeout(30*time.Second),
		client.WithHTTPClient(&http.Client{
			Timeout: 30 * time.Second,
			Transport: &apiKeyTransport{
				apiKey:    apiKey,
				transport: http.DefaultTransport,
			},
		}),
	)

	// Create ZeroDB client
	zdbClient := zerodb.New(
		zerodb.WithAPIClient(httpClient),
		zerodb.WithProjectID(projectID),
	)

	logger.DebugEvent().
		Str("base_url", baseURL).
		Str("project_id", projectID).
		Msg("Created ZeroDB client")

	return zdbClient, nil
}

// apiKeyTransport is an http.RoundTripper that adds the X-API-Key header to all requests.
type apiKeyTransport struct {
	apiKey    string
	transport http.RoundTripper
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	reqCopy := req.Clone(req.Context())
	reqCopy.Header.Set("X-API-Key", t.apiKey)
	return t.transport.RoundTrip(reqCopy)
}
