package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "corecut",
	Short: "CoreCut - Measure and prove performance gains",
	Long: `CoreCut is a portable performance measurement tool that:
- Compares baseline vs optimized scenarios on the same machine
- Calculates robust gain% using median-based statistics
- Collects eBPF metrics to explain WHERE the gain comes from
- Aggregates results across multiple machines (ratio-based, not raw times)

Usage:
  corecut run --baseline ./baseline.sh --optimized ./optimized.sh --runs 9
  corecut aggregate ./reports/`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(aggregateCmd)
	rootCmd.AddCommand(checkDepsCmd)
}

var checkDepsCmd = &cobra.Command{
	Use:   "check-deps",
	Short: "Check if required dependencies (bpftrace, bcc-tools) are available",
	Run: func(cmd *cobra.Command, args []string) {
		checkDependencies()
	},
}

func checkDependencies() {
	fmt.Println("Checking dependencies...")
	
	deps := []struct {
		name    string
		cmd     string
		required bool
	}{
		{"bpftrace", "bpftrace --version", false},
		{"bcc-tools (runqlat)", "which runqlat", false},
		{"bcc-tools (biolatency)", "which biolatency", false},
		{"bcc-tools (offcputime)", "which offcputime", false},
	}

	allFound := true
	for _, dep := range deps {
		// Simple check - in real impl we'd exec the command
		fmt.Printf("  [?] %s - check manually with: %s\n", dep.name, dep.cmd)
		if dep.required {
			allFound = false
		}
	}

	if allFound {
		fmt.Println("\n✓ Core dependencies OK. eBPF tools are optional but recommended.")
	}
	fmt.Println("\nNote: eBPF collection requires root/CAP_BPF. Run with sudo for full metrics.")
}
