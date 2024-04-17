package thruster

import (
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func init() {
	caddy.RegisterModule(ThrusterApp{})
	httpcaddyfile.RegisterGlobalOption("thruster", parseGlobalCaddyfileBlock)
}

type ThrusterApp struct {
	Command []string `json:"command,omitempty"`
	//TargetPort int      `json:"target_port,omitempty"`

	process *UpstreamProcess
}

// Provision implements caddy.Provisioner
func (a *ThrusterApp) Provision(ctx caddy.Context) error {
	if len(a.Command) == 0 {
		return fmt.Errorf("there is not any command")
	}

	//if a.TargetPort == 0 {
	//	a.TargetPort = 3000
	//}
	//
	//os.Setenv("PORT", fmt.Sprintf("%d", a.TargetPort))

	a.process = NewUpstreamProcess(a.Command[0], a.Command[1:], false)

	return nil
}

// Start starts the app.
func (a ThrusterApp) Start() error {
	_, err := a.process.Run()

	return err
}

// Stop stops the app.
func (a *ThrusterApp) Stop() error {
	return a.process.Stop()
}

// CaddyModule implements caddy.ModuleInfo
func (a ThrusterApp) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "thruster",
		New: func() caddy.Module { return new(ThrusterApp) },
	}
}

func parseGlobalCaddyfileBlock(d *caddyfile.Dispenser, _ interface{}) (interface{}, error) {
	var thruster ThrusterApp

	if d.NextArg() && d.NextArg() {
		thruster.Command = append([]string{d.Val()}, d.RemainingArgs()...)
	}

	for d.NextBlock(0) {
		switch d.Val() {
		case "command":
			thruster.Command = d.RemainingArgs()
			//case "target_port":
			//	if d.NextArg() {
			//		targetPort, err := strconv.Atoi(d.Val())
			//		if err != nil {
			//			return nil, d.Errf("parsing 'target_port': %v", err)
			//		}
			//		thruster.TargetPort = targetPort
			//	} else {
			//		return nil, d.ArgErr()
			//	}
		}
	}

	return httpcaddyfile.App{
		Name:  "thruster",
		Value: caddyconfig.JSON(thruster, nil),
	}, nil
}

var (
	_ caddy.App         = (*ThrusterApp)(nil)
	_ caddy.Module      = (*ThrusterApp)(nil)
	_ caddy.Provisioner = (*ThrusterApp)(nil)
)
