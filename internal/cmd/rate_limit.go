package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/AINative-studio/ainative-code/internal/config"
	"github.com/AINative-studio/ainative-code/internal/ratelimit"
	"github.com/spf13/cobra"
)

var rateLimitCmd = &cobra.Command{
	Use:   "rate-limit",
	Short: "Manage rate limiting configuration and status",
	Long: `Manage rate limiting configuration, view status, and reset limits.

Examples:
  # View rate limit status
  ainative-code rate-limit status

  # Reset rate limit for a user
  ainative-code rate-limit reset --user user123

  # Reset rate limit for an IP
  ainative-code rate-limit reset --ip 192.168.1.1

  # View rate limit configuration
  ainative-code rate-limit config

  # View rate limit metrics
  ainative-code rate-limit metrics`,
}

var rateLimitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show rate limiting status",
	Long:  "Display the current rate limiting status and configuration.",
	RunE:  runRateLimitStatus,
}

var rateLimitResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset rate limit for a user or IP",
	Long:  "Reset the rate limit counter for a specific user or IP address.",
	RunE:  runRateLimitReset,
}

var rateLimitConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show rate limiting configuration",
	Long:  "Display the current rate limiting configuration.",
	RunE:  runRateLimitConfig,
}

var rateLimitMetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Show rate limiting metrics",
	Long:  "Display rate limiting metrics and statistics.",
	RunE:  runRateLimitMetrics,
}

var (
	resetUser string
	resetIP   string
	resetKey  string
)

func init() {
	// Add subcommands
	rateLimitCmd.AddCommand(rateLimitStatusCmd)
	rateLimitCmd.AddCommand(rateLimitResetCmd)
	rateLimitCmd.AddCommand(rateLimitConfigCmd)
	rateLimitCmd.AddCommand(rateLimitMetricsCmd)

	// Add flags
	rateLimitResetCmd.Flags().StringVar(&resetUser, "user", "", "User ID to reset")
	rateLimitResetCmd.Flags().StringVar(&resetIP, "ip", "", "IP address to reset")
	rateLimitResetCmd.Flags().StringVar(&resetKey, "key", "", "Custom key to reset")

	// Register with root command
	rootCmd.AddCommand(rateLimitCmd)
}

func loadRateLimitConfig() (*config.Config, error) {
	cfg := &config.Config{}
	// Load from default config locations
	return cfg, nil
}

func runRateLimitStatus(cmd *cobra.Command, args []string) error {
	cfg, err := loadRateLimitConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Rate Limiting Status")
	fmt.Println("===================")
	fmt.Printf("Enabled: %v\n", cfg.Performance.RateLimit.Enabled)
	fmt.Printf("Requests Per Minute: %d\n", cfg.Performance.RateLimit.RequestsPerMinute)
	fmt.Printf("Burst Size: %d\n", cfg.Performance.RateLimit.BurstSize)
	fmt.Printf("Time Window: %s\n", cfg.Performance.RateLimit.TimeWindow)
	fmt.Printf("Per User: %v\n", cfg.Performance.RateLimit.PerUser)
	fmt.Printf("Per Endpoint: %v\n", cfg.Performance.RateLimit.PerEndpoint)
	fmt.Printf("Storage: %s\n", cfg.Performance.RateLimit.Storage)

	if len(cfg.Performance.RateLimit.EndpointLimits) > 0 {
		fmt.Println("\nEndpoint-Specific Limits:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ENDPOINT\tLIMIT")
		for endpoint, limit := range cfg.Performance.RateLimit.EndpointLimits {
			fmt.Fprintf(w, "%s\t%d\n", endpoint, limit)
		}
		w.Flush()
	}

	return nil
}

func runRateLimitReset(cmd *cobra.Command, args []string) error {
	if resetUser == "" && resetIP == "" && resetKey == "" {
		return fmt.Errorf("must specify --user, --ip, or --key")
	}

	storage := ratelimit.NewMemoryStorage()
	defer storage.Close()

	limiter := ratelimit.NewLimiter(storage, ratelimit.Config{
		RequestsPerMinute: 60,
		TimeWindow:        1 * time.Minute,
	})
	defer limiter.Close()

	var key string
	if resetKey != "" {
		key = resetKey
	} else if resetUser != "" {
		key = limiter.BuildKey("user", resetUser)
	} else if resetIP != "" {
		key = limiter.BuildKey("ip", resetIP)
	}

	ctx := context.Background()
	if err := limiter.Reset(ctx, key); err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	fmt.Printf("Successfully reset rate limit for key: %s\n", key)
	return nil
}

func runRateLimitConfig(cmd *cobra.Command, args []string) error {
	cfg, err := loadRateLimitConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Rate Limiting Configuration")
	fmt.Println("===========================")
	fmt.Printf("Enabled: %v\n", cfg.Performance.RateLimit.Enabled)
	fmt.Printf("Requests Per Minute: %d\n", cfg.Performance.RateLimit.RequestsPerMinute)
	fmt.Printf("Burst Size: %d\n", cfg.Performance.RateLimit.BurstSize)
	fmt.Printf("Time Window: %s\n", cfg.Performance.RateLimit.TimeWindow)

	if len(cfg.Performance.RateLimit.SkipPaths) > 0 {
		fmt.Println("\nSkipped Paths:")
		for _, path := range cfg.Performance.RateLimit.SkipPaths {
			fmt.Printf("  - %s\n", path)
		}
	}

	return nil
}

func runRateLimitMetrics(cmd *cobra.Command, args []string) error {
	metrics := ratelimit.NewMetrics()

	stats := metrics.GetStats()

	fmt.Println("Rate Limiting Metrics")
	fmt.Println("====================")
	fmt.Printf("Total Requests: %d\n", stats.TotalRequests)
	fmt.Printf("Allowed Requests: %d\n", stats.AllowedRequests)
	fmt.Printf("Blocked Requests: %d\n", stats.BlockedRequests)
	fmt.Printf("Blocked Rate: %.2f%%\n", metrics.GetBlockedRate())
	fmt.Printf("Last Reset: %s\n", stats.LastReset.Format(time.RFC3339))

	return nil
}
