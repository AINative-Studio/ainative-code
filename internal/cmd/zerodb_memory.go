package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/AINative-studio/ainative-code/internal/client/zerodb"
)

var (
	// Memory store flags
	storeAgentID   string
	storeContent   string
	storeRole      string
	storeSessionID string
	storeMetadata  string

	// Memory retrieve flags
	retrieveAgentID   string
	retrieveQuery     string
	retrieveLimit     int
	retrieveSessionID string

	// Memory clear flags
	clearAgentID   string
	clearSessionID string

	// Memory list flags
	listAgentID   string
	listSessionID string
	listLimit     int
	listOffset    int

	// Memory output flags
	memoryOutputJSON bool
)

// zerodbMemoryCmd represents the zerodb memory command
var zerodbMemoryCmd = &cobra.Command{
	Use:   "memory",
	Short: "Manage ZeroDB agent memory operations",
	Long: `Manage ZeroDB agent memory storage and retrieval.

Agent memory provides long-term conversation context and semantic search
capabilities for AI agents. Memories are stored with embeddings for
efficient similarity-based retrieval.

Examples:
  # Store agent memory
  ainative-code zerodb memory store --agent-id agent_123 --content "User prefers dark mode" --metadata '{"category":"preference"}'

  # Retrieve relevant memories
  ainative-code zerodb memory retrieve --agent-id agent_123 --query "user preferences" --limit 5

  # List all memories for an agent
  ainative-code zerodb memory list --agent-id agent_123

  # Clear agent memories
  ainative-code zerodb memory clear --agent-id agent_123`,
	Aliases: []string{"mem"},
}

// zerodbMemoryStoreCmd represents the zerodb memory store command
var zerodbMemoryStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Store agent memory",
	Long: `Store agent memory content with optional metadata.

The content is automatically embedded for semantic search capabilities.

Examples:
  # Store simple memory
  ainative-code zerodb memory store --agent-id agent_123 --content "User asked about pricing"

  # Store with role and session
  ainative-code zerodb memory store --agent-id agent_123 --content "User wants premium plan" --role user --session-id session_abc

  # Store with metadata
  ainative-code zerodb memory store --agent-id agent_123 --content "User prefers email notifications" --metadata '{"category":"preference","importance":"high"}'`,
	RunE: runMemoryStore,
}

// zerodbMemoryRetrieveCmd represents the zerodb memory retrieve command
var zerodbMemoryRetrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve agent memories using semantic search",
	Long: `Retrieve agent memories that are semantically similar to the query.

Uses vector similarity search to find the most relevant memories
based on the query content.

Examples:
  # Search for relevant memories
  ainative-code zerodb memory retrieve --agent-id agent_123 --query "user preferences"

  # Limit results
  ainative-code zerodb memory retrieve --agent-id agent_123 --query "pricing questions" --limit 3

  # Filter by session
  ainative-code zerodb memory retrieve --agent-id agent_123 --query "recent questions" --session-id session_abc

  # JSON output
  ainative-code zerodb memory retrieve --agent-id agent_123 --query "user info" --json`,
	RunE: runMemoryRetrieve,
}

// zerodbMemoryClearCmd represents the zerodb memory clear command
var zerodbMemoryClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear agent memories",
	Long: `Clear all memories for an agent or specific session.

This operation is permanent and cannot be undone.

Examples:
  # Clear all memories for an agent
  ainative-code zerodb memory clear --agent-id agent_123

  # Clear memories for a specific session
  ainative-code zerodb memory clear --agent-id agent_123 --session-id session_abc`,
	RunE: runMemoryClear,
}

// zerodbMemoryListCmd represents the zerodb memory list command
var zerodbMemoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agent memories",
	Long: `List all memories for an agent with pagination support.

Displays memories in chronological order (most recent first).

Examples:
  # List all memories for an agent
  ainative-code zerodb memory list --agent-id agent_123

  # List with pagination
  ainative-code zerodb memory list --agent-id agent_123 --limit 20 --offset 40

  # Filter by session
  ainative-code zerodb memory list --agent-id agent_123 --session-id session_abc

  # JSON output
  ainative-code zerodb memory list --agent-id agent_123 --json`,
	RunE: runMemoryList,
}

