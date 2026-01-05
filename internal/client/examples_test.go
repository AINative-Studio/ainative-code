package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth"
	"github.com/AINative-studio/ainative-code/internal/client"
)

// Example_zeroDBClient demonstrates using the API client for ZeroDB NoSQL operations
func Example_zeroDBClient() {
	// Create auth client (simplified for example)
	authClient := &mockAuthClient{
		tokens: &auth.TokenPair{
			AccessToken: &auth.AccessToken{
				Raw:       "your-access-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
		},
	}

	// Create API client configured for ZeroDB
	apiClient := client.New(
		client.WithBaseURL("https://api.ainative.studio"),
		client.WithAuthClient(authClient),
		client.WithTimeout(30*time.Second),
		client.WithMaxRetries(3),
	)

	ctx := context.Background()

	// Example: Create a NoSQL table
	createTablePayload := map[string]interface{}{
		"name": "users",
		"schema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name":  map[string]string{"type": "string"},
				"email": map[string]string{"type": "string"},
				"age":   map[string]string{"type": "number"},
			},
		},
	}

	resp, err := apiClient.Post(ctx, "/api/v1/projects/my-project/nosql/tables", createTablePayload)
	if err != nil {
		fmt.Printf("Failed to create table: %v\n", err)
		return
	}

	var createResult map[string]interface{}
	json.Unmarshal(resp, &createResult)
	fmt.Printf("Table created: %v\n", createResult)

	// Example: Insert a document
	insertPayload := map[string]interface{}{
		"table": "users",
		"data": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		},
	}

	resp, err = apiClient.Post(ctx, "/api/v1/projects/my-project/nosql/documents", insertPayload)
	if err != nil {
		fmt.Printf("Failed to insert document: %v\n", err)
		return
	}

	var insertResult map[string]interface{}
	json.Unmarshal(resp, &insertResult)
	fmt.Printf("Document inserted: %v\n", insertResult)

	// Example: Query documents with MongoDB-style filter
	queryPayload := map[string]interface{}{
		"table": "users",
		"filter": map[string]interface{}{
			"age": map[string]interface{}{
				"$gte": 18,
				"$lt":  65,
			},
		},
		"limit":  10,
		"offset": 0,
	}

	resp, err = apiClient.Post(ctx, "/api/v1/projects/my-project/nosql/query", queryPayload)
	if err != nil {
		fmt.Printf("Failed to query documents: %v\n", err)
		return
	}

	var queryResult map[string]interface{}
	json.Unmarshal(resp, &queryResult)
	fmt.Printf("Query results: %v\n", queryResult)
}

