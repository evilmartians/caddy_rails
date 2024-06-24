package utils

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	RailsExecutionFile = "bin/rails"
	DefaultPidPath     = "tmp/pids/server.pid"
)

var logger = NewCaddyRailsLogger()

type UpstreamProcess struct {
	Started  chan struct{}
	Cmd      *exec.Cmd
	SyncMode bool
	PidFile  string
	workers  []int
	stopCh   chan struct{}
	wg       sync.WaitGroup
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

	return &UpstreamProcess{
		Started:  make(chan struct{}, 1),
		Cmd:      exec.Command(command, arguments...),
		SyncMode: syncMode,
		PidFile:  pidFile,
		stopCh:   make(chan struct{}),
	}, nil
}

func (p *UpstreamProcess) Run(testMode bool) (int, error) {
	logger.Info("Starting Run method")
	if p.PidFile != "" {
		pid, err := p.readPidFile()
		if err == nil {
			p.Cmd = &exec.Cmd{Process: &os.Process{Pid: pid}}
			p.Started <- struct{}{}
			if p.SyncMode {
				select {}
			}
			return 0, nil
		}
	}

	p.Cmd.Stdout = os.Stdout
	p.Cmd.Stderr = os.Stderr

	err := p.Cmd.Start()
	if err != nil {
		return 0, err
	}

	p.wg.Add(1)
	p.Started <- struct{}{}

	err = p.waitForPidFile()
	if err != nil {
		return 0, err
	}

	p.checkForNewWorkers()

	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logger.Info("Received signal", zap.String("signal", sig.String()))
		p.terminateWorkers(sig)
		p.Shutdown()
		close(p.stopCh)
	}()

	if p.SyncMode {
		<-p.stopCh
		p.wg.Wait()
	}

	logger.Info("Run method finished successfully")
	return 0, nil
}

func (p *UpstreamProcess) Shutdown() error {
	logger.Info("Starting Shutdown method")
	if p.Cmd != nil && p.Cmd.Process != nil {
		err := p.Cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}

		logger.Info("Waiting for main process to finish")
		err = p.Cmd.Wait()
		if err != nil {
			logger.Error("Failed to wait for process", zap.Error(err))
			return err
		}

		p.wg.Done()
		logger.Info("Main process finished")
		return nil
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

	return errors.New("PidFile is not exists")
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

func (p *UpstreamProcess) waitForPidFile() error {
	timeout := time.After(20 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for PID file")
		case <-tick:
			if fileExists(p.PidFile) {
				logger.Info("PID file found")
				return nil
			}
		}
	}
}

func (p *UpstreamProcess) checkForNewWorkers() {
	output, err := exec.Command("pgrep", "-P", strconv.Itoa(p.Cmd.Process.Pid)).Output()
	if err != nil {
		logger.Warn("Failed to check for worker processes", zap.Error(err))
		return
	}

	pidStrings := strings.Fields(string(output))
	for _, pidStr := range pidStrings {
		pid, err := strconv.Atoi(pidStr)
		if err == nil {
			logger.Info("Added worker with pid", zap.Int("pid", pid))
			p.workers = append(p.workers, pid)
			p.wg.Add(1) // Add worker to wait group
		} else {
			logger.Error("Error with worker pid", zap.Int("pid", pid), zap.Error(err))
		}
	}
}

func (p *UpstreamProcess) terminateWorkers(sig os.Signal) {
	for _, pid := range p.workers {
		process, err := os.FindProcess(pid)
		if err != nil {
			logger.Warn("Failed to find worker process", zap.Int("pid", pid), zap.Error(err))
			continue
		}
		err = process.Signal(sig)
		if err != nil {
			logger.Warn("Failed to signal worker process", zap.Int("pid", pid), zap.Error(err))
		} else {
			logger.Info("Signaled worker process", zap.Int("pid", pid), zap.String("signal", sig.String()))
			p.wg.Done() // Mark worker as done
		}
	}

	p.workers = make([]int, 0)
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
