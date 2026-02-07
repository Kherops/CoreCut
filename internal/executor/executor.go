package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type RunResult struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	DurationMs  float64   `json:"duration_ms"`
	ExitCode    int       `json:"exit_code"`
	Stdout      string    `json:"stdout_tail"`
	Stderr      string    `json:"stderr_tail"`
	Error       string    `json:"error,omitempty"`
	Throughput  float64   `json:"throughput,omitempty"`
	PID         int       `json:"pid"`
}

type Executor struct {
	TimeoutSec int
	CooldownMs int
	EnvFile    string
	Env        []string
}

func New(timeoutSec, cooldownMs int, envFile string) *Executor {
	e := &Executor{
		TimeoutSec: timeoutSec,
		CooldownMs: cooldownMs,
		EnvFile:    envFile,
		Env:        os.Environ(),
	}

	if envFile != "" {
		e.loadEnvFile(envFile)
	}

	return e
}

func (e *Executor) loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "=") {
			e.Env = append(e.Env, line)
		}
	}
}

func (e *Executor) Run(script string, mode string) (RunResult, error) {
	result := RunResult{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.TimeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", script)
	cmd.Env = e.Env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	result.StartTime = time.Now()

	if err := cmd.Start(); err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.PID = cmd.Process.Pid

	err := cmd.Wait()
	result.EndTime = time.Now()
	result.DurationMs = float64(result.EndTime.Sub(result.StartTime).Microseconds()) / 1000.0

	result.ExitCode = cmd.ProcessState.ExitCode()
	result.Stdout = tailString(stdout.String(), 1000)
	result.Stderr = tailString(stderr.String(), 500)

	if ctx.Err() == context.DeadlineExceeded {
		result.Error = "timeout"
		return result, fmt.Errorf("timeout after %ds", e.TimeoutSec)
	}

	if err != nil {
		result.Error = err.Error()
	}

	// Parse throughput if mode=throughput
	if mode == "throughput" {
		result.Throughput = parseThroughput(stdout.String())
	}

	// Cooldown
	if e.CooldownMs > 0 {
		time.Sleep(time.Duration(e.CooldownMs) * time.Millisecond)
	}

	return result, nil
}

func tailString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return "..." + s[len(s)-maxLen:]
}

func parseThroughput(output string) float64 {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return 0
	}

	// Try to parse the last line as a number (throughput value)
	lastLine := strings.TrimSpace(lines[len(lines)-1])

	// Try formats: "THROUGHPUT: 1234.56" or just "1234.56"
	if strings.Contains(strings.ToUpper(lastLine), "THROUGHPUT") {
		parts := strings.Split(lastLine, ":")
		if len(parts) >= 2 {
			lastLine = strings.TrimSpace(parts[len(parts)-1])
		}
	}

	// Remove units like "ops/sec", "req/s", etc.
	lastLine = strings.Fields(lastLine)[0]

	val, err := strconv.ParseFloat(lastLine, 64)
	if err != nil {
		return 0
	}
	return val
}
