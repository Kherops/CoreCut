package report

import (
	"time"

	"github.com/processgain/internal/ebpf"
	"github.com/processgain/internal/executor"
	"github.com/processgain/internal/stats"
)

type Report struct {
	Version     string         `json:"version"`
	GeneratedAt time.Time      `json:"generated_at"`
	Machine     string         `json:"machine"`
	Tag         string         `json:"tag,omitempty"`
	Config      Config         `json:"config"`
	Baseline    ScenarioResult `json:"baseline"`
	Optimized   ScenarioResult `json:"optimized"`
	Comparison  stats.Comparison `json:"comparison"`
}

type Config struct {
	BaselineScript  string `json:"baseline_script"`
	OptimizedScript string `json:"optimized_script"`
	Mode            string `json:"mode"`
	WarmupRuns      int    `json:"warmup_runs"`
	MeasuredRuns    int    `json:"measured_runs"`
	Alternate       bool   `json:"alternate"`
	CooldownMs      int    `json:"cooldown_ms"`
	Timeout         int    `json:"timeout"`
}

type ScenarioResult struct {
	Runs  []executor.RunResult `json:"runs"`
	Stats stats.Stats          `json:"stats"`
	Ebpf  []ebpf.Metrics       `json:"ebpf,omitempty"`
}

type AggregateReport struct {
	Version        string         `json:"version"`
	GeneratedAt    time.Time      `json:"generated_at"`
	MachineCount   int            `json:"machine_count"`
	Reports        []Report       `json:"reports"`
	AggregateStats AggregateStats `json:"aggregate_stats"`
}

type AggregateStats struct {
	MedianGain float64 `json:"median_gain"`
	MeanGain   float64 `json:"mean_gain"`
	StdDevGain float64 `json:"std_dev_gain"`
	P10Gain    float64 `json:"p10_gain"`
	P90Gain    float64 `json:"p90_gain"`
	MinGain    float64 `json:"min_gain"`
	MaxGain    float64 `json:"max_gain"`
}
