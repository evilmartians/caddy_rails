package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const DefaultPidPath = "tmp/pids/server.pid"

type UpstreamProcess struct {
	Started  chan struct{}
	cmd      *exec.Cmd
	SyncMode bool
	PidFile  string
}

func NewUpstreamProcess(name string, arg []string, syncMode bool, pidFile string) *UpstreamProcess {
	if pidFile == "" {
		pidFile = DefaultPidPath
	}

	return &UpstreamProcess{
		Started:  make(chan struct{}, 1),
		cmd:      exec.Command(name, arg...),
		SyncMode: syncMode,
		PidFile:  pidFile,
	}
}

func (p *UpstreamProcess) Run() (int, error) {
	if p.PidFile != "" {
		pidData, err := os.ReadFile(p.PidFile)
		if err == nil {
			pid, err := strconv.Atoi(string(pidData))
			if err == nil {
				p.cmd = &exec.Cmd{Process: &os.Process{Pid: pid}}
				p.Started <- struct{}{}
				if p.SyncMode {
					return p.waitAndHandleExit()
				}
				go p.waitAndHandleExit()
				return 0, nil
			}
		}
	}

	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	err := p.cmd.Start()
	if err != nil {
		return 0, err
	}

	p.Started <- struct{}{}

	go p.handleSignals()

	if p.SyncMode {
		return p.waitAndHandleExit()
	}

	go p.waitAndHandleExit()
	return 0, nil
}

func (p *UpstreamProcess) waitAndHandleExit() (int, error) {
	err := p.cmd.Wait()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	return 0, err
}
func (p *UpstreamProcess) Stop() error {
	if p.PidFile != "" {
		pidData, err := os.ReadFile(p.PidFile)
		if err != nil {
			return err
		}
		pid, err := strconv.Atoi(string(pidData))
		if err != nil {
			return err
		}
		process, err := os.FindProcess(pid)
		if err != nil {
			return err
		}
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return err
		}

		// Wait for the process to exit with a timeout
		done := make(chan struct{})
		go func() {
			_, _ = process.Wait()
			close(done)
		}()
		select {
		case <-done:
			return nil
		case <-time.After(5 * time.Second):
			return errors.New("process did not terminate in time")
		}
	}

	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Signal(syscall.SIGTERM)
	}

	return errors.New("process is not running or PID file is not specified")
}

func (p *UpstreamProcess) PhasedRestart(serverType string) error {
	var signal os.Signal

	switch serverType {
	case "puma":
		signal = syscall.SIGUSR1
	case "unicorn":
		signal = syscall.SIGUSR2
	default:
		return fmt.Errorf("unknown server type: %s", serverType)
	}

	pidData, err := os.ReadFile(p.PidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %v", err)
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return fmt.Errorf("failed to convert PID to integer: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process with PID %d: %v", pid, err)
	}

	if err := process.Signal(signal); err != nil {
		return fmt.Errorf("failed to send signal to process: %v", err)
	}

	return nil
}

func (p *UpstreamProcess) Signal(sig os.Signal) error {
	return p.cmd.Process.Signal(sig)
}

func (p *UpstreamProcess) handleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	sig := <-ch
	slog.Info("Relaying signal to upstream process", "signal", sig.String())
	p.Signal(sig)
}
