package commands

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/evilmartians/caddy_thruster/internal/utils"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

func init() {
	caddycmd.RegisterCommand(caddycmd.Command{
		Name:  "serve-rails",
		Short: "Runs an external server and sets up a reverse proxy to it",
		Long: `
The proxy_runner command runs an external server specified as its argument and
sets up a reverse proxy to forward requests to it.`,
		CobraFunc: func(cmd *cobra.Command) {
			cmd.Flags().String("pid_file", "", "Path to the PID file to control an existing process")
			cmd.Flags().Bool("stop", false, "Stop the running process")
			cmd.Flags().Bool("phased-restart", false, "Perform a phased restart of the server")
			cmd.Flags().String("server-type", "puma", "The type of server (puma or unicorn) to control")
			cmd.Flags().Int("target_port", 3000, "The port that your server should run on. ProxyRunner will set PORT to this value when starting your server.")
			cmd.Flags().String("http_port", "80", "The port to listen on for HTTP traffic.")
			cmd.Flags().String("https_port", "443", "The port to listen on for HTTPS traffic.")
			cmd.Flags().StringP("listen", "l", "localhost", "The address to which to bind the listener")
			cmd.Flags().String("ssl_domain", "", "The domain name to use for SSL provisioning. If not set, SSL will be disabled.")
			cmd.Flags().StringP("files_path", "f", "", "The domain name to use for SSL provisioning. If not set, SSL will be disabled.")
			cmd.Flags().BoolP("debug", "v", false, "Enable verbose debug logs")
			cmd.Flags().Bool("access_log", true, "Enable the access log")
			cmd.Flags().Bool("no-compress", false, "Disable Zstandard and Gzip compression")
			cmd.Flags().Duration("http_idle_timeout", 60*time.Second, "The maximum time in seconds that a client can be idle before the connection is closed.")
			cmd.Flags().Duration("http_read_timeout", 30*time.Second, "The maximum time in seconds that a client can take to send the request headers.")
			cmd.Flags().Duration("http_write_timeout", 30*time.Second, "The maximum time in seconds during which the client must read the response.")
			cmd.RunE = caddycmd.WrapCommandFuncForCobra(cmdThruster)
		},
	})
}

func cmdThruster(fs caddycmd.Flags) (int, error) {
	caddy.TrapSignals()

	pidFile := fs.String("pid_file")
	stop := fs.Bool("stop")
	phasedRestart := fs.Bool("phased-restart")
	serverType := fs.String("server-type")

	if stop || phasedRestart {
		upstream := utils.NewUpstreamProcess("", nil, false, pidFile)

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
	}

	if loadConfigIfNeeded() {
		select {}
	}

	if fs.NArg() < 1 {
		return 1, fmt.Errorf("usage: serve_rails <command> [arguments...]")
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
		return false
	}

	if config, _, err := caddycmd.LoadConfig(configFile, ""); err == nil {
		if err = caddy.Load(config, true); err == nil {
			return true
		}
	}

	return false
}

func runUpstreamProcess(fs caddycmd.Flags, pidFile string) (int, error) {
	upstream := utils.NewUpstreamProcess(fs.Arg(0), fs.Args()[1:], true, pidFile)
	exitCode, err := upstream.Run()
	if err != nil {
		return 1, fmt.Errorf("failed to run upstream process: %v", err)
	}
	return exitCode, nil
}
