package ebpf

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type Metrics struct {
	RunqueueLatencyUs float64            `json:"runqueue_latency_us,omitempty"`
	RunqlatHistogram  map[string]int64   `json:"runqlat_histogram,omitempty"`
	OffCpuTimeMs      float64            `json:"offcpu_time_ms,omitempty"`
	OffCpuTopStacks   []StackTrace       `json:"offcpu_top_stacks,omitempty"`
	IoLatencyUs       float64            `json:"io_latency_us,omitempty"`
	BiolatHistogram   map[string]int64   `json:"biolat_histogram,omitempty"`
	TopSyscalls       map[string]int64   `json:"top_syscalls,omitempty"`
	SyscallLatencyUs  map[string]float64 `json:"syscall_latency_us,omitempty"`
}

type StackTrace struct {
	Stack   string  `json:"stack"`
	TimeMs  float64 `json:"time_ms"`
	Percent float64 `json:"percent"`
}

type AggregatedMetrics struct {
	RunqueueLatencyUs float64          `json:"runqueue_latency_us"`
	OffCpuTimeMs      float64          `json:"offcpu_time_ms"`
	IoLatencyUs       float64          `json:"io_latency_us"`
	TopSyscalls       map[string]int64 `json:"top_syscalls"`
}

type Collector struct {
	available   bool
	hasBpftrace bool
	hasBcc      bool
	mu          sync.Mutex
	running     bool
	stopChan    chan struct{}
	pid         int
	results     *Metrics
	cmds        []*exec.Cmd
}

func NewCollector() *Collector {
	c := &Collector{
		stopChan: make(chan struct{}),
	}
	c.checkAvailability()
	return c
}

func (c *Collector) checkAvailability() {
	// Check if running as root
	if os.Geteuid() != 0 {
		c.available = false
		return
	}

	// Check for bpftrace
	if _, err := exec.LookPath("bpftrace"); err == nil {
		c.hasBpftrace = true
		c.available = true
	}

	// Check for bcc tools
	for _, tool := range []string{"runqlat", "biolatency", "offcputime"} {
		if _, err := exec.LookPath(tool); err == nil {
			c.hasBcc = true
			c.available = true
			break
		}
		// Also check in /usr/share/bcc/tools/
		if _, err := os.Stat("/usr/share/bcc/tools/" + tool); err == nil {
			c.hasBcc = true
			c.available = true
			break
		}
	}
}

func (c *Collector) IsAvailable() bool {
	return c.available
}

func (c *Collector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.available || c.running {
		return
	}

	c.running = true
	c.stopChan = make(chan struct{})
	c.results = &Metrics{
		RunqlatHistogram: make(map[string]int64),
		BiolatHistogram:  make(map[string]int64),
		TopSyscalls:      make(map[string]int64),
		SyscallLatencyUs: make(map[string]float64),
	}
	c.cmds = nil

	// Start collection tools in background
	go c.collectRunqlat()
	go c.collectBiolatency()
}

func (c *Collector) Stop() *Metrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return &Metrics{}
	}

	close(c.stopChan)
	c.running = false

	// Stop all running commands
	for _, cmd := range c.cmds {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	return c.results
}

