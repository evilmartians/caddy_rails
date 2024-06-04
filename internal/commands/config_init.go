package commands

import (
	"bytes"
	_ "embed"
	"fmt"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed CaddyfileTemplate
var caddyfileTemplate string

type TemplateParams struct {
	Debug             bool
	SSLDomain         string
	AccessLog         bool
	HttpHost          string
	HttpPort          string
	AnyCableEnable    bool
	CompressionEnable bool
	BackendPort       string
	HttpsPort         string
}

func init() {
	caddycmd.RegisterCommand(caddycmd.Command{
		Name:  "config-init",
		Short: "Generates a Caddyfile with custom configuration",
		Long: `Generates a Caddyfile in the specified directory with options for SSL, logging, and debug.
This tool simplifies initial Caddyfile setup for custom server configurations.`,
		CobraFunc: func(cmd *cobra.Command) {
			cmd.Flags().String("folder-path", ".", "Directory to generate the Caddyfile in. Defaults to the current directory.")
			cmd.Flags().String("http-host", "localhost", "Host address for the HTTP server.")
			cmd.Flags().String("http-port", "80", "Port for HTTP traffic.")
			cmd.Flags().String("https-port", "443", "Port for HTTPS traffic.")
			cmd.Flags().Bool("debug-enable", false, "Enable debug mode for detailed logs")
			cmd.Flags().Bool("anycable-enable", false, "Enable AnyCable")
			cmd.Flags().String("ssl-domain", "", "Domain name for SSL. If empty, SSL is disabled")
			cmd.Flags().String("backend-port", "3000", "Port that the backend service listens on")
			cmd.Flags().Bool("access-log", false, "Enable logging of access requests")
			cmd.Flags().Bool("compression-enable", true, "Enable response compression using gzip and zstd")
			cmd.RunE = caddycmd.WrapCommandFuncForCobra(cmdGenerateCaddyfile)
		},
	})
}

func cmdGenerateCaddyfile(fs caddycmd.Flags) (int, error) {
	folderPath := fs.String("folder-path")
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, 0700)
	}

	params := TemplateParams{
		Debug:             fs.Bool("debug-enable"),
		SSLDomain:         fs.String("ssl-domain"),
		AccessLog:         fs.Bool("access-log"),
		CompressionEnable: fs.Bool("compression-enable"),
		AnyCableEnable:    fs.Bool("anycable-enable"),
		HttpHost:          fs.String("http-host"),
		HttpPort:          fs.String("http-port"),
		BackendPort:       fs.String("backend-port"),
		HttpsPort:         fs.String("https-port"),
	}

	tmpl, err := template.New("Caddyfile").Parse(caddyfileTemplate)
	if err != nil {
		return 1, fmt.Errorf("error parsing template: %v", err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, params); err != nil {
		return 1, fmt.Errorf("error executing template: %v", err)
	}

	filename := filepath.Join(folderPath, "Caddyfile")
	if err := os.WriteFile(filename, out.Bytes(), 0666); err != nil {
		return 1, fmt.Errorf("error writing file: %v", err)
	}

	log.Printf("Caddyfile generated at %s", filename)
	return 0, nil
}
