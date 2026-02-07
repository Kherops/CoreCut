package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/processgain/internal/report"
	"github.com/processgain/internal/stats"
	"github.com/spf13/cobra"
)

var aggregateOutputDir string

var aggregateCmd = &cobra.Command{
	Use:   "aggregate <reports-folder>",
	Short: "Aggregate results from multiple machines",
	Long: `Read report.json files from multiple machines and compute aggregate statistics.

This command:
- Collects gain% from each machine (never compares raw times across machines)
- Computes median, P10/P90 of gains across all machines
- Generates an aggregate HTML dashboard

Example:
  processgain aggregate ./reports/`,
	Args: cobra.ExactArgs(1),
	RunE: runAggregate,
}

func init() {
	aggregateCmd.Flags().StringVarP(&aggregateOutputDir, "output", "o", "", "Output directory (defaults to input folder)")
}

func runAggregate(cmd *cobra.Command, args []string) error {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen, color.Bold)

	reportsFolder := args[0]
	if aggregateOutputDir == "" {
		aggregateOutputDir = reportsFolder
	}

	bold.Println("\n╔══════════════════════════════════════════════════════════════╗")
	bold.Println("║           ProcessGain - Multi-Machine Aggregation            ║")
	bold.Println("╚══════════════════════════════════════════════════════════════╝")

	// Find all JSON reports
	var reports []report.Report
	var reportFiles []string

	err := filepath.Walk(reportsFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" && filepath.Base(path) != "aggregate.json" {
			reportFiles = append(reportFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan reports folder: %w", err)
	}

	if len(reportFiles) == 0 {
		return fmt.Errorf("no JSON reports found in %s", reportsFolder)
	}

	fmt.Printf("\n📂 Found %d report files\n", len(reportFiles))

	// Load all reports
	for _, path := range reportFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("   ⚠ Skipping %s: %v\n", path, err)
			continue
		}

		var r report.Report
		if err := json.Unmarshal(data, &r); err != nil {
			fmt.Printf("   ⚠ Skipping %s: invalid JSON\n", path)
			continue
		}

		reports = append(reports, r)
		fmt.Printf("   ✓ Loaded: %s (machine: %s, gain: %.2f%%)\n",
			filepath.Base(path), r.Machine, r.Comparison.GainPercent)
	}

	if len(reports) == 0 {
		return fmt.Errorf("no valid reports loaded")
	}

	// Extract gains from all machines
	var gains []float64
	for _, r := range reports {
		gains = append(gains, r.Comparison.GainPercent)
	}

	// Calculate aggregate statistics
	sort.Float64s(gains)
	aggStats := stats.Calculate(gains)

	fmt.Println("\n" + bold.Sprint("═══════════════════════════════════════════════════════════════"))
	fmt.Println(bold.Sprint("                    AGGREGATE RESULTS"))
	fmt.Println(bold.Sprint("═══════════════════════════════════════════════════════════════"))

	fmt.Printf("\n🎯 ")
	green.Printf("MEDIAN GAIN: %.2f%%\n", aggStats.Median)
	fmt.Printf("   P10/P90: %.2f%% / %.2f%%\n", aggStats.P10, aggStats.P90)
	fmt.Printf("   Mean: %.2f%%, StdDev: %.2f%%\n", aggStats.Mean, aggStats.StdDev)
	fmt.Printf("   Machines: %d\n", len(reports))

	// Per-machine table
	fmt.Println("\n" + bold.Sprint("Per-Machine Results:"))
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Machine", "Gain %", "Baseline (ms)", "Optimized (ms)", "Verdict", "Tag"})
	table.SetBorder(false)

	for _, r := range reports {
		verdict := "✓"
		if !r.Comparison.Conclusive {
			verdict = "?"
		}
		table.Append([]string{
			r.Machine,
			fmt.Sprintf("%.2f", r.Comparison.GainPercent),
			fmt.Sprintf("%.2f", r.Baseline.Stats.Median),
			fmt.Sprintf("%.2f", r.Optimized.Stats.Median),
			verdict,
			r.Tag,
		})
	}
	table.Render()

	// Generate aggregate report
	fmt.Println("\n📄 Generating aggregate reports...")

	aggReport := report.AggregateReport{
		Version:     "1.0",
		GeneratedAt: time.Now().UTC(),
		MachineCount: len(reports),
		Reports:     reports,
		AggregateStats: report.AggregateStats{
			MedianGain: aggStats.Median,
			MeanGain:   aggStats.Mean,
			StdDevGain: aggStats.StdDev,
			P10Gain:    aggStats.P10,
			P90Gain:    aggStats.P90,
			MinGain:    aggStats.Min,
			MaxGain:    aggStats.Max,
		},
	}

	// Write aggregate JSON
	jsonPath := filepath.Join(aggregateOutputDir, "aggregate.json")
	jsonData, _ := json.MarshalIndent(aggReport, "", "  ")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write aggregate JSON: %w", err)
	}
	fmt.Printf("   ✓ JSON: %s\n", jsonPath)

	// Write aggregate HTML
	htmlPath := filepath.Join(aggregateOutputDir, "aggregate.html")
	if err := report.GenerateAggregateHTML(aggReport, htmlPath); err != nil {
		return fmt.Errorf("failed to write aggregate HTML: %w", err)
	}
	fmt.Printf("   ✓ HTML: %s\n", htmlPath)

	fmt.Println("\n" + green.Sprint("✓ Aggregation complete!"))
	fmt.Println("\n💡 Note: Gains are compared per-machine (A vs B), then aggregated.")
	fmt.Println("   Raw times are NEVER compared across machines.")

	return nil
}
