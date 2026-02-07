package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/processgain/internal/ebpf"
	"github.com/processgain/internal/executor"
	"github.com/processgain/internal/report"
	"github.com/processgain/internal/stats"
	"github.com/spf13/cobra"
)

var (
	baselineScript  string
	optimizedScript string
	warmupRuns      int
	runs            int
	alternate       bool
	cooldownMs      int
	timeout         int
	envFile         string
	tag             string
	mode            string
	outputDir       string
	noEbpf          bool
	machineName     string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run A/B performance comparison between baseline and optimized scenarios",
	Long: `Execute baseline and optimized scenarios with proper warmup, alternation,
and statistical analysis. Collects eBPF metrics when available.

Example:
  processgain run --baseline ./baseline.sh --optimized ./optimized.sh --runs 9 --warmup 1 --alternate`,
	RunE: runBenchmark,
}

func init() {
	runCmd.Flags().StringVarP(&baselineScript, "baseline", "b", "", "Path to baseline scenario script (required)")
	runCmd.Flags().StringVarP(&optimizedScript, "optimized", "o", "", "Path to optimized scenario script (required)")
	runCmd.Flags().IntVarP(&warmupRuns, "warmup", "w", 1, "Number of warmup runs (discarded)")
	runCmd.Flags().IntVarP(&runs, "runs", "r", 9, "Number of measured runs per scenario")
	runCmd.Flags().BoolVarP(&alternate, "alternate", "a", true, "Alternate A/B/A/B execution (recommended)")
	runCmd.Flags().IntVar(&cooldownMs, "cooldown-ms", 500, "Cooldown between runs in milliseconds")
	runCmd.Flags().IntVarP(&timeout, "timeout", "t", 300, "Timeout per run in seconds")
	runCmd.Flags().StringVar(&envFile, "env-file", "", "Environment file to source before runs")
	runCmd.Flags().StringVar(&tag, "tag", "", "Tag for this run (e.g., commit hash, branch)")
	runCmd.Flags().StringVarP(&mode, "mode", "m", "duration", "Measurement mode: duration, throughput")
	runCmd.Flags().StringVar(&outputDir, "output", "./reports", "Output directory for reports")
	runCmd.Flags().BoolVar(&noEbpf, "no-ebpf", false, "Disable eBPF collection")
	runCmd.Flags().StringVar(&machineName, "machine", "", "Machine name (auto-detected if empty)")

	runCmd.MarkFlagRequired("baseline")
	runCmd.MarkFlagRequired("optimized")
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	bold := color.New(color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	bold.Println("\n╔══════════════════════════════════════════════════════════════╗")
	bold.Println("║              ProcessGain - Performance Measurement           ║")
	bold.Println("╚══════════════════════════════════════════════════════════════╝")

	// Validate scripts exist
	if _, err := os.Stat(baselineScript); os.IsNotExist(err) {
		return fmt.Errorf("baseline script not found: %s", baselineScript)
	}
	if _, err := os.Stat(optimizedScript); os.IsNotExist(err) {
		return fmt.Errorf("optimized script not found: %s", optimizedScript)
	}

	// Get machine info
	machine := machineName
	if machine == "" {
		hostname, _ := os.Hostname()
		machine = hostname
	}

	fmt.Printf("\n📊 Configuration:\n")
	fmt.Printf("   Machine:    %s\n", machine)
	fmt.Printf("   Baseline:   %s\n", baselineScript)
	fmt.Printf("   Optimized:  %s\n", optimizedScript)
	fmt.Printf("   Mode:       %s\n", mode)
	fmt.Printf("   Warmup:     %d runs\n", warmupRuns)
	fmt.Printf("   Measured:   %d runs per scenario\n", runs)
	fmt.Printf("   Alternate:  %v\n", alternate)
	fmt.Printf("   Cooldown:   %d ms\n", cooldownMs)
	if tag != "" {
		fmt.Printf("   Tag:        %s\n", tag)
	}

	// Check eBPF availability
	ebpfCollector := ebpf.NewCollector()
	ebpfAvailable := !noEbpf && ebpfCollector.IsAvailable()
	if ebpfAvailable {
		green.Println("\n✓ eBPF collection enabled (running as root)")
	} else {
		yellow.Println("\n⚠ eBPF collection disabled (no root or --no-ebpf)")
	}

	exec := executor.New(timeout, cooldownMs, envFile)

	// Warmup phase
	if warmupRuns > 0 {
		fmt.Printf("\n🔥 Warmup phase (%d runs each)...\n", warmupRuns)
		for i := 0; i < warmupRuns; i++ {
			fmt.Printf("   Warmup baseline %d/%d...", i+1, warmupRuns)
			_, err := exec.Run(baselineScript, mode)
			if err != nil {
				red.Printf(" FAILED: %v\n", err)
			} else {
				fmt.Println(" done")
			}

			fmt.Printf("   Warmup optimized %d/%d...", i+1, warmupRuns)
			_, err = exec.Run(optimizedScript, mode)
			if err != nil {
				red.Printf(" FAILED: %v\n", err)
			} else {
				fmt.Println(" done")
			}
		}
	}

	// Measurement phase
	fmt.Printf("\n📏 Measurement phase (%d runs each)...\n", runs)

	var baselineResults []executor.RunResult
	var optimizedResults []executor.RunResult
	var baselineEbpf []ebpf.Metrics
	var optimizedEbpf []ebpf.Metrics

	if alternate {
		// A/B/A/B alternation
		for i := 0; i < runs; i++ {
			// Baseline run
			fmt.Printf("   [%d/%d] Baseline...", i+1, runs)
			var ebpfMetrics *ebpf.Metrics
			if ebpfAvailable {
				ebpfCollector.Start()
			}
			result, err := exec.Run(baselineScript, mode)
			if ebpfAvailable {
				ebpfMetrics = ebpfCollector.Stop()
				baselineEbpf = append(baselineEbpf, *ebpfMetrics)
			}
			if err != nil {
				red.Printf(" FAILED: %v\n", err)
				result.Error = err.Error()
			} else {
				fmt.Printf(" %.2fms\n", result.DurationMs)
			}
			baselineResults = append(baselineResults, result)

			// Optimized run
			fmt.Printf("   [%d/%d] Optimized...", i+1, runs)
			if ebpfAvailable {
				ebpfCollector.Start()
			}
			result, err = exec.Run(optimizedScript, mode)
			if ebpfAvailable {
				ebpfMetrics = ebpfCollector.Stop()
				optimizedEbpf = append(optimizedEbpf, *ebpfMetrics)
			}
			if err != nil {
				red.Printf(" FAILED: %v\n", err)
				result.Error = err.Error()
			} else {
				fmt.Printf(" %.2fms\n", result.DurationMs)
			}
			optimizedResults = append(optimizedResults, result)
		}
	} else {
		// Sequential: all baseline then all optimized
		for i := 0; i < runs; i++ {
			fmt.Printf("   Baseline [%d/%d]...", i+1, runs)
			var ebpfMetrics *ebpf.Metrics
			if ebpfAvailable {
				ebpfCollector.Start()
			}
			result, err := exec.Run(baselineScript, mode)
			if ebpfAvailable {
				ebpfMetrics = ebpfCollector.Stop()
				baselineEbpf = append(baselineEbpf, *ebpfMetrics)
			}
			if err != nil {
				red.Printf(" FAILED: %v\n", err)
				result.Error = err.Error()
			} else {
				fmt.Printf(" %.2fms\n", result.DurationMs)
			}
			baselineResults = append(baselineResults, result)
		}
		for i := 0; i < runs; i++ {
			fmt.Printf("   Optimized [%d/%d]...", i+1, runs)
			var ebpfMetrics *ebpf.Metrics
			if ebpfAvailable {
				ebpfCollector.Start()
			}
			result, err := exec.Run(optimizedScript, mode)
			if ebpfAvailable {
				ebpfMetrics = ebpfCollector.Stop()
				optimizedEbpf = append(optimizedEbpf, *ebpfMetrics)
			}
			if err != nil {
				red.Printf(" FAILED: %v\n", err)
				result.Error = err.Error()
			} else {
				fmt.Printf(" %.2fms\n", result.DurationMs)
			}
			optimizedResults = append(optimizedResults, result)
		}
	}

	// Calculate statistics
	fmt.Println("\n📈 Calculating statistics...")

	baselineDurations := extractDurations(baselineResults)
	optimizedDurations := extractDurations(optimizedResults)

	baselineStats := stats.Calculate(baselineDurations)
	optimizedStats := stats.Calculate(optimizedDurations)
	comparison := stats.Compare(baselineDurations, optimizedDurations, alternate)

	// Display results
	fmt.Println("\n" + bold.Sprint("═══════════════════════════════════════════════════════════════"))
	fmt.Println(bold.Sprint("                         RESULTS"))
	fmt.Println(bold.Sprint("═══════════════════════════════════════════════════════════════"))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Metric", "Baseline", "Optimized"})
	table.SetBorder(false)
	table.Append([]string{"Median (ms)", fmt.Sprintf("%.2f", baselineStats.Median), fmt.Sprintf("%.2f", optimizedStats.Median)})
	table.Append([]string{"Mean (ms)", fmt.Sprintf("%.2f", baselineStats.Mean), fmt.Sprintf("%.2f", optimizedStats.Mean)})
	table.Append([]string{"Std Dev (ms)", fmt.Sprintf("%.2f", baselineStats.StdDev), fmt.Sprintf("%.2f", optimizedStats.StdDev)})
	table.Append([]string{"CV (%)", fmt.Sprintf("%.2f", baselineStats.CV), fmt.Sprintf("%.2f", optimizedStats.CV)})
	table.Append([]string{"P10 (ms)", fmt.Sprintf("%.2f", baselineStats.P10), fmt.Sprintf("%.2f", optimizedStats.P10)})
	table.Append([]string{"P90 (ms)", fmt.Sprintf("%.2f", baselineStats.P90), fmt.Sprintf("%.2f", optimizedStats.P90)})
	table.Render()

	fmt.Println()

	// Gain display
	gainColor := green
	if comparison.GainPercent < 0 {
		gainColor = red
	}

	fmt.Printf("🎯 ")
	gainColor.Printf("GAIN: %.2f%%", comparison.GainPercent)
	fmt.Printf(" (median baseline %.2fms → optimized %.2fms)\n", baselineStats.Median, optimizedStats.Median)
	fmt.Printf("   P10/P90 of gain: %.2f%% / %.2f%%\n", comparison.GainP10, comparison.GainP90)

	if comparison.Conclusive {
		green.Println("   ✓ Result is CONCLUSIVE (low overlap, stable measurements)")
	} else {
		yellow.Println("   ⚠ Result is INCONCLUSIVE (high variance or overlap)")
	}

	// eBPF summary if available
	if ebpfAvailable && len(baselineEbpf) > 0 {
		fmt.Println("\n" + bold.Sprint("eBPF Insights:"))
		displayEbpfComparison(baselineEbpf, optimizedEbpf)
	}

	// Generate report
	fmt.Println("\n📄 Generating reports...")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	reportData := report.Report{
		Version:     "1.0",
		GeneratedAt: time.Now().UTC(),
		Machine:     machine,
		Tag:         tag,
		Config: report.Config{
			BaselineScript:  baselineScript,
			OptimizedScript: optimizedScript,
			Mode:            mode,
			WarmupRuns:      warmupRuns,
			MeasuredRuns:    runs,
			Alternate:       alternate,
			CooldownMs:      cooldownMs,
			Timeout:         timeout,
		},
		Baseline: report.ScenarioResult{
			Runs:  baselineResults,
			Stats: baselineStats,
			Ebpf:  baselineEbpf,
		},
		Optimized: report.ScenarioResult{
			Runs:  optimizedResults,
			Stats: optimizedStats,
			Ebpf:  optimizedEbpf,
		},
		Comparison: comparison,
	}

	// Write JSON report
	jsonPath := filepath.Join(outputDir, fmt.Sprintf("report_%s_%s.json", machine, time.Now().Format("20060102_150405")))
	jsonData, _ := json.MarshalIndent(reportData, "", "  ")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}
	fmt.Printf("   ✓ JSON: %s\n", jsonPath)

	// Write HTML report
	htmlPath := filepath.Join(outputDir, fmt.Sprintf("report_%s_%s.html", machine, time.Now().Format("20060102_150405")))
	if err := report.GenerateHTML(reportData, htmlPath); err != nil {
		return fmt.Errorf("failed to write HTML report: %w", err)
	}
	fmt.Printf("   ✓ HTML: %s\n", htmlPath)

	fmt.Println("\n" + green.Sprint("✓ Benchmark complete!"))

	return nil
}

