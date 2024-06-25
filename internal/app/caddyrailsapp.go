package app

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/evilmartians/caddy_rails/internal/logger"
	"github.com/evilmartians/caddy_rails/internal/process"
	"os"
)

var caddyLogger = logger.NewCaddyRailsLogger()

func init() {
	caddy.RegisterModule(CaddyRailsApp{})
	httpcaddyfile.RegisterGlobalOption("serve", parseGlobalCaddyfileBlock)
}

type CaddyRailsApp struct {
	Command []string `json:"command,omitempty"`
	PidFile string   `json:"pid_file,omitempty"`

	process *process.UpstreamProcess
}

func (a *CaddyRailsApp) Provision(_ caddy.Context) error {
	return nil
}

func (a *CaddyRailsApp) Start() error {
	var command string
	var arguments []string

	if len(a.Command) == 0 {
		command = ""
	} else {
		command = a.Command[0]
		arguments = a.Command[1:]
	}

	proc, err := process.NewUpstreamProcess(command, arguments, false, a.PidFile)
	if err != nil {
		return err
	}

	a.process = proc

	caddyLogger.Info("CaddyRails was started")

	return a.process.Run()
}

func (a *CaddyRailsApp) Stop() error {
	err := a.process.Shutdown(os.Interrupt)
	if err != nil {
		return err
	}

	caddyLogger.Info("CaddyRails stopped")

	return nil
}

func (a CaddyRailsApp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "serve",
		New: func() caddy.Module { return new(CaddyRailsApp) },
	}
}

func parseGlobalCaddyfileBlock(d *caddyfile.Dispenser, _ interface{}) (interface{}, error) {
	var caddyRails CaddyRailsApp

	if d.NextArg() && d.NextArg() {
		caddyRails.Command = append([]string{d.Val()}, d.RemainingArgs()...)
	}

	for d.NextBlock(0) {
		switch d.Val() {
		case "command":
			caddyRails.Command = d.RemainingArgs()
		case "pid_file":
			if d.NextArg() {
				caddyRails.PidFile = d.Val()
			} else {
				return nil, d.ArgErr()
			}
		}
	}

	return httpcaddyfile.App{
		Name:  "serve",
		Value: caddyconfig.JSON(caddyRails, nil),
	}, nil
}

var (
	_ caddy.App         = (*CaddyRailsApp)(nil)
	_ caddy.Module      = (*CaddyRailsApp)(nil)
	_ caddy.Provisioner = (*CaddyRailsApp)(nil)
)
