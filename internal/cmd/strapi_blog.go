package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/AINative-studio/ainative-code/internal/client"
	"github.com/AINative-studio/ainative-code/internal/client/strapi"
	"github.com/AINative-studio/ainative-code/internal/logger"
)

var (
	// Blog create flags
	blogTitle   string
	blogContent string
	blogAuthor  string
	blogSlug    string
	blogTags    []string
	blogStatus  string

	// Blog list flags
	blogListStatus string
	blogListAuthor string
	blogListLimit  int
	blogListPage   int

	// Blog update flags
	blogID int

	// Blog output format flags
	blogOutputJSON bool
)

// strapiBlogCmd represents the strapi blog command
var strapiBlogCmd = &cobra.Command{
	Use:   "blog",
	Short: "Manage Strapi blog posts",
	Long: `Manage Strapi blog posts including creating, listing, updating, publishing, and deleting posts.

Blog posts support markdown content and can be managed through simple CLI commands.

Examples:
  # Create a draft blog post
  ainative-code strapi blog create --title "My Post" --content "# Hello World" --author "John"

  # Create from markdown file
  ainative-code strapi blog create --title "My Post" --content @post.md --author "John"

  # List all published posts
  ainative-code strapi blog list --status published

  # List posts by author
  ainative-code strapi blog list --author "John Doe" --limit 10

  # Update a blog post
  ainative-code strapi blog update --id 42 --title "Updated Title" --content "New content"

  # Publish a blog post
  ainative-code strapi blog publish --id 42

  # Delete a blog post
  ainative-code strapi blog delete --id 42

  # Get JSON output
  ainative-code strapi blog list --json`,
}

// strapiBlogCreateCmd creates a new blog post
var strapiBlogCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new blog post",
	Long: `Create a new blog post in Strapi CMS.

Content can be provided directly via the --content flag or from a file using @filename syntax.
Markdown formatting is supported.

Examples:
  # Create with inline content
  ainative-code strapi blog create --title "Hello World" --content "# My First Post" --author "John"

  # Create from markdown file
  ainative-code strapi blog create --title "My Post" --content @post.md --author "John"

  # Create with tags and custom slug
  ainative-code strapi blog create --title "Tech Article" --content @article.md --author "Jane" --tags go,rust --slug tech-article

  # Create and immediately publish
  ainative-code strapi blog create --title "News" --content @news.md --author "Editor" --status published`,
	RunE: runStrapiBlogCreate,
}

// strapiBlogListCmd lists blog posts
var strapiBlogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List blog posts",
	Long: `List blog posts from Strapi CMS with optional filtering.

Supports filtering by status and author, with pagination controls.

Examples:
  # List all posts
  ainative-code strapi blog list

  # List published posts only
  ainative-code strapi blog list --status published

  # List posts by author
  ainative-code strapi blog list --author "John Doe"

  # List with pagination
  ainative-code strapi blog list --limit 25 --page 2

  # List as JSON
  ainative-code strapi blog list --json`,
	Aliases: []string{"ls"},
	RunE:    runStrapiBlogList,
}

// strapiBlogUpdateCmd updates an existing blog post
var strapiBlogUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing blog post",
	Long: `Update an existing blog post in Strapi CMS.

You can update the title, content, author, or other fields. Only provided fields will be updated.

Examples:
  # Update title only
  ainative-code strapi blog update --id 42 --title "New Title"

  # Update content from file
  ainative-code strapi blog update --id 42 --content @updated.md

  # Update multiple fields
  ainative-code strapi blog update --id 42 --title "Updated" --author "Jane Doe" --tags go,rust`,
	RunE: runStrapiBlogUpdate,
}

// strapiBlogPublishCmd publishes a blog post
var strapiBlogPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a blog post",
	Long: `Publish a blog post by changing its status to published.

This will make the post visible to the public and set the publishedAt timestamp.

Examples:
  # Publish a post
  ainative-code strapi blog publish --id 42

  # Publish with JSON output
  ainative-code strapi blog publish --id 42 --json`,
	RunE: runStrapiBlogPublish,
}

// strapiBlogDeleteCmd deletes a blog post
var strapiBlogDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a blog post",
	Long: `Delete a blog post from Strapi CMS.

WARNING: This action cannot be undone.

Examples:
  # Delete a post
  ainative-code strapi blog delete --id 42`,
	Aliases: []string{"rm"},
	RunE:    runStrapiBlogDelete,
}