func extractDurations(results []executor.RunResult) []float64 {
	durations := make([]float64, 0, len(results))
	for _, r := range results {
		if r.Error == "" {
			durations = append(durations, r.DurationMs)
		}
	}
	return durations
}

func displayEbpfComparison(baseline, optimized []ebpf.Metrics) {
	// Aggregate eBPF metrics
	baselineAgg := ebpf.Aggregate(baseline)
	optimizedAgg := ebpf.Aggregate(optimized)

	if baselineAgg.RunqueueLatencyUs > 0 || optimizedAgg.RunqueueLatencyUs > 0 {
		fmt.Printf("   Runqueue latency (avg): %.2fμs → %.2fμs\n",
			baselineAgg.RunqueueLatencyUs, optimizedAgg.RunqueueLatencyUs)
	}
	if baselineAgg.OffCpuTimeMs > 0 || optimizedAgg.OffCpuTimeMs > 0 {
		fmt.Printf("   Off-CPU time (total): %.2fms → %.2fms\n",
			baselineAgg.OffCpuTimeMs, optimizedAgg.OffCpuTimeMs)
	}
	if baselineAgg.IoLatencyUs > 0 || optimizedAgg.IoLatencyUs > 0 {
		fmt.Printf("   I/O latency (avg): %.2fμs → %.2fμs\n",
			baselineAgg.IoLatencyUs, optimizedAgg.IoLatencyUs)
	}
	if len(baselineAgg.TopSyscalls) > 0 {
		fmt.Println("   Top syscalls (baseline):", formatSyscalls(baselineAgg.TopSyscalls))
	}
	if len(optimizedAgg.TopSyscalls) > 0 {
		fmt.Println("   Top syscalls (optimized):", formatSyscalls(optimizedAgg.TopSyscalls))
	}
}

func formatSyscalls(syscalls map[string]int64) string {
	result := ""
	count := 0
	for name, cnt := range syscalls {
		if count > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%s:%d", name, cnt)
		count++
		if count >= 5 {
			break
		}
	}
	return result
}
