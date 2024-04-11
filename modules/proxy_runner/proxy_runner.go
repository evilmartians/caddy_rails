package proxy_runner

import (
	"fmt"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"time"
)

func init() {
	caddycmd.RegisterCommand(caddycmd.Command{
		Name:  "proxy-runner",
		Short: "Runs an external server and sets up a reverse proxy to it",
		Long: `
The proxy_runner command runs an external server specified as its argument and
sets up a reverse proxy to forward requests to it.`,
		CobraFunc: func(cmd *cobra.Command) {
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
			cmd.RunE = caddycmd.WrapCommandFuncForCobra(cmdProxyRunner)
		},
	})
}

func cmdProxyRunner(fs caddycmd.Flags) (int, error) {
	if fs.NArg() < 1 {
		fmt.Println("Usage: caddy_thruster proxy_runner <command> [arguments...]")
		return 1, fmt.Errorf("incorrect usage")
	}

	slog.Info("Server started", "listen_domain", fs.String("listen"), "http", fs.String("http_port"), "https", fs.String("https_port"))

	upstream := NewUpstreamProcess(fs.Arg(0), fs.Args()[1:]...)

	err := startCaddyReverseProxy(fs)
	if err != nil {
		fmt.Println("Error starting Caddy server:", err)
		return 1, err
	}

	os.Setenv("PORT", fmt.Sprintf("%d", fs.Int("target_port")))

	exitCode, err := upstream.Run()
	if err != nil {
		panic(err)
	}

	return exitCode, nil
}
