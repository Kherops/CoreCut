# CoreCut

**Measure and prove performance gains with statistical rigor.**

CoreCut is a portable performance measurement tool that compares baseline vs optimized scenarios on the same machine, calculates robust gain percentages, and aggregates results across multiple machines.

## Key Features

- **Machine-agnostic**: Never compares raw times across machines. Always compares A vs B on the same machine, then aggregates ratios.
- **Process-only optimization**: No BIOS/kernel/hardware tuning required. Measures gains from code/process improvements.
- **Robust statistics**: Uses median-based calculations, warmup runs, A/B alternation, and confidence scoring.
- **eBPF insights**: When available, shows WHERE the time went (runqueue latency, off-CPU time, I/O latency, syscalls).
- **Beautiful reports**: Generates HTML dashboards suitable for stakeholder presentations.

## Installation

### Prerequisites

- Go 1.21+
- (Optional) bpftrace or bcc-tools for eBPF metrics
- (Optional) Root/sudo for eBPF collection

### Build

```bash
make build
# or
go build -o processgain .
```

### Install system-wide

```bash
sudo make install
```

## Quick Start

### 1. Create your scenarios

**baseline.sh** - Your current implementation:
```bash
#!/bin/bash
# Your baseline process
./my-app --mode=sequential
```

**optimized.sh** - Your optimized implementation:
```bash
#!/bin/bash
# Your optimized process
./my-app --mode=parallel --batch-size=100
```

### 2. Run the benchmark

```bash
corecut run \
  --baseline ./baseline.sh \
  --optimized ./optimized.sh \
  --runs 9 \
  --warmup 1 \
  --alternate
```

### 3. View results

Open `reports/report_<machine>_<timestamp>.html` in your browser.

## Usage

### Run Command

```bash
corecut run [flags]

Flags:
  -b, --baseline string     Path to baseline scenario script (required)
  -o, --optimized string    Path to optimized scenario script (required)
  -r, --runs int            Number of measured runs per scenario (default 9)
  -w, --warmup int          Number of warmup runs (default 1)
  -a, --alternate           Alternate A/B/A/B execution (default true)
      --cooldown-ms int     Cooldown between runs in milliseconds (default 500)
  -t, --timeout int         Timeout per run in seconds (default 300)
  -m, --mode string         Measurement mode: duration, throughput (default "duration")
      --output string       Output directory for reports (default "./reports")
      --tag string          Tag for this run (e.g., commit hash)
      --machine string      Machine name (auto-detected if empty)
      --no-ebpf             Disable eBPF collection
      --env-file string     Environment file to source before runs
```

### Aggregate Command

Combine results from multiple machines:

```bash
corecut aggregate ./reports/
```

This reads all `report_*.json` files and generates:
- `aggregate.json` - Combined data
- `aggregate.html` - Dashboard showing median gain across all machines

### Check Dependencies

```bash
corecut check-deps
```

## Measurement Modes

### Duration Mode (default)

Measures total execution time of each scenario.

```bash
corecut run --mode duration --baseline ./slow.sh --optimized ./fast.sh
```

### Throughput Mode

For scenarios that output a throughput value. Your script should print:
```
THROUGHPUT: 1234.56
```

```bash
corecut run --mode throughput --baseline ./baseline.py --optimized ./optimized.py
```

## eBPF Metrics

When running as root with bpftrace or bcc-tools installed, CoreCut collects:

| Metric | Description | Tool |
|--------|-------------|------|
| Runqueue Latency | Time waiting in CPU scheduler | runqlat |
| Off-CPU Time | Time blocked (I/O, locks, sleep) | offcputime |
| I/O Latency | Block device I/O latency | biolatency |
| Syscall Stats | Top syscalls by count/latency | syscount |

These metrics help explain **where** the performance gain comes from.

### Running with eBPF

```bash
sudo corecut run --baseline ./a.sh --optimized ./b.sh
```

### Running without eBPF

```bash
corecut run --baseline ./a.sh --optimized ./b.sh --no-ebpf
```

## Statistics

CoreCut uses robust statistical methods:

- **Median**: Primary metric (resistant to outliers)
- **Coefficient of Variation (CV)**: Measures run-to-run stability
- **P10/P90**: Shows distribution spread
- **Pairwise gain calculation**: When alternating, computes gain for each A/B pair
- **Overlap detection**: Determines if distributions are separable

### Gain Calculation

```
gain% = (median(baseline) - median(optimized)) / median(baseline) × 100
```

### Conclusiveness

A result is marked **conclusive** when:
- CV < 15% for both scenarios
- Distribution overlap < 30%
- Gain direction is consistent (P10 and P90 have same sign)

## Multi-Machine Aggregation

CoreCut is designed for comparing results across different machines:

1. Run benchmarks on each machine
2. Collect all `report_*.json` files in one folder
3. Run `corecut aggregate ./reports/`

**Important**: Only gain percentages are aggregated. Raw times are never compared across machines because hardware differences make such comparisons meaningless.

## Examples

See the `examples/` directory:

- `baseline.sh` / `optimized.sh` - Shell script examples (duration mode)
- `throughput_baseline.py` / `throughput_optimized.py` - Python examples (throughput mode)

Run the examples:

```bash
make run-example      # Duration mode
make run-throughput   # Throughput mode
```

## Output Files

After running, you'll find in the output directory:

| File | Description |
|------|-------------|
| `report_<machine>_<timestamp>.json` | Raw data in JSON format |
| `report_<machine>_<timestamp>.html` | Visual HTML report |
| `aggregate.json` | Combined multi-machine data |
| `aggregate.html` | Multi-machine dashboard |

## Best Practices

1. **Use warmup runs**: At least 1-2 warmup runs to prime caches
2. **Use alternation**: A/B/A/B pattern reduces time-based drift effects
3. **Run enough iterations**: 9+ runs for statistical significance
4. **Same inputs**: Ensure baseline and optimized use identical data
5. **Quiet machine**: Minimize background processes during benchmarks
6. **Tag your runs**: Use `--tag` to track commits/versions

## License

MIT License - See LICENSE file
# CoreCut
