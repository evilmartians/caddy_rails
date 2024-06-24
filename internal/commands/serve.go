package commands

import (
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/evilmartians/caddy_rails/internal/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

const (
	errStopUpstream       = "failed to stop upstream process: %v"
	errPhasedRestart      = "failed to phased restart upstream process: %v"
	errRunUpstreamProcess = "failed to run upstream process: %v"
	errLoadCaddyConfig    = "Caddyfile loading error"
	errAccessCaddyFile    = "Error accessing Caddyfile"
	infoCaddyFileAbsent   = "Caddyfile does not exist. The server is running direct way"
	infoCaddyFileLoaded   = "Caddyfile is correct and loaded. The server is running via Caddyfile"
)

var logger = utils.NewCaddyRailsLogger()

func init() {
	caddycmd.RegisterCommand(caddycmd.Command{
		Name:  "serve",
		Short: "Runs an external server and sets up a reverse proxy to it",
		Long: `
The serve command runs an external server specified as its argument and
sets up a reverse proxy to forward requests to it.`,
		CobraFunc: func(cmd *cobra.Command) {
			cmd.Flags().String("pid-file", "", "Path to the PID file to control an existing process")
			cmd.Flags().Bool("stop", false, "Stop the running process")
			cmd.Flags().Bool("phased-restart", false, "Perform a phased restart of the server")
			cmd.Flags().Bool("anycable-enabled", false, "Activate AnyCable")
			cmd.Flags().String("server-type", "puma", "The type of server (puma or unicorn) to control")
			cmd.Flags().Int("target-port", 3000, "The port that your server should run on. caddy-rails will set this values to the PORT env variable when starting your server.")
			cmd.Flags().Int("http-port", 80, "The port to listen on for HTTP traffic.")
			cmd.Flags().Int("https-port", 443, "The port to listen on for HTTPS traffic.")
			cmd.Flags().StringP("listen", "l", "localhost", "The address to which to bind the listener")
			cmd.Flags().String("ssl-domain", "", "The domain name to use for SSL provisioning. If not set, SSL will be disabled.")
			cmd.Flags().StringP("files-path", "f", "", "The domain name to use for SSL provisioning. If not set, SSL will be disabled.")
			cmd.Flags().BoolP("debug", "v", false, "Enable verbose debug logs")
			cmd.Flags().Bool("access-log", true, "Enable the access log")
			cmd.Flags().Bool("no-compress", false, "Disable Zstandard and Gzip compression")
			cmd.Flags().Duration("http-idle-timeout", 60*time.Second, "The maximum time in seconds that a client can be idle before the connection is closed.")
			cmd.Flags().Duration("http-read-timeout", 30*time.Second, "The maximum time in seconds that a client can take to send the request headers.")
			cmd.Flags().Duration("http-write-timeout", 30*time.Second, "The maximum time in seconds during which the client must read the response.")
			cmd.RunE = caddycmd.WrapCommandFuncForCobra(cmdServeRails)
		},
	})
}

func cmdServeRails(fs caddycmd.Flags) (int, error) {
	caddy.TrapSignals()

	pidFile := fs.String("pid-file")
	stop := fs.Bool("stop")
	phasedRestart := fs.Bool("phased-restart")
	serverType := fs.String("server-type")

	if stop || phasedRestart {
		upstream, err := utils.NewUpstreamProcess("", nil, false, pidFile)
		if err != nil {
			return 1, err
		}

		if stop {
			if err := upstream.Stop(); err != nil {
				return 1, fmt.Errorf(errStopUpstream, err)
			}
			return 0, nil
		}

		if phasedRestart {
			if err := upstream.PhasedRestart(serverType); err != nil {
				return 1, fmt.Errorf(errPhasedRestart, err)
			}
			return 0, nil
		}
	}

	if loadConfigIfNeeded() {
		select {}
	}

	if err := utils.StartCaddyReverseProxy(fs); err != nil {
		return 1, err
	}

	return runUpstreamProcess(fs, pidFile)
}

func loadConfigIfNeeded() bool {
	curdir, _ := os.Getwd()
	configFile := filepath.Join(curdir, "Caddyfile")

	if _, err := os.Stat(configFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Info(infoCaddyFileAbsent)
		} else {
			logger.Error(errAccessCaddyFile, zap.Error(err))
		}
		return false
	}

	config, _, err := caddycmd.LoadConfig(configFile, "")
	if err != nil {
		logger.Error(errLoadCaddyConfig, zap.Error(err))

		return false
	}

	err = caddy.Load(config, true)
	if err != nil {
		logger.Error(errLoadCaddyConfig, zap.Error(err))

		return false
	}

	logger.Info(infoCaddyFileLoaded)

	return true
}

func runUpstreamProcess(fs caddycmd.Flags, pidFile string) (int, error) {
	// Set PORT to be inherited by the upstream process.
	os.Setenv("PORT", fmt.Sprintf("%d", fs.Int("target-port")))

	var args []string
	if len(fs.Args()) > 0 {
		args = fs.Args()[1:]
	}

	upstream, err := utils.NewUpstreamProcess(fs.Arg(0), args, true, pidFile)
	if err != nil {
		return 1, err
	}

	exitCode, err := upstream.Run()
	if err != nil {
		return 1, fmt.Errorf(errRunUpstreamProcess, err)
	}
	return exitCode, nil
}