func init() {
	strapiCmd.AddCommand(strapiBlogCmd)

	// Add subcommands
	strapiBlogCmd.AddCommand(strapiBlogCreateCmd)
	strapiBlogCmd.AddCommand(strapiBlogListCmd)
	strapiBlogCmd.AddCommand(strapiBlogUpdateCmd)
	strapiBlogCmd.AddCommand(strapiBlogPublishCmd)
	strapiBlogCmd.AddCommand(strapiBlogDeleteCmd)

	// Create command flags
	strapiBlogCreateCmd.Flags().StringVar(&blogTitle, "title", "", "blog post title (required)")
	strapiBlogCreateCmd.Flags().StringVar(&blogContent, "content", "", "blog post content or @filename (required)")
	strapiBlogCreateCmd.Flags().StringVar(&blogAuthor, "author", "", "blog post author")
	strapiBlogCreateCmd.Flags().StringVar(&blogSlug, "slug", "", "blog post slug (URL-friendly identifier)")
	strapiBlogCreateCmd.Flags().StringSliceVar(&blogTags, "tags", []string{}, "blog post tags (comma-separated)")
	strapiBlogCreateCmd.Flags().StringVar(&blogStatus, "status", "draft", "blog post status (draft or published)")
	strapiBlogCreateCmd.Flags().BoolVar(&blogOutputJSON, "json", false, "output as JSON")
	strapiBlogCreateCmd.MarkFlagRequired("title")
	strapiBlogCreateCmd.MarkFlagRequired("content")

	// List command flags
	strapiBlogListCmd.Flags().StringVar(&blogListStatus, "status", "", "filter by status (draft or published)")
	strapiBlogListCmd.Flags().StringVar(&blogListAuthor, "author", "", "filter by author")
	strapiBlogListCmd.Flags().IntVar(&blogListLimit, "limit", 25, "maximum number of posts to return")
	strapiBlogListCmd.Flags().IntVar(&blogListPage, "page", 1, "page number for pagination")
	strapiBlogListCmd.Flags().BoolVar(&blogOutputJSON, "json", false, "output as JSON")

	// Update command flags
	strapiBlogUpdateCmd.Flags().IntVar(&blogID, "id", 0, "blog post ID (required)")
	strapiBlogUpdateCmd.Flags().StringVar(&blogTitle, "title", "", "new blog post title")
	strapiBlogUpdateCmd.Flags().StringVar(&blogContent, "content", "", "new blog post content or @filename")
	strapiBlogUpdateCmd.Flags().StringVar(&blogAuthor, "author", "", "new blog post author")
	strapiBlogUpdateCmd.Flags().StringVar(&blogSlug, "slug", "", "new blog post slug")
	strapiBlogUpdateCmd.Flags().StringSliceVar(&blogTags, "tags", []string{}, "new blog post tags (comma-separated)")
	strapiBlogUpdateCmd.Flags().BoolVar(&blogOutputJSON, "json", false, "output as JSON")
	strapiBlogUpdateCmd.MarkFlagRequired("id")

	// Publish command flags
	strapiBlogPublishCmd.Flags().IntVar(&blogID, "id", 0, "blog post ID (required)")
	strapiBlogPublishCmd.Flags().BoolVar(&blogOutputJSON, "json", false, "output as JSON")
	strapiBlogPublishCmd.MarkFlagRequired("id")

	// Delete command flags
	strapiBlogDeleteCmd.Flags().IntVar(&blogID, "id", 0, "blog post ID (required)")
	strapiBlogDeleteCmd.MarkFlagRequired("id")
}

