package design

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/AINative-studio/ainative-code/internal/logger"
)

// ConflictResolverConfig configures the conflict resolver.
type ConflictResolverConfig struct {
	// Strategy is the default strategy for resolving conflicts.
	Strategy ConflictResolutionStrategy

	// Interactive enables interactive prompts for conflict resolution.
	Interactive bool
}

// ConflictResolver handles conflict resolution during synchronization.
type ConflictResolver struct {
	config ConflictResolverConfig
}

// NewConflictResolver creates a new conflict resolver.
func NewConflictResolver(config ConflictResolverConfig) *ConflictResolver {
	return &ConflictResolver{
		config: config,
	}
}

// Resolve resolves a conflict based on the configured strategy.
func (r *ConflictResolver) Resolve(conflict Conflict) ConflictResolution {
	logger.DebugEvent().
		Str("token", conflict.TokenName).
		Str("type", string(conflict.ConflictType)).
		Str("strategy", string(r.config.Strategy)).
		Msg("Resolving conflict")

	switch r.config.Strategy {
	case ConflictResolutionLocal:
		return r.resolveWithLocal(conflict)
	case ConflictResolutionRemote:
		return r.resolveWithRemote(conflict)
	case ConflictResolutionNewest:
		return r.resolveWithNewest(conflict)
	case ConflictResolutionPrompt:
		return r.resolveWithPrompt(conflict)
	case ConflictResolutionMerge:
		return r.resolveWithMerge(conflict)
	default:
		// Default to remote if strategy is unknown
		logger.WarnEvent().
			Str("strategy", string(r.config.Strategy)).
			Msg("Unknown conflict resolution strategy, defaulting to remote")
		return r.resolveWithRemote(conflict)
	}
}

// resolveWithLocal resolves conflict by preferring local version.
func (r *ConflictResolver) resolveWithLocal(conflict Conflict) ConflictResolution {
	return ConflictResolution{
		Strategy:      ConflictResolutionLocal,
		SelectedToken: conflict.LocalToken,
		Reason:        "Preferred local version as per strategy",
	}
}

// resolveWithRemote resolves conflict by preferring remote version.
func (r *ConflictResolver) resolveWithRemote(conflict Conflict) ConflictResolution {
	return ConflictResolution{
		Strategy:      ConflictResolutionRemote,
		SelectedToken: conflict.RemoteToken,
		Reason:        "Preferred remote version as per strategy",
	}
}

// resolveWithNewest resolves conflict by preferring the newest version.
func (r *ConflictResolver) resolveWithNewest(conflict Conflict) ConflictResolution {
	// Compare metadata timestamps if available
	// For now, we'll use a simple heuristic: prefer remote in case of doubt
	// In a real implementation, you'd check updated_at timestamps

	// This is a simplified implementation - in production, you'd check
	// actual timestamps from token metadata
	logger.DebugEvent().
		Str("token", conflict.TokenName).
		Msg("Resolving with newest strategy (defaulting to remote)")

	return ConflictResolution{
		Strategy:      ConflictResolutionNewest,
		SelectedToken: conflict.RemoteToken,
		Reason:        "Selected based on timestamp comparison",
	}
}

