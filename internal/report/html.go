package report

import (
	"fmt"
	"html/template"
	"os"
	"strings"
)

const singleReportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ProcessGain Report - {{.Machine}}</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        .gain-positive { color: #10b981; }
        .gain-negative { color: #ef4444; }
        .card { @apply bg-white rounded-xl shadow-lg p-6; }
    </style>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <!-- Header -->
        <div class="text-center mb-8">
            <h1 class="text-4xl font-bold text-gray-800 mb-2">ProcessGain Report</h1>
            <p class="text-gray-600">Machine: <strong>{{.Machine}}</strong> | Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}</p>
            {{if .Tag}}<p class="text-gray-500">Tag: {{.Tag}}</p>{{end}}
        </div>

        <!-- Main Gain Card -->
        <div class="bg-white rounded-xl shadow-lg p-8 mb-8 text-center">
            <h2 class="text-2xl font-semibold text-gray-700 mb-4">Performance Gain</h2>
            <div class="text-6xl font-bold mb-4 {{if ge .Comparison.GainPercent 0.0}}gain-positive{{else}}gain-negative{{end}}">
                {{printf "%.2f" .Comparison.GainPercent}}%
            </div>
            <p class="text-gray-600 mb-2">
                Baseline median: <strong>{{printf "%.2f" .Baseline.Stats.Median}}ms</strong> → 
                Optimized median: <strong>{{printf "%.2f" .Optimized.Stats.Median}}ms</strong>
            </p>
            <p class="text-gray-500">
                P10/P90 of gain: {{printf "%.2f" .Comparison.GainP10}}% / {{printf "%.2f" .Comparison.GainP90}}%
            </p>
            <div class="mt-4">
                {{if .Comparison.Conclusive}}
                <span class="inline-flex items-center px-4 py-2 rounded-full bg-green-100 text-green-800">
                    ✓ Conclusive Result
                </span>
                {{else}}
                <span class="inline-flex items-center px-4 py-2 rounded-full bg-yellow-100 text-yellow-800">
                    ⚠ Inconclusive (high variance or overlap)
                </span>
                {{end}}
            </div>
        </div>

        <!-- Statistics Comparison -->
        <div class="grid md:grid-cols-2 gap-6 mb-8">
            <div class="bg-white rounded-xl shadow-lg p-6">
                <h3 class="text-xl font-semibold text-gray-700 mb-4">Baseline Statistics</h3>
                <table class="w-full">
                    <tr class="border-b"><td class="py-2 text-gray-600">Median</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Baseline.Stats.Median}} ms</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">Mean</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Baseline.Stats.Mean}} ms</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">Std Dev</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Baseline.Stats.StdDev}} ms</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">CV</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Baseline.Stats.CV}}%</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">P10</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Baseline.Stats.P10}} ms</td></tr>
                    <tr><td class="py-2 text-gray-600">P90</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Baseline.Stats.P90}} ms</td></tr>
                </table>
            </div>
            <div class="bg-white rounded-xl shadow-lg p-6">
                <h3 class="text-xl font-semibold text-gray-700 mb-4">Optimized Statistics</h3>
                <table class="w-full">
                    <tr class="border-b"><td class="py-2 text-gray-600">Median</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Optimized.Stats.Median}} ms</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">Mean</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Optimized.Stats.Mean}} ms</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">Std Dev</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Optimized.Stats.StdDev}} ms</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">CV</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Optimized.Stats.CV}}%</td></tr>
                    <tr class="border-b"><td class="py-2 text-gray-600">P10</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Optimized.Stats.P10}} ms</td></tr>
                    <tr><td class="py-2 text-gray-600">P90</td><td class="py-2 font-mono text-right">{{printf "%.2f" .Optimized.Stats.P90}} ms</td></tr>
                </table>
            </div>
        </div>

        <!-- Duration Chart -->
        <div class="bg-white rounded-xl shadow-lg p-6 mb-8">
            <h3 class="text-xl font-semibold text-gray-700 mb-4">Run Durations</h3>
            <canvas id="durationsChart" height="100"></canvas>
        </div>

        <!-- eBPF Insights -->
        {{if .Baseline.Ebpf}}
        <div class="bg-white rounded-xl shadow-lg p-6 mb-8">
            <h3 class="text-xl font-semibold text-gray-700 mb-4">eBPF Insights (Where did the time go?)</h3>
            <div class="grid md:grid-cols-2 gap-4">
                <div>
                    <h4 class="font-semibold text-gray-600 mb-2">Runqueue Latency</h4>
                    <p class="text-sm text-gray-500">Time waiting in CPU scheduler queue</p>
                    <canvas id="runqlatChart" height="150"></canvas>
                </div>
                <div>
                    <h4 class="font-semibold text-gray-600 mb-2">I/O Latency</h4>
                    <p class="text-sm text-gray-500">Block I/O operation latency</p>
                    <canvas id="biolatChart" height="150"></canvas>
                </div>
            </div>
        </div>
        {{end}}

        <!-- Configuration -->
        <div class="bg-white rounded-xl shadow-lg p-6 mb-8">
            <h3 class="text-xl font-semibold text-gray-700 mb-4">Test Configuration</h3>
            <div class="grid md:grid-cols-2 gap-4 text-sm">
                <div><span class="text-gray-600">Baseline Script:</span> <code class="bg-gray-100 px-2 py-1 rounded">{{.Config.BaselineScript}}</code></div>
                <div><span class="text-gray-600">Optimized Script:</span> <code class="bg-gray-100 px-2 py-1 rounded">{{.Config.OptimizedScript}}</code></div>
                <div><span class="text-gray-600">Mode:</span> {{.Config.Mode}}</div>
                <div><span class="text-gray-600">Warmup Runs:</span> {{.Config.WarmupRuns}}</div>
                <div><span class="text-gray-600">Measured Runs:</span> {{.Config.MeasuredRuns}}</div>
                <div><span class="text-gray-600">Alternating:</span> {{.Config.Alternate}}</div>
                <div><span class="text-gray-600">Cooldown:</span> {{.Config.CooldownMs}}ms</div>
                <div><span class="text-gray-600">Timeout:</span> {{.Config.Timeout}}s</div>
            </div>
        </div>

        <!-- Footer -->
        <div class="text-center text-gray-500 text-sm">
            <p>Generated by ProcessGain | Comparison is relative (A vs B on same machine)</p>
            <p>Raw times should NEVER be compared across different machines</p>
        </div>
    </div>

    <script>
        // Duration chart
        const baselineDurations = [{{range .Baseline.Runs}}{{.DurationMs}},{{end}}];
        const optimizedDurations = [{{range .Optimized.Runs}}{{.DurationMs}},{{end}}];
        const labels = baselineDurations.map((_, i) => 'Run ' + (i + 1));

        new Chart(document.getElementById('durationsChart'), {
            type: 'line',
            data: {
                labels: labels,
                datasets: [
                    {
                        label: 'Baseline',
                        data: baselineDurations,
                        borderColor: '#6366f1',
                        backgroundColor: 'rgba(99, 102, 241, 0.1)',
                        tension: 0.1
                    },
                    {
                        label: 'Optimized',
                        data: optimizedDurations,
                        borderColor: '#10b981',
                        backgroundColor: 'rgba(16, 185, 129, 0.1)',
                        tension: 0.1
                    }
                ]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: { position: 'top' }
                },
                scales: {
                    y: {
                        beginAtZero: false,
                        title: { display: true, text: 'Duration (ms)' }
                    }
                }
            }
        });
    </script>
