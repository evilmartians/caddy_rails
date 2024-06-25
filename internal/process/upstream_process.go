package process

import (
	"errors"
	"fmt"
	"github.com/evilmartians/caddy_rails/internal/logger"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

const (
	RailsExecutionFile = "bin/rails"
	DefaultPidPath     = "tmp/pids/server.pid"
)

var log = logger.NewCaddyRailsLogger()

type UpstreamProcess struct {
	Cmd      *exec.Cmd
	SyncMode bool
	PidFile  string
}

func NewUpstreamProcess(command string, arg []string, syncMode bool, pidFile string) (*UpstreamProcess, error) {
	if pidFile == "" {
		pidFile = DefaultPidPath
	}

	command, arguments := determineCommand(command, arg)

	if command == "" && !fileExists(pidFile) {
		log.Error("For running an application, you must provide either an argument to the command serve-rails or ensure the presence of " + RailsExecutionFile)
		return nil, fmt.Errorf("for running an application, you must provide either an argument to the command serve-rails or ensure the presence of %s file", RailsExecutionFile)
	}

	return &UpstreamProcess{
		Cmd:      exec.Command(command, arguments...),
		SyncMode: syncMode,
		PidFile:  pidFile,
	}, nil
}

func (p *UpstreamProcess) Run() error {
	if p.PidFile != "" {
		pid, err := p.readPidFile()
		if err == nil {
			p.Cmd = &exec.Cmd{Process: &os.Process{Pid: pid}}
			log.Info("Connected to the server with PID", zap.Int("pid", pid))
			return nil
		}
	}

	p.Cmd.Stdout = os.Stdout
	p.Cmd.Stderr = os.Stderr

	err := p.Cmd.Start()

	return err
}

func (p *UpstreamProcess) Shutdown(sig os.Signal) error {
	if p.Cmd != nil && p.Cmd.Process != nil {
		err := p.signal(sig)
		if err != nil {
			return err
		}

		err = p.Cmd.Wait()
		return err
	}

	return errors.New("process is not running or PID file is not specified")
}

func (p *UpstreamProcess) Stop() error {
	if p.PidFile != "" {
		pid, err := p.readPidFile()
		if err != nil {
			return err
		}

		return p.signalProcess(pid, syscall.SIGTERM)
	}

	return errors.New("PidFile does not exist")
}

func (p *UpstreamProcess) PhasedRestart(serverType string) error {
	sig, err := determineSignal(serverType)
	if err != nil {
		return err
	}

	pid, err := p.readPidFile()
	if err != nil {
		return err
	}

	log.Info("Hot restarting the rails application", zap.Int("pid", pid))
	return p.signalProcess(pid, sig)
}

func (p *UpstreamProcess) readPidFile() (int, error) {
	pidData, err := os.ReadFile(p.PidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %v", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		return 0, fmt.Errorf("failed to convert PID to integer: %v", err)
	}

	return pid, nil
}

func (p *UpstreamProcess) signalProcess(pid int, sig os.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(sig)
}

func determineSignal(serverType string) (os.Signal, error) {
	switch serverType {
	case "puma":
		return syscall.SIGUSR1, nil
	case "unicorn":
		return syscall.SIGUSR2, nil
	default:
		return nil, fmt.Errorf("unknown server type: %s", serverType)
	}
}

func determineCommand(command string, arg []string) (string, []string) {
	if command == "" {
		if fileExists(RailsExecutionFile) {
			return RailsExecutionFile, []string{"server"}
		}
	}
	return command, arg
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}

	return false
}

func (p *UpstreamProcess) signal(sig os.Signal) error {
	return p.Cmd.Process.Signal(sig)
}
