package xray

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type Runner struct {
	xrayBin string
	mu      sync.Mutex
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	logBuf  *RingBuffer
	LogCh   chan string
}

func NewRunner(xrayBin string) *Runner {
	return &Runner{
		xrayBin: xrayBin,
		logBuf:  NewRingBuffer(500),
		LogCh:   make(chan string, 100),
	}
}

func (r *Runner) Start(configPath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil {
		return fmt.Errorf("xray already running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, r.xrayBin, "run", "-config", configPath)

	pr, pw, err := os.Pipe()
	if err != nil {
		cancel()
		return err
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		cancel()
		pr.Close()
		pw.Close()
		return err
	}

	r.cmd = cmd
	r.cancel = cancel

	go func() {
		pw.Close()
		defer pr.Close()
		sc := bufio.NewScanner(pr)
		for sc.Scan() {
			line := sc.Text()
			r.logBuf.Push(line)
			select {
			case r.LogCh <- line:
			default:
			}
		}
	}()

	return nil
}

func (r *Runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil {
		return nil
	}
	r.cancel()
	r.cmd.Wait()
	r.cmd = nil
	r.cancel = nil
	return nil
}

func (r *Runner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cmd != nil
}

func (r *Runner) RecentLogs() []string {
	return r.logBuf.Lines()
}
