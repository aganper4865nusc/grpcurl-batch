package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/example/grpcurl-batch/internal/config"
)

var cfg = config.DefaultConfig()

var rootCmd = &cobra.Command{
	Use:   "grpcurl-batch",
	Short: "Batch-execute gRPC calls from a YAML manifest",
	Long: `grpcurl-batch reads a YAML manifest describing one or more gRPC calls
and executes them concurrently with configurable retry and timeout controls.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Running manifest: %s\n", cfg.ManifestPath)
		// TODO: wire manifest.Load -> runner.New -> reporter.New
		return nil
	},
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.Manifest", "", "path to the YAML manifest file (required)")
	rootCmd.Flags().StringVarP(&cfg.OutputFormat, "output", "o", cfg.OutputFormat, "output format: text or json")
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", cfg.Concurrency, "maximum number of concurrent gRPC calls")
	rootCmd.Flags().DurationVarP(&cfg.Timeout, "timeout", "t", cfg.Timeout, "per-call timeout")
	rootCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", cfg.Verbose, "enable verbose logging")

	_ = rootCmd.MarkFlagRequired("manifest")
	_ = time.Second // ensure time import is used via config defaults
}
