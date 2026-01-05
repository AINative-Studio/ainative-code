// Package client provides a unified HTTP client for AINative platform API interactions.
//
// The client package implements:
//   - JWT bearer token authentication
//   - Automatic token refresh on 401 Unauthorized responses
//   - Request/response logging for observability
//   - Configurable retry logic with exponential backoff
//   - Timeout configuration per request
//   - Base URL configuration for different AINative services
//
// Basic Usage:
//
//	authClient := auth.NewClient(authConfig)
//	tokens, err := authClient.GetStoredTokens(ctx)
//
//	client := client.New(
//	    client.WithAuthClient(authClient),
//	    client.WithBaseURL("https://api.ainative.studio"),
//	    client.WithTimeout(30 * time.Second),
//	)
//
//	resp, err := client.Get(ctx, "/api/v1/resource")
//
// The client automatically:
//   - Injects JWT bearer tokens into requests
//   - Refreshes expired tokens when receiving 401 responses
//   - Logs all requests and responses
//   - Retries failed requests with exponential backoff
//   - Respects context cancellation and timeouts
package client
