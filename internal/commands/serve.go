package commands

import (
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/evilmartians/caddy_rails/internal/logger"
	"github.com/evilmartians/caddy_rails/internal/process"
	"github.com/evilmartians/caddy_rails/internal/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

var caddyLogger = logger.NewCaddyRailsLogger()

func init() {
	caddycmd.RegisterCommand(caddycmd.Command{
		Name:  "serve",
		Short: "Runs an external server and sets up a reverse proxy to it",
		CobraFunc: func(cmd *cobra.Command) {
			addServeFlags(cmd)
			cmd.RunE = caddycmd.WrapCommandFuncForCobra(cmdServeRails)
		},
	})
}

func addServeFlags(cmd *cobra.Command) {
	cmd.Flags().String("pid-file", "", "Path to the PID file to control an existing process")
	cmd.Flags().Bool("stop", false, "Stop the running process")
	cmd.Flags().Bool("phased-restart", false, "Perform a phased restart of the server")
	cmd.Flags().String("server-type", "puma", "The type of server (puma or unicorn) to control")
	cmd.Flags().Int("target-port", 3000, "The port that your server should run on")
	cmd.Flags().Int("http-port", 80, "The port to listen on for HTTP traffic.")
	cmd.Flags().Int("https-port", 443, "The port to listen on for HTTPS traffic.")
	cmd.Flags().StringP("listen", "l", "localhost", "The address to which to bind the listener")
	cmd.Flags().String("ssl-domain", "", "The domain name to use for SSL provisioning. If not set, SSL will be disabled.")
	cmd.Flags().StringP("files-path", "f", "", "The path to the files to serve")
	cmd.Flags().BoolP("debug", "v", false, "Enable verbose debug logs")
	cmd.Flags().Bool("access-log", true, "Enable the access log")
	cmd.Flags().Bool("no-compress", false, "Disable Zstandard and Gzip compression")
	cmd.Flags().Bool("anycable-enabled", false, "Activate AnyCable")
	cmd.Flags().Duration("http-idle-timeout", 60*time.Second, "The maximum idle time for HTTP connections.")
	cmd.Flags().Duration("http-read-timeout", 30*time.Second, "The maximum read timeout for HTTP connections.")
	cmd.Flags().Duration("http-write-timeout", 30*time.Second, "The maximum write timeout for HTTP connections.")
}

func cmdServeRails(fs caddycmd.Flags) (int, error) {
	caddy.TrapSignals()

	pidFile := fs.String("pid-file")
	stop := fs.Bool("stop")
	phasedRestart := fs.Bool("phased-restart")
	serverType := fs.String("server-type")

	if stop || phasedRestart {
		return handleProcessControl(pidFile, stop, phasedRestart, serverType)
	}

	loaded, err := loadConfigIfNeeded()
	if err != nil {
		return 1, err
	}
	if loaded {
		select {}
	}

	if fs.Int("target-port") != 0 {
		os.Setenv("PORT", fmt.Sprintf("%d", fs.Int("target-port")))
	}

	if err := utils.StartCaddyReverseProxy(fs); err != nil {
		return 1, err
	}

	select {}
}

func handleProcessControl(pidFile string, stop, phasedRestart bool, serverType string) (int, error) {
	upstream, err := process.NewUpstreamProcess("", nil, false, pidFile)
	if err != nil {
		return 1, err
	}

	if stop {
		if err := upstream.Stop(); err != nil {
			return 1, fmt.Errorf("failed to stop upstream process: %v", err)
		}
		return 0, nil
	}

	if phasedRestart {
		if err := upstream.PhasedRestart(serverType); err != nil {
			return 1, fmt.Errorf("failed to phased restart upstream process: %v", err)
		}
		return 0, nil
	}

	return 1, errors.New("invalid process control command")
}

func loadConfigIfNeeded() (bool, error) {
	curdir, _ := os.Getwd()
	configFile := filepath.Join(curdir, "Caddyfile")

	if _, err := os.Stat(configFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			caddyLogger.Info("Caddyfile does not exist. Running directly")
		} else {
			caddyLogger.Error("Error accessing Caddyfile", zap.Error(err))
		}
		return false, nil
	}

	config, _, err := caddycmd.LoadConfig(configFile, "")
	if err != nil {
		caddyLogger.Error("Caddyfile loading error", zap.Error(err))
		return false, err
	}

	if err := caddy.Load(config, true); err != nil {
		caddyLogger.Error("Caddyfile loading error", zap.Error(err))
		return false, err
	}

	caddyLogger.Info("Caddyfile loaded successfully")
	return true, nil
}
