package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// These will be set during build using ldflags
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
	builtBy   = "manual"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Display version information about AINative Code including:
- Version number
- Git commit hash
- Build date
- Go version
- OS and architecture`,
	Aliases: []string{"v", "ver"},
	Run:     runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Version flags
	versionCmd.Flags().BoolP("short", "s", false, "print version number only")
	versionCmd.Flags().Bool("json", false, "output version info as JSON")
}

func runVersion(cmd *cobra.Command, args []string) {
	short, _ := cmd.Flags().GetBool("short")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if short {
		fmt.Println(version)
		return
	}

	if jsonOutput {
		fmt.Printf(`{
  "version": "%s",
  "commit": "%s",
  "buildDate": "%s",
  "builtBy": "%s",
  "goVersion": "%s",
  "platform": "%s/%s"
}
`, version, commit, buildDate, builtBy, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return
	}

	fmt.Printf("AINative Code v%s\n", version)
	fmt.Printf("Commit:     %s\n", commit)
	fmt.Printf("Built:      %s\n", buildDate)
	fmt.Printf("Built by:   %s\n", builtBy)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// GetVersion returns the version string
func GetVersion() string {
	return version
}

// GetCommit returns the git commit hash
func GetCommit() string {
	return commit
}

// GetBuildDate returns the build date
func GetBuildDate() string {
	return buildDate
}

// SetVersion sets the version (used by build scripts)
func SetVersion(v string) {
	version = v
}

// SetCommit sets the commit hash (used by build scripts)
func SetCommit(c string) {
	commit = c
}

// SetBuildDate sets the build date (used by build scripts)
func SetBuildDate(d string) {
	buildDate = d
}

// SetBuiltBy sets the builder (used by build scripts)
func SetBuiltBy(b string) {
	builtBy = b
}
