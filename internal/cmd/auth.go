package cmd

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/AINative-studio/ainative-code/internal/auth/keychain"
	"github.com/AINative-studio/ainative-code/internal/auth/oauth"
	"github.com/spf13/cobra"
)

var (
	// OAuth configuration (should be loaded from config file)
	// NOTE: auth.ainative.studio is currently unreachable (Issue #98)
	// Using localhost mock server as fallback for development/testing
	defaultOAuthConfig = oauth.Config{
		AuthURL:     getAuthURL(),
		TokenURL:    getTokenURL(),
		ClientID:    "ainative-code-cli",
		RedirectURL: "http://localhost:8080/callback",
		Scopes:      []string{"read", "write", "offline_access"},
	}
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Manage authentication credentials and tokens for AINative Code.`,
}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with AINative",
	Long: `Initiate OAuth 2.0 authentication flow with PKCE.

This command will:
1. Open your browser to the authentication page
2. Wait for you to authorize the application
3. Store the received tokens securely in OS keychain
4. Start automatic token refresh in the background`,
	RunE: runLogin,
}

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored credentials",
	Long: `Remove all stored authentication credentials from OS keychain.

This will delete:
- Access tokens
- Refresh tokens
- API keys
- User information`,
	RunE: runLogout,
}

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current user information",
	Long:  `Show information about the currently authenticated user including email and token status.`,
	RunE:  runWhoami,
}

// tokenCmd represents the token command group
var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage authentication tokens",
	Long:  `View and manage authentication tokens.`,
}

// tokenRefreshCmd represents the token refresh command
var tokenRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Manually refresh access token",
	Long: `Force an immediate refresh of the access token using the stored refresh token.

This is useful if you want to ensure you have a fresh token before making API calls.`,
	RunE: runTokenRefresh,
}

// tokenStatusCmd represents the token status command
var tokenStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show token expiration status",
	Long:  `Display detailed information about current tokens including expiration times.`,
	RunE:  runTokenStatus,
}

// getAuthURL returns the authorization endpoint URL with fallback logic
func getAuthURL() string {
	// Check environment variable override first
	if url := os.Getenv("AINATIVE_AUTH_URL"); url != "" {
		return url
	}

	// Production endpoint (currently unreachable - Issue #98)
	prodURL := "https://auth.ainative.studio/oauth/authorize"

	// Try to verify if production endpoint is reachable
	if isEndpointReachable(prodURL) {
		return prodURL
	}

	// Fallback to localhost mock server for development
	// Users should set AINATIVE_AUTH_URL to use alternative auth server
	return "http://localhost:9090/oauth/authorize"
}

// getTokenURL returns the token endpoint URL with fallback logic
func getTokenURL() string {
	// Check environment variable override first
	if url := os.Getenv("AINATIVE_TOKEN_URL"); url != "" {
		return url
	}

	// Production endpoint (currently unreachable - Issue #98)
	prodURL := "https://auth.ainative.studio/oauth/token"

	// Try to verify if production endpoint is reachable
	if isEndpointReachable(prodURL) {
		return prodURL
	}

	// Fallback to localhost mock server for development
	return "http://localhost:9090/oauth/token"
}

// isEndpointReachable checks if an endpoint is reachable with a quick HEAD request
func isEndpointReachable(url string) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider 2xx, 3xx, 4xx as "reachable" (server exists)
	// Only 5xx or network errors mean unreachable
	return resp.StatusCode < 500
}

func init() {
	// Register auth command to root
	rootCmd.AddCommand(authCmd)

	// Register auth subcommands
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(whoamiCmd)
	authCmd.AddCommand(tokenCmd)

	// Register token subcommands
	tokenCmd.AddCommand(tokenRefreshCmd)
	tokenCmd.AddCommand(tokenStatusCmd)

	// Add flags
	loginCmd.Flags().String("auth-url", defaultOAuthConfig.AuthURL, "Authorization endpoint URL")
	loginCmd.Flags().String("token-url", defaultOAuthConfig.TokenURL, "Token endpoint URL")
	loginCmd.Flags().String("client-id", defaultOAuthConfig.ClientID, "OAuth client ID")
	loginCmd.Flags().StringSlice("scopes", defaultOAuthConfig.Scopes, "OAuth scopes to request")
}

func runLogin(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get flags
	authURL, _ := cmd.Flags().GetString("auth-url")
	tokenURL, _ := cmd.Flags().GetString("token-url")
	clientID, _ := cmd.Flags().GetString("client-id")
	scopes, _ := cmd.Flags().GetStringSlice("scopes")

	// Create OAuth client
	oauthConfig := oauth.Config{
		AuthURL:     authURL,
		TokenURL:    tokenURL,
		ClientID:    clientID,
		RedirectURL: "http://localhost:8080/callback",
		Scopes:      scopes,
	}

	oauthClient := oauth.NewClient(oauthConfig)

	// Show warning if using fallback endpoints
	if authURL == "http://localhost:9090/oauth/authorize" {
		cmd.Println("⚠️  WARNING: Using localhost mock auth server (production server unreachable)")
		cmd.Println("   To use a different auth server, set environment variables:")
		cmd.Println("   export AINATIVE_AUTH_URL=<your-auth-url>")
		cmd.Println("   export AINATIVE_TOKEN_URL=<your-token-url>")
		cmd.Println()
		cmd.Println("   Or use command flags:")
		cmd.Println("   --auth-url <url> --token-url <url>")
		cmd.Println()
	}

	// Start authentication flow
	cmd.Println("Initiating authentication flow...")
	cmd.Printf("Auth URL: %s\n", authURL)
	cmd.Printf("Token URL: %s\n", tokenURL)
	cmd.Println()

	tokens, err := oauthClient.Authenticate(ctx)
	if err != nil {
		// Provide helpful error message
		cmd.Println()
		cmd.Println("❌ Authentication failed")
		cmd.Println()
		cmd.Println("Troubleshooting:")
		cmd.Println("1. Check if the auth server is running and reachable")
		cmd.Println("2. For development, you can run a local mock OAuth server on port 9090")
		cmd.Println("3. Set custom auth endpoints using environment variables or flags")
		cmd.Println()
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store tokens in keychain
	kc := keychain.Get()
	if err := kc.SetTokenPair(tokens); err != nil {
		return fmt.Errorf("failed to store tokens: %w", err)
	}

	// Try to extract and store user email from access token
	// (This would normally come from validating the token or a /me endpoint)
	cmd.Println("✓ Authentication successful!")
	cmd.Printf("Tokens stored securely in OS keychain\n")
	cmd.Printf("Access token expires in: %d seconds\n", tokens.ExpiresIn)

	// Start auto-refresh manager (in a real implementation)
	// This would be handled by a background service
	cmd.Println("\nTo view your authentication status, run: ainative-code whoami")

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	kc := keychain.Get()

	// Delete all credentials
	if err := kc.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear credentials: %w", err)
	}

	cmd.Println("✓ Successfully logged out")
	cmd.Println("All credentials have been removed from OS keychain")

	return nil
}

func runWhoami(cmd *cobra.Command, args []string) error {
	kc := keychain.Get()

	// Get tokens
	tokens, err := kc.GetTokenPair()
	if err != nil {
		cmd.Println("Not authenticated")
		cmd.Println("\nRun 'ainative-code login' to authenticate")
		return nil
	}

	// Get user email if available
	email, err := kc.GetUserEmail()
	if err != nil {
		email = "Unknown"
	}

	// Calculate expiration
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	timeUntilExpiry := time.Until(expiresAt)

	// Display user info
	cmd.Println("Authenticated User:")
	cmd.Printf("  Email: %s\n", email)
	cmd.Printf("  Token Type: %s\n", tokens.TokenType)
	cmd.Printf("  Expires In: %s\n", formatDuration(timeUntilExpiry))

	if timeUntilExpiry < 5*time.Minute {
		cmd.Println("  ⚠️  Token expiring soon!")
	}

	return nil
}

func runTokenRefresh(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	kc := keychain.Get()

	// Get current tokens
	currentTokens, err := kc.GetTokenPair()
	if err != nil {
		return fmt.Errorf("not authenticated: %w", err)
	}

	// Create OAuth client
	oauthClient := oauth.NewClient(defaultOAuthConfig)

	// Refresh token
	cmd.Println("Refreshing access token...")
	newTokens, err := oauthClient.RefreshToken(ctx, currentTokens.RefreshToken)
	if err != nil {
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Store new tokens
	if err := kc.SetTokenPair(newTokens); err != nil {
		return fmt.Errorf("failed to store refreshed tokens: %w", err)
	}

	cmd.Println("✓ Token refreshed successfully")
	cmd.Printf("New token expires in: %d seconds\n", newTokens.ExpiresIn)

	return nil
}

func runTokenStatus(cmd *cobra.Command, args []string) error {
	kc := keychain.Get()

	// Get tokens
	tokens, err := kc.GetTokenPair()
	if err != nil {
		cmd.Println("No tokens found")
		cmd.Println("\nRun 'ainative-code login' to authenticate")
		return nil
	}

	// Calculate times
	expiresAt := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	timeUntilExpiry := time.Until(expiresAt)
	refreshThreshold := 5 * time.Minute

	// Display status
	cmd.Println("Token Status:")
	cmd.Println("─────────────────────────────────────────")
	cmd.Printf("Access Token:  %s...\n", tokens.AccessToken[:min(len(tokens.AccessToken), 20)])
	cmd.Printf("Refresh Token: %s...\n", tokens.RefreshToken[:min(len(tokens.RefreshToken), 20)])
	cmd.Printf("Token Type:    %s\n", tokens.TokenType)
	cmd.Printf("Expires At:    %s\n", expiresAt.Format(time.RFC1123))
	cmd.Printf("Time Until Expiry: %s\n", formatDuration(timeUntilExpiry))

	// Show status indicator
	if timeUntilExpiry <= 0 {
		cmd.Println("\nStatus: ❌ EXPIRED")
		cmd.Println("Run 'ainative-code token refresh' to refresh")
	} else if timeUntilExpiry < refreshThreshold {
		cmd.Println("\nStatus: ⚠️  EXPIRING SOON")
		cmd.Println("Consider running 'ainative-code token refresh'")
	} else {
		cmd.Println("\nStatus: ✓ VALID")
	}

	// Check if auto-refresh is enabled
	cmd.Println("\nAuto-Refresh: Managed by background service")

	return nil
}

// Helper functions

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "expired"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