func runStrapiBlogCreate(cmd *cobra.Command, args []string) error {
	// Suppress INFO/DEBUG logs if JSON output is requested
	if blogOutputJSON {
		defer logger.SuppressInfoLogsForJSON()()
	}

	ctx := context.Background()

	// Get Strapi configuration
	strapiURL := viper.GetString("strapi.url")
	if strapiURL == "" {
		return fmt.Errorf("Strapi URL not configured. Use 'ainative-code strapi config --url <url>' to set it up")
	}

	// Process content (handle @filename syntax)
	content, err := processContentInput(blogContent)
	if err != nil {
		return fmt.Errorf("failed to process content: %w", err)
	}

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(strapiURL),
		client.WithTimeout(30*time.Second),
	)

	strapiClient := strapi.New(
		strapi.WithAPIClient(apiClient),
		strapi.WithBaseURL(strapiURL),
	)

	// Build request
	req := &strapi.CreateBlogPostRequest{
		Data: &strapi.BlogPostData{
			Title:   blogTitle,
			Content: content,
			Author:  blogAuthor,
			Slug:    blogSlug,
			Tags:    blogTags,
			Status:  blogStatus,
		},
	}

	logger.InfoEvent().
		Str("title", blogTitle).
		Str("author", blogAuthor).
		Str("status", blogStatus).
		Msg("Creating blog post")

	// Create blog post
	post, err := strapiClient.CreateBlogPost(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create blog post: %w", err)
	}

	// Output result
	if blogOutputJSON {
		return outputAsJSON(post)
	}

	fmt.Println("Blog post created successfully!")
	fmt.Printf("\nID: %d\n", post.ID)
	fmt.Printf("Title: %s\n", post.Attributes.Title)
	if post.Attributes.Author != "" {
		fmt.Printf("Author: %s\n", post.Attributes.Author)
	}
	fmt.Printf("Status: %s\n", post.Attributes.Status)
	if post.Attributes.Slug != "" {
		fmt.Printf("Slug: %s\n", post.Attributes.Slug)
	}
	fmt.Printf("Created: %s\n", post.Attributes.CreatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func runStrapiBlogList(cmd *cobra.Command, args []string) error {
	// Suppress INFO/DEBUG logs if JSON output is requested
	if blogOutputJSON {
		defer logger.SuppressInfoLogsForJSON()()
	}

	ctx := context.Background()

	// Get Strapi configuration
	strapiURL := viper.GetString("strapi.url")
	if strapiURL == "" {
		return fmt.Errorf("Strapi URL not configured. Use 'ainative-code strapi config --url <url>' to set it up")
	}

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(strapiURL),
		client.WithTimeout(30*time.Second),
	)

	strapiClient := strapi.New(
		strapi.WithAPIClient(apiClient),
		strapi.WithBaseURL(strapiURL),
	)

	// Build list options
	opts := &strapi.ListOptions{
		Page:     blogListPage,
		PageSize: blogListLimit,
		Sort:     []string{"createdAt:desc"},
		Filters:  make(map[string]interface{}),
	}

	if blogListStatus != "" {
		opts.Filters["status"] = blogListStatus
	}
	if blogListAuthor != "" {
		opts.Filters["author"] = blogListAuthor
	}

	logger.DebugEvent().
		Str("status", blogListStatus).
		Str("author", blogListAuthor).
		Int("limit", blogListLimit).
		Int("page", blogListPage).
		Msg("Listing blog posts")

	// List blog posts
	posts, meta, err := strapiClient.ListBlogPosts(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list blog posts: %w", err)
	}

	// Output result
	if blogOutputJSON {
		result := map[string]interface{}{
			"data": posts,
			"meta": meta,
		}
		return outputAsJSON(result)
	}

	if len(posts) == 0 {
		fmt.Println("No blog posts found.")
		return nil
	}

	// Display as table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "ID\tTITLE\tAUTHOR\tSTATUS\tCREATED\n")
	fmt.Fprintf(w, "--\t-----\t------\t------\t-------\n")

	for _, post := range posts {
		createdAt := post.Attributes.CreatedAt.Format("2006-01-02")
		author := post.Attributes.Author
		if author == "" {
			author = "-"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
			post.ID,
			truncateString(post.Attributes.Title, 40),
			truncateString(author, 20),
			post.Attributes.Status,
			createdAt,
		)
	}
	w.Flush()

	// Display pagination info
	if meta != nil && meta.Pagination != nil {
		fmt.Printf("\nPage %d of %d (Total: %d posts)\n",
			meta.Pagination.Page,
			meta.Pagination.PageCount,
			meta.Pagination.Total,
		)
	}

	return nil
}

