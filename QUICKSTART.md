# CoreCut Quick Start

## 1. Build the CLI

```bash
cd /home/xzora/hackaton/processgain
go mod tidy
make build
```

## 2. Run Example Benchmark

```bash
# Make scripts executable
chmod +x examples/*.sh examples/*.py

# Run duration benchmark
./corecut run \
  --baseline ./examples/baseline.sh \
  --optimized ./examples/optimized.sh \
  --runs 5 \
  --warmup 1 \
  --alternate

# Or with sudo for eBPF metrics
sudo ./corecut run \
  --baseline ./examples/baseline.sh \
  --optimized ./examples/optimized.sh \
  --runs 5 \
  --warmup 1
```

## 3. View Results

Open the generated HTML report:
```bash
xdg-open reports/report_*.html
# or
firefox reports/report_*.html
```

## 4. Run Dashboard (Optional)

```bash
cd web
npm install
npm run dev
```

Then open http://localhost:3000 and load your `report_*.json` file.

## 5. Multi-Machine Aggregation

After running on multiple machines, collect all `report_*.json` files and:

```bash
./corecut aggregate ./reports/
xdg-open reports/aggregate.html
```

## Commands Reference

| Command | Description |
|---------|-------------|
| `corecut run` | Run A/B benchmark |
| `corecut aggregate <folder>` | Aggregate multi-machine results |
| `corecut check-deps` | Check eBPF dependencies |

## Key Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--baseline` | required | Baseline script path |
| `--optimized` | required | Optimized script path |
| `--runs` | 9 | Number of measured runs |
| `--warmup` | 1 | Warmup runs (discarded) |
| `--alternate` | true | A/B/A/B execution pattern |
| `--mode` | duration | `duration` or `throughput` |
| `--no-ebpf` | false | Disable eBPF collection |
| `--tag` | "" | Version/commit tag |