func init() {
	zerodbCmd.AddCommand(zerodbMemoryCmd)

	// Add memory subcommands
	zerodbMemoryCmd.AddCommand(zerodbMemoryStoreCmd)
	zerodbMemoryCmd.AddCommand(zerodbMemoryRetrieveCmd)
	zerodbMemoryCmd.AddCommand(zerodbMemoryClearCmd)
	zerodbMemoryCmd.AddCommand(zerodbMemoryListCmd)

	// Memory store flags
	zerodbMemoryStoreCmd.Flags().StringVar(&storeAgentID, "agent-id", "", "agent ID (required)")
	zerodbMemoryStoreCmd.Flags().StringVar(&storeContent, "content", "", "memory content (required)")
	zerodbMemoryStoreCmd.Flags().StringVar(&storeRole, "role", "", "message role (user, assistant, system)")
	zerodbMemoryStoreCmd.Flags().StringVar(&storeSessionID, "session-id", "", "session ID")
	zerodbMemoryStoreCmd.Flags().StringVar(&storeMetadata, "metadata", "", "additional metadata as JSON")
	zerodbMemoryStoreCmd.MarkFlagRequired("agent-id")
	zerodbMemoryStoreCmd.MarkFlagRequired("content")

	// Memory retrieve flags
	zerodbMemoryRetrieveCmd.Flags().StringVar(&retrieveAgentID, "agent-id", "", "agent ID (required)")
	zerodbMemoryRetrieveCmd.Flags().StringVar(&retrieveQuery, "query", "", "search query (required)")
	zerodbMemoryRetrieveCmd.Flags().IntVar(&retrieveLimit, "limit", 10, "maximum number of memories to retrieve")
	zerodbMemoryRetrieveCmd.Flags().StringVar(&retrieveSessionID, "session-id", "", "filter by session ID")
	zerodbMemoryRetrieveCmd.MarkFlagRequired("agent-id")
	zerodbMemoryRetrieveCmd.MarkFlagRequired("query")

	// Memory clear flags
	zerodbMemoryClearCmd.Flags().StringVar(&clearAgentID, "agent-id", "", "agent ID (required)")
	zerodbMemoryClearCmd.Flags().StringVar(&clearSessionID, "session-id", "", "session ID to clear (optional)")
	zerodbMemoryClearCmd.MarkFlagRequired("agent-id")

	// Memory list flags
	zerodbMemoryListCmd.Flags().StringVar(&listAgentID, "agent-id", "", "agent ID (required)")
	zerodbMemoryListCmd.Flags().StringVar(&listSessionID, "session-id", "", "filter by session ID")
	zerodbMemoryListCmd.Flags().IntVar(&listLimit, "limit", 100, "maximum number of memories to return")
	zerodbMemoryListCmd.Flags().IntVar(&listOffset, "offset", 0, "number of memories to skip")
	zerodbMemoryListCmd.MarkFlagRequired("agent-id")

	// Global output flag for all memory commands
	zerodbMemoryCmd.PersistentFlags().BoolVar(&memoryOutputJSON, "json", false, "output in JSON format")
}