func runStrapiBlogUpdate(cmd *cobra.Command, args []string) error {
	// Suppress INFO/DEBUG logs if JSON output is requested
	if blogOutputJSON {
		defer logger.SuppressInfoLogsForJSON()()
	}

	ctx := context.Background()

	// Get Strapi configuration
	strapiURL := viper.GetString("strapi.url")
	if strapiURL == "" {
		return fmt.Errorf("Strapi URL not configured. Use 'ainative-code strapi config --url <url>' to set it up")
	}

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(strapiURL),
		client.WithTimeout(30*time.Second),
	)

	strapiClient := strapi.New(
		strapi.WithAPIClient(apiClient),
		strapi.WithBaseURL(strapiURL),
	)

	// Build update request with only provided fields
	data := &strapi.BlogPostData{}
	hasChanges := false

	if blogTitle != "" {
		data.Title = blogTitle
		hasChanges = true
	}

	if blogContent != "" {
		content, err := processContentInput(blogContent)
		if err != nil {
			return fmt.Errorf("failed to process content: %w", err)
		}
		data.Content = content
		hasChanges = true
	}

	if blogAuthor != "" {
		data.Author = blogAuthor
		hasChanges = true
	}

	if blogSlug != "" {
		data.Slug = blogSlug
		hasChanges = true
	}

	if len(blogTags) > 0 {
		data.Tags = blogTags
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no fields to update. Please provide at least one field (--title, --content, --author, --slug, or --tags)")
	}

	req := &strapi.UpdateBlogPostRequest{
		Data: data,
	}

	logger.InfoEvent().
		Int("id", blogID).
		Msg("Updating blog post")

	// Update blog post
	post, err := strapiClient.UpdateBlogPost(ctx, blogID, req)
	if err != nil {
		return fmt.Errorf("failed to update blog post: %w", err)
	}

	// Output result
	if blogOutputJSON {
		return outputAsJSON(post)
	}

	fmt.Println("Blog post updated successfully!")
	fmt.Printf("\nID: %d\n", post.ID)
	fmt.Printf("Title: %s\n", post.Attributes.Title)
	if post.Attributes.Author != "" {
		fmt.Printf("Author: %s\n", post.Attributes.Author)
	}
	fmt.Printf("Status: %s\n", post.Attributes.Status)
	if post.Attributes.Slug != "" {
		fmt.Printf("Slug: %s\n", post.Attributes.Slug)
	}
	fmt.Printf("Updated: %s\n", post.Attributes.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}

func runStrapiBlogPublish(cmd *cobra.Command, args []string) error {
	// Suppress INFO/DEBUG logs if JSON output is requested
	if blogOutputJSON {
		defer logger.SuppressInfoLogsForJSON()()
	}

	ctx := context.Background()

	// Get Strapi configuration
	strapiURL := viper.GetString("strapi.url")
	if strapiURL == "" {
		return fmt.Errorf("Strapi URL not configured. Use 'ainative-code strapi config --url <url>' to set it up")
	}

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(strapiURL),
		client.WithTimeout(30*time.Second),
	)

	strapiClient := strapi.New(
		strapi.WithAPIClient(apiClient),
		strapi.WithBaseURL(strapiURL),
	)

	logger.InfoEvent().
		Int("id", blogID).
		Msg("Publishing blog post")

	// Publish blog post
	post, err := strapiClient.PublishBlogPost(ctx, blogID)
	if err != nil {
		return fmt.Errorf("failed to publish blog post: %w", err)
	}

	// Output result
	if blogOutputJSON {
		return outputAsJSON(post)
	}

	fmt.Println("Blog post published successfully!")
	fmt.Printf("\nID: %d\n", post.ID)
	fmt.Printf("Title: %s\n", post.Attributes.Title)
	fmt.Printf("Status: %s\n", post.Attributes.Status)
	if post.Attributes.PublishedAt != nil {
		fmt.Printf("Published: %s\n", post.Attributes.PublishedAt.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func runStrapiBlogDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get Strapi configuration
	strapiURL := viper.GetString("strapi.url")
	if strapiURL == "" {
		return fmt.Errorf("Strapi URL not configured. Use 'ainative-code strapi config --url <url>' to set it up")
	}

	// Create API client
	apiClient := client.New(
		client.WithBaseURL(strapiURL),
		client.WithTimeout(30*time.Second),
	)

	strapiClient := strapi.New(
		strapi.WithAPIClient(apiClient),
		strapi.WithBaseURL(strapiURL),
	)

	// Confirmation prompt
	fmt.Printf("WARNING: You are about to delete blog post ID %d. This action cannot be undone.\n", blogID)
	fmt.Print("Type 'yes' to confirm: ")

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	logger.InfoEvent().
		Int("id", blogID).
		Msg("Deleting blog post")

	// Delete blog post
	if err := strapiClient.DeleteBlogPost(ctx, blogID); err != nil {
		return fmt.Errorf("failed to delete blog post: %w", err)
	}

	fmt.Printf("Blog post ID %d deleted successfully.\n", blogID)
	return nil
}

// processContentInput processes content input, handling @filename syntax
func processContentInput(input string) (string, error) {
	if strings.HasPrefix(input, "@") {
		// Read from file
		filename := strings.TrimPrefix(input, "@")
		data, err := os.ReadFile(filename)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", filename, err)
		}
		return string(data), nil
	}
	return input, nil
}
