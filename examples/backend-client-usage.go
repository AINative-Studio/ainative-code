package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AINative-studio/ainative-code/internal/backend"
)

// This example demonstrates how the Go CLI will use the backend client
// to communicate with the Python backend at http://localhost:8000
func main() {
	// Initialize client with Python backend URL
	client := backend.NewClient("http://localhost:8000")
	ctx := context.Background()

	// Example 1: Register a new user
	fmt.Println("=== Example 1: Register ===")
	registerResp, err := client.Register(ctx, "demo@example.com", "SecurePass123!")
	if err != nil {
		log.Printf("Registration failed: %v", err)
	} else {
		fmt.Printf("✓ Registered successfully\n")
		fmt.Printf("  User ID: %s\n", registerResp.User.ID)
		fmt.Printf("  Email: %s\n", registerResp.User.Email)
		fmt.Printf("  Access Token: %s...\n", registerResp.AccessToken[:20])
	}

	// Example 2: Login with existing user
	fmt.Println("\n=== Example 2: Login ===")
	loginResp, err := client.Login(ctx, "demo@example.com", "SecurePass123!")
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Printf("✓ Logged in successfully\n")
	fmt.Printf("  User ID: %s\n", loginResp.User.ID)
	fmt.Printf("  Email: %s\n", loginResp.User.Email)

	token := loginResp.AccessToken
	refreshToken := loginResp.RefreshToken

	// Example 3: Get current user info
	fmt.Println("\n=== Example 3: Get Current User ===")
	user, err := client.GetMe(ctx, token)
	if err != nil {
		log.Printf("GetMe failed: %v", err)
	} else {
		fmt.Printf("✓ Current user retrieved\n")
		fmt.Printf("  User ID: %s\n", user.ID)
		fmt.Printf("  Email: %s\n", user.Email)
	}

	// Example 4: Send a chat completion request
	fmt.Println("\n=== Example 4: Chat Completion ===")
	chatReq := &backend.ChatCompletionRequest{
		Messages: []backend.Message{
			{Role: "user", Content: "What is the capital of France?"},
		},
		Model:       "claude-sonnet-4-5",
		Temperature: 0.7,
		MaxTokens:   100,
	}

	chatResp, err := client.ChatCompletion(ctx, token, chatReq)
	if err != nil {
		log.Printf("Chat completion failed: %v", err)
	} else {
		fmt.Printf("✓ Chat completion successful\n")
		fmt.Printf("  Model: %s\n", chatResp.Model)
		fmt.Printf("  Response ID: %s\n", chatResp.ID)
		fmt.Printf("  Assistant: %s\n", chatResp.Choices[0].Message.Content)
	}

	// Example 5: Refresh access token
	fmt.Println("\n=== Example 5: Refresh Token ===")
	refreshResp, err := client.RefreshToken(ctx, refreshToken)
	if err != nil {
		log.Printf("Token refresh failed: %v", err)
	} else {
		fmt.Printf("✓ Token refreshed successfully\n")
		fmt.Printf("  New Access Token: %s...\n", refreshResp.AccessToken[:20])
		token = refreshResp.AccessToken
	}

	// Example 6: Health check
	fmt.Println("\n=== Example 6: Health Check ===")
	err = client.HealthCheck(ctx)
	if err != nil {
		log.Printf("Health check failed: %v", err)
	} else {
		fmt.Printf("✓ Backend is healthy\n")
	}

	// Example 7: Logout
	fmt.Println("\n=== Example 7: Logout ===")
	err = client.Logout(ctx, token)
	if err != nil {
		log.Printf("Logout failed: %v", err)
	} else {
		fmt.Printf("✓ Logged out successfully\n")
	}

	// Example 8: Custom timeout configuration
	fmt.Println("\n=== Example 8: Custom Timeout ===")
	customClient := backend.NewClient("http://localhost:8000", backend.WithTimeout(60*time.Second))
	fmt.Printf("✓ Client created with 60s timeout\n")
	fmt.Printf("  Timeout: %v\n", customClient.Timeout)

	// Example 9: Error handling - Invalid credentials
	fmt.Println("\n=== Example 9: Error Handling ===")
	_, err = client.Login(ctx, "demo@example.com", "wrongpassword")
	if err != nil {
		fmt.Printf("✓ Error handled correctly: %v\n", err)
	}

	fmt.Println("\n=== All Examples Complete ===")
}
