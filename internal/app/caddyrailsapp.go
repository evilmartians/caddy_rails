package app

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/evilmartians/caddy_thruster/internal/utils"
)

func init() {
	caddy.RegisterModule(CaddyRailsApp{})
	httpcaddyfile.RegisterGlobalOption("serve-rails", parseGlobalCaddyfileBlock)
}

type CaddyRailsApp struct {
	Command []string `json:"command,omitempty"`
	PidFile string   `json:"pid_file,omitempty"`

	process *utils.UpstreamProcess
}

// Provision implements caddy.Provisioner
func (a *CaddyRailsApp) Provision(ctx caddy.Context) error {
	if len(a.Command) == 0 {
		return fmt.Errorf("there is not any command")
	}

	//
	//os.Setenv("PORT", fmt.Sprintf("%d", a.TargetPort))

	a.process = utils.NewUpstreamProcess(a.Command[0], a.Command[1:], false, a.PidFile)

	return nil
}

// Start starts the app.
func (a CaddyRailsApp) Start() error {
	_, err := a.process.Run()

	return err
}

// Stop stops the app.
func (a *CaddyRailsApp) Stop() error {
	caddy.Log().Info("CaddyRails stopped 🐘")

	return nil

	//return a.process.Stop()
}

// CaddyModule implements caddy.ModuleInfo
func (a CaddyRailsApp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "serve-rails",
		New: func() caddy.Module { return new(CaddyRailsApp) },
	}
}

func parseGlobalCaddyfileBlock(d *caddyfile.Dispenser, _ interface{}) (interface{}, error) {
	var caddy_rails CaddyRailsApp

	if d.NextArg() && d.NextArg() {
		caddy_rails.Command = append([]string{d.Val()}, d.RemainingArgs()...)
	}

	for d.NextBlock(0) {
		switch d.Val() {
		case "command":
			caddy_rails.Command = d.RemainingArgs()
		case "pid_file":
			if d.NextArg() {
				caddy_rails.PidFile = d.Val()
			} else {
				return nil, d.ArgErr()
			}
		}
	}

	return httpcaddyfile.App{
		Name:  "serve-rails",
		Value: caddyconfig.JSON(caddy_rails, nil),
	}, nil
}

var (
	_ caddy.App         = (*CaddyRailsApp)(nil)
	_ caddy.Module      = (*CaddyRailsApp)(nil)
	_ caddy.Provisioner = (*CaddyRailsApp)(nil)
)