func runMemoryStore(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse metadata JSON if provided
	var metadata map[string]interface{}
	if storeMetadata != "" {
		if err := json.Unmarshal([]byte(storeMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
	}

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Store memory
	req := &zerodb.MemoryStoreRequest{
		AgentID:   storeAgentID,
		Content:   storeContent,
		Role:      storeRole,
		SessionID: storeSessionID,
		Metadata:  metadata,
	}

	memory, err := zdbClient.StoreMemory(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}

	// Output result
	if memoryOutputJSON {
		return zerodbOutputJSON(memory)
	}

	fmt.Printf("Memory stored successfully!\n")
	fmt.Printf("  ID:         %s\n", memory.ID)
	fmt.Printf("  Agent ID:   %s\n", memory.AgentID)
	if memory.SessionID != "" {
		fmt.Printf("  Session ID: %s\n", memory.SessionID)
	}
	if memory.Role != "" {
		fmt.Printf("  Role:       %s\n", memory.Role)
	}
	fmt.Printf("  Created At: %s\n", memory.CreatedAt.Format(time.RFC3339))
	fmt.Printf("  Content:    %s\n", truncateString(memory.Content, 100))

	return nil
}

func runMemoryRetrieve(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Retrieve memories
	req := &zerodb.MemoryRetrieveRequest{
		AgentID:   retrieveAgentID,
		Query:     retrieveQuery,
		Limit:     retrieveLimit,
		SessionID: retrieveSessionID,
	}

	memories, err := zdbClient.RetrieveMemory(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to retrieve memories: %w", err)
	}

	// Output result
	if memoryOutputJSON {
		return zerodbOutputJSON(memories)
	}

	if len(memories) == 0 {
		fmt.Println("No relevant memories found.")
		return nil
	}

	fmt.Printf("Found %d relevant memor%s:\n\n", len(memories), pluralize(len(memories), "y", "ies"))
	for i, memory := range memories {
		fmt.Printf("Memory %d (similarity: %.2f):\n", i+1, memory.Similarity)
		fmt.Printf("  ID:         %s\n", memory.ID)
		if memory.SessionID != "" {
			fmt.Printf("  Session ID: %s\n", memory.SessionID)
		}
		if memory.Role != "" {
			fmt.Printf("  Role:       %s\n", memory.Role)
		}
		fmt.Printf("  Created At: %s\n", memory.CreatedAt.Format(time.RFC3339))
		fmt.Printf("  Content:    %s\n", memory.Content)
		if len(memory.Metadata) > 0 {
			metaJSON, _ := json.MarshalIndent(memory.Metadata, "  ", "  ")
			fmt.Printf("  Metadata:   %s\n", string(metaJSON))
		}
		fmt.Println()
	}

	return nil
}

func runMemoryClear(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// Clear memories
	req := &zerodb.MemoryClearRequest{
		AgentID:   clearAgentID,
		SessionID: clearSessionID,
	}

	resp, err := zdbClient.ClearMemory(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to clear memories: %w", err)
	}

	// Output result
	if memoryOutputJSON {
		return zerodbOutputJSON(resp)
	}

	if clearSessionID != "" {
		fmt.Printf("Cleared %d memor%s for agent %s (session: %s)\n",
			resp.Deleted,
			pluralize(resp.Deleted, "y", "ies"),
			clearAgentID,
			clearSessionID)
	} else {
		fmt.Printf("Cleared %d memor%s for agent %s\n",
			resp.Deleted,
			pluralize(resp.Deleted, "y", "ies"),
			clearAgentID)
	}

	if resp.Message != "" {
		fmt.Printf("  %s\n", resp.Message)
	}

	return nil
}

func runMemoryList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Create ZeroDB client
	zdbClient, err := createZeroDBClient()
	if err != nil {
		return err
	}

	// List memories
	req := &zerodb.MemoryListRequest{
		AgentID:   listAgentID,
		SessionID: listSessionID,
		Limit:     listLimit,
		Offset:    listOffset,
	}

	memories, total, err := zdbClient.ListMemory(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	// Output result
	if memoryOutputJSON {
		result := map[string]interface{}{
			"memories": memories,
			"total":    total,
			"limit":    listLimit,
			"offset":   listOffset,
		}
		return zerodbOutputJSON(result)
	}

	if len(memories) == 0 {
		fmt.Println("No memories found.")
		return nil
	}

	// Create table writer for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tROLE\tCREATED\tCONTENT")
	fmt.Fprintln(w, "--\t----\t-------\t-------")

	for _, memory := range memories {
		role := memory.Role
		if role == "" {
			role = "-"
		}
		content := truncateString(memory.Content, 50)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			memory.ID,
			role,
			memory.CreatedAt.Format("2006-01-02 15:04"),
			content,
		)
	}

	w.Flush()

	fmt.Printf("\nShowing %d-%d of %d total memor%s\n",
		listOffset+1,
		listOffset+len(memories),
		total,
		pluralize(total, "y", "ies"))

	return nil
}

// Helper function for pluralization
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