func (c *Collector) collectRunqlat() {
	var cmd *exec.Cmd
	var stdout bytes.Buffer

	// Try bcc runqlat first
	if path, err := exec.LookPath("runqlat"); err == nil {
		cmd = exec.Command(path, "-m", "1")
	} else if _, err := os.Stat("/usr/share/bcc/tools/runqlat"); err == nil {
		cmd = exec.Command("/usr/share/bcc/tools/runqlat", "-m", "1")
	} else if c.hasBpftrace {
		// Fallback to bpftrace one-liner
		script := `tracepoint:sched:sched_wakeup { @qtime[args->pid] = nsecs; }
tracepoint:sched:sched_switch { if (@qtime[args->next_pid]) { @usecs = hist((nsecs - @qtime[args->next_pid]) / 1000); delete(@qtime[args->next_pid]); } }
interval:s:1 { exit(); }`
		cmd = exec.Command("bpftrace", "-e", script)
	}

	if cmd == nil {
		return
	}

	cmd.Stdout = &stdout
	c.mu.Lock()
	c.cmds = append(c.cmds, cmd)
	c.mu.Unlock()

	cmd.Start()

	select {
	case <-c.stopChan:
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	cmd.Wait()

	// Parse output
	c.mu.Lock()
	c.results.RunqlatHistogram = parseHistogram(stdout.String())
	c.results.RunqueueLatencyUs = calculateAvgFromHistogram(c.results.RunqlatHistogram)
	c.mu.Unlock()
}

func (c *Collector) collectBiolatency() {
	var cmd *exec.Cmd
	var stdout bytes.Buffer

	if path, err := exec.LookPath("biolatency"); err == nil {
		cmd = exec.Command(path, "-m", "1")
	} else if _, err := os.Stat("/usr/share/bcc/tools/biolatency"); err == nil {
		cmd = exec.Command("/usr/share/bcc/tools/biolatency", "-m", "1")
	}

	if cmd == nil {
		return
	}

	cmd.Stdout = &stdout
	c.mu.Lock()
	c.cmds = append(c.cmds, cmd)
	c.mu.Unlock()

	cmd.Start()

	select {
	case <-c.stopChan:
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	cmd.Wait()

	c.mu.Lock()
	c.results.BiolatHistogram = parseHistogram(stdout.String())
	c.results.IoLatencyUs = calculateAvgFromHistogram(c.results.BiolatHistogram)
	c.mu.Unlock()
}

func parseHistogram(output string) map[string]int64 {
	hist := make(map[string]int64)
	
	// Parse bcc-style histogram output
	// Format: "     usecs               : count     distribution"
	//         "         0 -> 1          : 5        |****                                    |"
	re := regexp.MustCompile(`^\s*(\d+)\s*->\s*(\d+)\s*:\s*(\d+)`)
	
	for _, line := range strings.Split(output, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			bucket := matches[1] + "-" + matches[2]
			count, _ := strconv.ParseInt(matches[3], 10, 64)
			hist[bucket] = count
		}
	}

	return hist
}

func calculateAvgFromHistogram(hist map[string]int64) float64 {
	if len(hist) == 0 {
		return 0
	}

	var totalCount int64
	var weightedSum float64

	re := regexp.MustCompile(`(\d+)-(\d+)`)
	for bucket, count := range hist {
		matches := re.FindStringSubmatch(bucket)
		if len(matches) >= 3 {
			low, _ := strconv.ParseFloat(matches[1], 64)
			high, _ := strconv.ParseFloat(matches[2], 64)
			mid := (low + high) / 2
			weightedSum += mid * float64(count)
			totalCount += count
		}
	}

	if totalCount == 0 {
		return 0
	}

	return weightedSum / float64(totalCount)
}

func Aggregate(metrics []Metrics) AggregatedMetrics {
	if len(metrics) == 0 {
		return AggregatedMetrics{}
	}

	agg := AggregatedMetrics{
		TopSyscalls: make(map[string]int64),
	}

	var runqSum, offcpuSum, ioSum float64
	var runqCount, offcpuCount, ioCount int

	for _, m := range metrics {
		if m.RunqueueLatencyUs > 0 {
			runqSum += m.RunqueueLatencyUs
			runqCount++
		}
		if m.OffCpuTimeMs > 0 {
			offcpuSum += m.OffCpuTimeMs
			offcpuCount++
		}
		if m.IoLatencyUs > 0 {
			ioSum += m.IoLatencyUs
			ioCount++
		}
		for syscall, count := range m.TopSyscalls {
			agg.TopSyscalls[syscall] += count
		}
	}

	if runqCount > 0 {
		agg.RunqueueLatencyUs = runqSum / float64(runqCount)
	}
	if offcpuCount > 0 {
		agg.OffCpuTimeMs = offcpuSum / float64(offcpuCount)
	}
	if ioCount > 0 {
		agg.IoLatencyUs = ioSum / float64(ioCount)
	}

	return agg
}