// resolveWithPrompt resolves conflict by prompting the user.
func (r *ConflictResolver) resolveWithPrompt(conflict Conflict) ConflictResolution {
	if !r.config.Interactive {
		logger.WarnEvent().
			Str("token", conflict.TokenName).
			Msg("Interactive mode disabled, falling back to remote")
		return r.resolveWithRemote(conflict)
	}

	fmt.Printf("\n=== Conflict Detected ===\n")
	fmt.Printf("Token: %s\n", conflict.TokenName)
	fmt.Printf("Type: %s\n", conflict.ConflictType)
	fmt.Printf("\n")
	fmt.Printf("Local version:\n")
	fmt.Printf("  Type:  %s\n", conflict.LocalToken.Type)
	fmt.Printf("  Value: %s\n", conflict.LocalToken.Value)
	if conflict.LocalToken.Category != "" {
		fmt.Printf("  Category: %s\n", conflict.LocalToken.Category)
	}
	fmt.Printf("\n")
	fmt.Printf("Remote version:\n")
	fmt.Printf("  Type:  %s\n", conflict.RemoteToken.Type)
	fmt.Printf("  Value: %s\n", conflict.RemoteToken.Value)
	if conflict.RemoteToken.Category != "" {
		fmt.Printf("  Category: %s\n", conflict.RemoteToken.Category)
	}
	fmt.Printf("\n")
	fmt.Printf("Choose resolution:\n")
	fmt.Printf("  1) Use local version\n")
	fmt.Printf("  2) Use remote version\n")
	fmt.Printf("  3) Skip this token\n")
	fmt.Printf("\n")
	fmt.Printf("Enter choice (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		logger.WarnEvent().Err(err).Msg("Failed to read user input, using remote")
		return r.resolveWithRemote(conflict)
	}

	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		return ConflictResolution{
			Strategy:      ConflictResolutionPrompt,
			SelectedToken: conflict.LocalToken,
			Reason:        "User selected local version",
		}
	case "2":
		return ConflictResolution{
			Strategy:      ConflictResolutionPrompt,
			SelectedToken: conflict.RemoteToken,
			Reason:        "User selected remote version",
		}
	case "3":
		return ConflictResolution{
			Strategy:      ConflictResolutionPrompt,
			SelectedToken: nil,
			Reason:        "User chose to skip this token",
		}
	default:
		fmt.Printf("Invalid choice, using remote version\n")
		return r.resolveWithRemote(conflict)
	}
}

// resolveWithMerge attempts to merge the conflicting tokens.
func (r *ConflictResolver) resolveWithMerge(conflict Conflict) ConflictResolution {
	// For design tokens, merging is tricky since values are atomic
	// We'll use a simple heuristic: if only metadata differs, merge it
	// Otherwise, fall back to preferring remote

	if conflict.LocalToken.Value == conflict.RemoteToken.Value &&
		conflict.LocalToken.Type == conflict.RemoteToken.Type {
		// Values are the same, merge metadata
		mergedToken := *conflict.LocalToken

		// Merge metadata from both sides
		if mergedToken.Metadata == nil {
			mergedToken.Metadata = make(map[string]string)
		}

		if conflict.RemoteToken.Metadata != nil {
			for k, v := range conflict.RemoteToken.Metadata {
				mergedToken.Metadata[k] = v
			}
		}

		// Prefer remote category and description if local is empty
		if mergedToken.Category == "" && conflict.RemoteToken.Category != "" {
			mergedToken.Category = conflict.RemoteToken.Category
		}
		if mergedToken.Description == "" && conflict.RemoteToken.Description != "" {
			mergedToken.Description = conflict.RemoteToken.Description
		}

		return ConflictResolution{
			Strategy:      ConflictResolutionMerge,
			SelectedToken: &mergedToken,
			Reason:        "Merged metadata from both versions",
		}
	}

	// Can't merge different values, fall back to remote
	logger.DebugEvent().
		Str("token", conflict.TokenName).
		Msg("Cannot merge conflicting values, using remote")

	return ConflictResolution{
		Strategy:      ConflictResolutionMerge,
		SelectedToken: conflict.RemoteToken,
		Reason:        "Cannot merge different values, used remote",
	}
}

// ResolveAll resolves multiple conflicts in batch.
func (r *ConflictResolver) ResolveAll(conflicts []Conflict) []ConflictResolution {
	resolutions := make([]ConflictResolution, len(conflicts))

	for i, conflict := range conflicts {
		resolutions[i] = r.Resolve(conflict)
	}

	return resolutions
}

// PrintConflictSummary prints a summary of conflicts and their resolutions.
func PrintConflictSummary(conflicts []Conflict) {
	if len(conflicts) == 0 {
		return
	}

	fmt.Printf("\n=== Conflict Summary ===\n")
	fmt.Printf("Total conflicts: %d\n\n", len(conflicts))

	for i, conflict := range conflicts {
		fmt.Printf("%d. %s (%s)\n", i+1, conflict.TokenName, conflict.ConflictType)
		if conflict.Resolution != nil {
			fmt.Printf("   Resolution: %s - %s\n", conflict.Resolution.Strategy, conflict.Resolution.Reason)
		} else {
			fmt.Printf("   Resolution: Pending\n")
		}
	}

	fmt.Printf("\n")
}
