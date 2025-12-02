package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/memodb-io/Acontext/acontext-cli/cmd"
	"github.com/memodb-io/Acontext/acontext-cli/internal/logo"
	"github.com/memodb-io/Acontext/acontext-cli/internal/telemetry"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type contextKey string

const startTimeKey contextKey = "start_time"

var version = "dev"

func main() {
	// Print logo on first run
	if len(os.Args) > 1 && os.Args[1] != "--help" && os.Args[1] != "-h" {
		fmt.Println(logo.Logo)
	}

	if cmdErr := rootCmd.Execute(); cmdErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
		executedCmd, _, _ := rootCmd.Find(os.Args[1:])
		if executedCmd == nil {
			executedCmd = rootCmd
		}
		trackCommandAndWait(executedCmd, os.Args[1:], cmdErr, false)
		os.Exit(1)
	}
}

// trackCommandAndWait tracks a command execution asynchronously and waits for completion
func trackCommandAndWait(cmd *cobra.Command, args []string, err error, success bool) {
	// Skip telemetry for dev version
	if version == "dev" {
		return
	}

	// Get start time from context and calculate duration
	var duration time.Duration
	if success {
		startTime, ok := cmd.Context().Value(startTimeKey).(time.Time)
		if !ok {
			startTime = time.Now()
		}
		duration = time.Since(startTime)
	}

	// Build command path, collect flags, and filter args
	commandPath := buildCommandPath(cmd)
	flags := collectFlags(cmd)
	filteredArgs := filterArgs(args)

	// Start async telemetry tracking and wait for completion
	// This ensures telemetry is sent even for blocking commands
	wg := telemetry.TrackCommandAsync(
		commandPath,
		filteredArgs,
		flags,
		success,
		err,
		duration,
		version,
	)

	// Wait for telemetry to complete (with timeout to avoid hanging forever)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait up to 5 seconds for telemetry to complete
	select {
	case <-done:
		// Telemetry completed successfully
	case <-time.After(5 * time.Second):
		// Timeout - don't block forever
	}
}

// buildCommandPath builds the full command path (e.g., "docker.up", "create")
func buildCommandPath(cmd *cobra.Command) string {
	var parts []string

	// Walk up the command tree
	current := cmd
	for current != nil && current.Use != "acontext" {
		parts = append([]string{current.Use}, parts...)
		current = current.Parent()
	}

	if len(parts) == 0 {
		return "root"
	}

	return strings.Join(parts, ".")
}

// collectFlags collects all flags that were set
func collectFlags(cmd *cobra.Command) map[string]string {
	flags := make(map[string]string)

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		flags[flag.Name] = flag.Value.String()
	})

	return flags
}

// filterArgs filters out help-related arguments
func filterArgs(args []string) []string {
	var filtered []string
	for _, arg := range args {
		if arg != "--help" && arg != "-h" && arg != "help" {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

var rootCmd = &cobra.Command{
	Use:   "acontext",
	Short: "Acontext CLI - Build context-aware AI applications",
	Long: `Acontext CLI is a command-line tool for quickly creating Acontext projects.
	
It helps you:
  - Create projects with templates for Python or TypeScript
  - Initialize Git repositories
  - Deploy local development environments with Docker

Get started by running: acontext create
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Store start time for telemetry
		ctx := context.WithValue(cmd.Context(), startTimeKey, time.Now())
		cmd.SetContext(ctx)
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// Track successful command execution
		// This is called after the command's Run/RunE completes successfully
		trackCommandAndWait(cmd, args, nil, true)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(logo.Logo)
		fmt.Println()
		fmt.Println("Welcome to Acontext CLI!")
		fmt.Println()
		fmt.Println("Quick Commands:")
		fmt.Println("  acontext create     Create a new project")
		fmt.Println("  acontext docker     Manage Docker services (up/down/status/logs/env)")
		fmt.Println("  acontext version    Show version information")
		fmt.Println("  acontext help       Show help information")
		fmt.Println()
		fmt.Println("Get started: acontext create")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(cmd.CreateCmd)
	rootCmd.AddCommand(cmd.DockerCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Acontext CLI version %s\n", version)
	},
}