</body>
</html>`

const aggregateReportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ProcessGain - Aggregate Report</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        .gain-positive { color: #10b981; }
        .gain-negative { color: #ef4444; }
    </style>
</head>
<body class="bg-gray-100 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <!-- Header -->
        <div class="text-center mb-8">
            <h1 class="text-4xl font-bold text-gray-800 mb-2">ProcessGain - Aggregate Report</h1>
            <p class="text-gray-600">{{.MachineCount}} machines | Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05 UTC"}}</p>
        </div>

        <!-- Global Gain Card -->
        <div class="bg-white rounded-xl shadow-lg p-8 mb-8 text-center">
            <h2 class="text-2xl font-semibold text-gray-700 mb-4">Global Performance Gain</h2>
            <div class="text-6xl font-bold mb-4 {{if ge .AggregateStats.MedianGain 0.0}}gain-positive{{else}}gain-negative{{end}}">
                {{printf "%.2f" .AggregateStats.MedianGain}}%
            </div>
            <p class="text-gray-600 mb-2">Median gain across {{.MachineCount}} machines</p>
            <p class="text-gray-500">
                P10/P90: {{printf "%.2f" .AggregateStats.P10Gain}}% / {{printf "%.2f" .AggregateStats.P90Gain}}%
            </p>
            <p class="text-gray-500">
                Range: {{printf "%.2f" .AggregateStats.MinGain}}% to {{printf "%.2f" .AggregateStats.MaxGain}}%
            </p>
        </div>

        <!-- Gains Distribution Chart -->
        <div class="bg-white rounded-xl shadow-lg p-6 mb-8">
            <h3 class="text-xl font-semibold text-gray-700 mb-4">Gain Distribution Across Machines</h3>
            <canvas id="gainsChart" height="80"></canvas>
        </div>

        <!-- Per-Machine Table -->
        <div class="bg-white rounded-xl shadow-lg p-6 mb-8">
            <h3 class="text-xl font-semibold text-gray-700 mb-4">Per-Machine Results</h3>
            <div class="overflow-x-auto">
                <table class="w-full text-sm">
                    <thead>
                        <tr class="border-b-2 border-gray-200">
                            <th class="py-3 px-4 text-left">Machine</th>
                            <th class="py-3 px-4 text-right">Gain %</th>
                            <th class="py-3 px-4 text-right">Baseline (ms)</th>
                            <th class="py-3 px-4 text-right">Optimized (ms)</th>
                            <th class="py-3 px-4 text-center">Verdict</th>
                            <th class="py-3 px-4 text-left">Tag</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Reports}}
                        <tr class="border-b hover:bg-gray-50">
                            <td class="py-3 px-4 font-medium">{{.Machine}}</td>
                            <td class="py-3 px-4 text-right font-mono {{if ge .Comparison.GainPercent 0.0}}gain-positive{{else}}gain-negative{{end}}">
                                {{printf "%.2f" .Comparison.GainPercent}}%
                            </td>
                            <td class="py-3 px-4 text-right font-mono">{{printf "%.2f" .Baseline.Stats.Median}}</td>
                            <td class="py-3 px-4 text-right font-mono">{{printf "%.2f" .Optimized.Stats.Median}}</td>
                            <td class="py-3 px-4 text-center">
                                {{if .Comparison.Conclusive}}
                                <span class="text-green-600">✓</span>
                                {{else}}
                                <span class="text-yellow-600">?</span>
                                {{end}}
                            </td>
                            <td class="py-3 px-4 text-gray-500">{{.Tag}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Important Note -->
        <div class="bg-blue-50 border-l-4 border-blue-400 p-4 mb-8">
            <div class="flex">
                <div class="ml-3">
                    <p class="text-sm text-blue-700">
                        <strong>Note:</strong> All comparisons are relative (A vs B on the same machine). 
                        Raw execution times are NEVER compared across machines. 
                        Only gain percentages are aggregated.
                    </p>
                </div>
            </div>
        </div>

        <!-- Footer -->
        <div class="text-center text-gray-500 text-sm">
            <p>Generated by ProcessGain</p>
        </div>
    </div>

    <script>
        const gains = [{{range .Reports}}{{.Comparison.GainPercent}},{{end}}];
        const machines = [{{range .Reports}}"{{.Machine}}",{{end}}];

        new Chart(document.getElementById('gainsChart'), {
            type: 'bar',
            data: {
                labels: machines,
                datasets: [{
                    label: 'Gain %',
                    data: gains,
                    backgroundColor: gains.map(g => g >= 0 ? 'rgba(16, 185, 129, 0.7)' : 'rgba(239, 68, 68, 0.7)'),
                    borderColor: gains.map(g => g >= 0 ? '#10b981' : '#ef4444'),
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                plugins: {
                    legend: { display: false }
                },
                scales: {
                    y: {
                        title: { display: true, text: 'Gain (%)' }
                    }
                }
            }
        });
    </script>
</body>
</html>`

func GenerateHTML(r Report, outputPath string) error {
	tmpl, err := template.New("report").Parse(singleReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, r)
}

func GenerateAggregateHTML(r AggregateReport, outputPath string) error {
	// Escape machine names for JavaScript
	tmpl, err := template.New("aggregate").Funcs(template.FuncMap{
		"js": func(s string) template.JS {
			return template.JS(strings.ReplaceAll(s, `"`, `\"`))
		},
	}).Parse(aggregateReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, r)
}
