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
	HttpsEnable       bool
	EnableCompression bool
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
			cmd.Flags().String("folder_path", ".", "Directory to generate the Caddyfile in. Defaults to the current directory.")
			cmd.Flags().String("http_host", "localhost", "Host address for the HTTP server.")
			cmd.Flags().String("http_port", "80", "Port for HTTP traffic.")
			cmd.Flags().String("https_port", "443", "Port for HTTPS traffic.")
			cmd.Flags().Bool("enable_debug", false, "Enable debug mode for detailed logs")
			cmd.Flags().String("ssl_domain", "", "Domain name for SSL. If empty, SSL is disabled")
			cmd.Flags().String("backend_port", "8080", "Port that the backend service listens on")
			cmd.Flags().Bool("access_log", false, "Enable logging of access requests")
			cmd.Flags().Bool("https_enable", false, "Enable HTTPS configuration")
			cmd.Flags().Bool("enable_compression", true, "Enable response compression using gzip and zstd")
			cmd.RunE = caddycmd.WrapCommandFuncForCobra(cmdGenerateCaddyfile)
		},
	})
}

func cmdGenerateCaddyfile(fs caddycmd.Flags) (int, error) {
	folderPath := fs.String("folder_path")
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, 0700)
	}

	params := TemplateParams{
		Debug:             fs.Bool("enable_debug"),
		SSLDomain:         fs.String("ssl_domain"),
		AccessLog:         fs.Bool("access_log"),
		HttpsEnable:       fs.Bool("https_enable"),
		EnableCompression: fs.Bool("enable_compression"),
		HttpHost:          fs.String("http_host"),
		HttpPort:          fs.String("http_port"),
		BackendPort:       fs.String("backend_port"),
		HttpsPort:         fs.String("https_port"),
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
