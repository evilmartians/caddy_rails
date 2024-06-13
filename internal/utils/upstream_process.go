package utils

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

const (
	RailsExecutionFile = "bin/rails"
	DefaultPidPath     = "tmp/pids/server.pid"
)

var logger = NewCaddyRailsLogger()

type UpstreamProcess struct {
	Started  chan struct{}
	cmd      *exec.Cmd
	SyncMode bool
	PidFile  string
}

func NewUpstreamProcess(command string, arg []string, syncMode bool, pidFile string) (*UpstreamProcess, error) {
	if pidFile == "" {
		pidFile = DefaultPidPath
	}

	command, arguments := determineCommand(command, arg)

	if command == "" && !fileExists(pidFile) {
		logger.Error("For running an application, you must provide either an argument to the command serve-rails or ensure the presence of " + RailsExecutionFile)

		return nil, fmt.Errorf("for running an application, you must provide either an argument to the command serve-rails or ensure the presence of %s file", RailsExecutionFile)
	}

	//logger.Info("Running a rails application by command: ", zap.String(command, strings.Join(arguments, " ")))

	return &UpstreamProcess{
		Started:  make(chan struct{}, 1),
		cmd:      exec.Command(command, arguments...),
		SyncMode: syncMode,
		PidFile:  pidFile,
	}, nil
}

func (p *UpstreamProcess) Run() (int, error) {
	if p.PidFile != "" {
		pid, err := p.readPidFile()

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

func (p *UpstreamProcess) Stop() error {
	if p.PidFile != "" {
		pid, err := p.readPidFile()
		if err != nil {
			return err
		}

		logger.Info("Stopping the rails application", zap.Int("pid", pid))

		return p.signalProcess(pid, syscall.SIGTERM)
	}

	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Signal(syscall.SIGTERM)
	}

	return errors.New("process is not running or PID file is not specified")
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

	logger.Info("Hot restarting the rails application", zap.Int("pid", pid))

	return p.signalProcess(pid, sig)
}

func (p *UpstreamProcess) waitAndHandleExit() (int, error) {
	err := p.cmd.Wait()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	return 0, err
}

func (p *UpstreamProcess) readPidFile() (int, error) {
	pidData, err := os.ReadFile(p.PidFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read PID file: %v", err)
	}

	pid, err := strconv.Atoi(string(pidData))
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

func (p *UpstreamProcess) signal(sig os.Signal) error {
	return p.cmd.Process.Signal(sig)
}

func (p *UpstreamProcess) handleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	sig := <-ch

	logger.Info("Relaying signal to upstream process", zap.String("signal", sig.String()))

	p.signal(sig)
}

func determineCommand(command string, arg []string) (string, []string) {
	if command == "" {
		if fileExists(RailsExecutionFile) {
			return RailsExecutionFile, []string{"server"}
		}
	}
	return command, arg
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

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}

	return false
}