// Example_designServiceClient demonstrates using the API client for Design Service
func Example_designServiceClient() {
	// Create auth client
	authClient := &mockAuthClient{
		tokens: &auth.TokenPair{
			AccessToken: &auth.AccessToken{
				Raw:       "your-access-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
		},
	}

	// Create API client configured for Design Service
	apiClient := client.New(
		client.WithBaseURL("https://design.ainative.studio"),
		client.WithAuthClient(authClient),
		client.WithTimeout(60*time.Second), // Longer timeout for design operations
		client.WithMaxRetries(3),
	)

	ctx := context.Background()

	// Example: Generate a design from prompt
	designPayload := map[string]interface{}{
		"prompt":      "Create a modern login page with dark mode",
		"style":       "minimalist",
		"colorScheme": "blue",
		"framework":   "tailwind",
	}

	resp, err := apiClient.Post(ctx, "/api/v1/generate", designPayload)
	if err != nil {
		fmt.Printf("Failed to generate design: %v\n", err)
		return
	}

	var designResult map[string]interface{}
	json.Unmarshal(resp, &designResult)
	fmt.Printf("Design generated: %v\n", designResult)

	// Example: Get design by ID
	designID := "design-123"
	resp, err = apiClient.Get(ctx, fmt.Sprintf("/api/v1/designs/%s", designID))
	if err != nil {
		fmt.Printf("Failed to get design: %v\n", err)
		return
	}

	var design map[string]interface{}
	json.Unmarshal(resp, &design)
	fmt.Printf("Design details: %v\n", design)
}

// Example_strapiCMSClient demonstrates using the API client for Strapi CMS
func Example_strapiCMSClient() {
	// Create auth client
	authClient := &mockAuthClient{
		tokens: &auth.TokenPair{
			AccessToken: &auth.AccessToken{
				Raw:       "your-access-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
		},
	}

	// Create API client configured for Strapi
	apiClient := client.New(
		client.WithBaseURL("https://cms.ainative.studio"),
		client.WithAuthClient(authClient),
		client.WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Example: Create a blog post
	blogPostPayload := map[string]interface{}{
		"data": map[string]interface{}{
			"title":   "Getting Started with AINative Code",
			"content": "AINative Code is a powerful CLI tool...",
			"author":  "John Doe",
			"tags":    []string{"tutorial", "getting-started"},
			"status":  "draft",
		},
	}

	resp, err := apiClient.Post(ctx, "/api/blog-posts", blogPostPayload)
	if err != nil {
		fmt.Printf("Failed to create blog post: %v\n", err)
		return
	}

	var post map[string]interface{}
	json.Unmarshal(resp, &post)
	fmt.Printf("Blog post created: %v\n", post)

	// Example: Query blog posts with filters and pagination
	resp, err = apiClient.Get(ctx, "/api/blog-posts",
		client.WithQueryParam("filters[status][$eq]", "published"),
		client.WithQueryParam("sort", "publishedAt:desc"),
		client.WithQueryParam("pagination[page]", "1"),
		client.WithQueryParam("pagination[pageSize]", "10"),
	)
	if err != nil {
		fmt.Printf("Failed to query blog posts: %v\n", err)
		return
	}

	var posts map[string]interface{}
	json.Unmarshal(resp, &posts)
	fmt.Printf("Blog posts: %v\n", posts)

	// Example: Update a blog post
	postID := "1"
	updatePayload := map[string]interface{}{
		"data": map[string]interface{}{
			"status": "published",
		},
	}

	resp, err = apiClient.Put(ctx, fmt.Sprintf("/api/blog-posts/%s", postID), updatePayload)
	if err != nil {
		fmt.Printf("Failed to update blog post: %v\n", err)
		return
	}

	var updatedPost map[string]interface{}
	json.Unmarshal(resp, &updatedPost)
	fmt.Printf("Blog post updated: %v\n", updatedPost)
}

// Example_rlhfClient demonstrates using the API client for RLHF (Reinforcement Learning from Human Feedback)
func Example_rlhfClient() {
	// Create auth client
	authClient := &mockAuthClient{
		tokens: &auth.TokenPair{
			AccessToken: &auth.AccessToken{
				Raw:       "your-access-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
		},
	}

	// Create API client configured for RLHF service
	apiClient := client.New(
		client.WithBaseURL("https://rlhf.ainative.studio"),
		client.WithAuthClient(authClient),
		client.WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Example: Submit feedback for a model response
	feedbackPayload := map[string]interface{}{
		"sessionID":  "session-123",
		"responseID": "resp-456",
		"feedback": map[string]interface{}{
			"rating": 5,
			"helpful": true,
			"comment": "This response was very helpful and accurate",
			"categories": []string{"accuracy", "clarity", "helpfulness"},
		},
		"context": map[string]interface{}{
			"userQuery": "How do I implement OAuth 2.0?",
			"modelResponse": "OAuth 2.0 is an authorization framework...",
		},
	}

	resp, err := apiClient.Post(ctx, "/api/v1/feedback", feedbackPayload)
	if err != nil {
		fmt.Printf("Failed to submit feedback: %v\n", err)
		return
	}

	var feedbackResult map[string]interface{}
	json.Unmarshal(resp, &feedbackResult)
	fmt.Printf("Feedback submitted: %v\n", feedbackResult)

	// Example: Get feedback analytics
	resp, err = apiClient.Get(ctx, "/api/v1/analytics/feedback",
		client.WithQueryParam("startDate", "2026-01-01"),
		client.WithQueryParam("endDate", "2026-01-31"),
		client.WithQueryParam("aggregateBy", "day"),
	)
	if err != nil {
		fmt.Printf("Failed to get analytics: %v\n", err)
		return
	}

	var analytics map[string]interface{}
	json.Unmarshal(resp, &analytics)
	fmt.Printf("Analytics: %v\n", analytics)

	// Example: Submit a preference comparison (A/B testing)
	comparisonPayload := map[string]interface{}{
		"sessionID": "session-789",
		"prompt":    "Explain dependency injection",
		"responseA": map[string]interface{}{
			"id":      "resp-a-1",
			"content": "Dependency injection is a design pattern...",
		},
		"responseB": map[string]interface{}{
			"id":      "resp-b-1",
			"content": "DI is when you pass dependencies to a class...",
		},
		"preferred": "A",
		"reason":    "More detailed and professional explanation",
	}

	resp, err = apiClient.Post(ctx, "/api/v1/comparisons", comparisonPayload)
	if err != nil {
		fmt.Printf("Failed to submit comparison: %v\n", err)
		return
	}

	var comparison map[string]interface{}
	json.Unmarshal(resp, &comparison)
	fmt.Printf("Comparison submitted: %v\n", comparison)
}

// Example_customOptions demonstrates advanced client configuration
func Example_customOptions() {
	// Create auth client
	authClient := &mockAuthClient{
		tokens: &auth.TokenPair{
			AccessToken: &auth.AccessToken{
				Raw:       "your-access-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
		},
	}

	// Create client with custom HTTP client and options
	apiClient := client.New(
		client.WithBaseURL("https://api.ainative.studio"),
		client.WithAuthClient(authClient),
		client.WithTimeout(60*time.Second),
		client.WithMaxRetries(5), // More retries for flaky networks
	)

	ctx := context.Background()

	// Example: Make request with custom headers
	resp, err := apiClient.Get(ctx, "/api/v1/users/me",
		client.WithHeader("X-Request-ID", "unique-request-id"),
		client.WithHeader("X-Client-Version", "1.0.0"),
	)
	if err != nil {
		fmt.Printf("Failed to get user: %v\n", err)
		return
	}

	var user map[string]interface{}
	json.Unmarshal(resp, &user)
	fmt.Printf("User: %v\n", user)

	// Example: Make unauthenticated request (public endpoint)
	resp, err = apiClient.Get(ctx, "/api/v1/public/status",
		client.WithSkipAuth(), // Don't inject JWT token
	)
	if err != nil {
		fmt.Printf("Failed to get status: %v\n", err)
		return
	}

	var status map[string]interface{}
	json.Unmarshal(resp, &status)
	fmt.Printf("Status: %v\n", status)

	// Example: Make request with multiple query parameters
	resp, err = apiClient.Get(ctx, "/api/v1/search",
		client.WithQueryParam("q", "golang"),
		client.WithQueryParam("type", "repository"),
		client.WithQueryParam("sort", "stars"),
		client.WithQueryParam("order", "desc"),
	)
	if err != nil {
		fmt.Printf("Failed to search: %v\n", err)
		return
	}

	var searchResults map[string]interface{}
	json.Unmarshal(resp, &searchResults)
	fmt.Printf("Search results: %v\n", searchResults)
}

// Example_errorHandling demonstrates error handling patterns
func Example_errorHandling() {
	authClient := &mockAuthClient{
		tokens: &auth.TokenPair{
			AccessToken: &auth.AccessToken{
				Raw:       "your-access-token",
				ExpiresAt: time.Now().Add(1 * time.Hour),
				UserID:    "user-123",
				Email:     "user@example.com",
				Issuer:    "ainative-auth",
				Audience:  "ainative-code",
			},
		},
	}

	apiClient := client.New(
		client.WithBaseURL("https://api.ainative.studio"),
		client.WithAuthClient(authClient),
		client.WithTimeout(30*time.Second),
		client.WithMaxRetries(3),
	)

	ctx := context.Background()

	// Example: Handle different error types
	resp, err := apiClient.Get(ctx, "/api/v1/resource")
	if err != nil {
		// Check for specific error types
		if client.IsAuthError(err) {
			fmt.Println("Authentication error - please re-authenticate")
			// Re-authenticate user
			return
		}

		if client.IsRetryable(err) {
			fmt.Println("Retryable error - the client already retried automatically")
			// You could implement additional retry logic here
			return
		}

		// Check for HTTP errors
		var httpErr *client.HTTPError
		if httpError, ok := err.(*client.HTTPError); ok {
			httpErr = httpError
			switch {
			case httpErr.StatusCode == 404:
				fmt.Println("Resource not found")
			case httpErr.StatusCode == 403:
				fmt.Println("Access forbidden - check permissions")
			case httpErr.StatusCode >= 500:
				fmt.Println("Server error - try again later")
			default:
				fmt.Printf("HTTP error %d: %s\n", httpErr.StatusCode, httpErr.Message)
			}
			return
		}

		// Generic error handling
		fmt.Printf("Request failed: %v\n", err)
		return
	}

	// Success - process response
	var result map[string]interface{}
	json.Unmarshal(resp, &result)
	fmt.Printf("Success: %v\n", result)
}
