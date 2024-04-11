package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/prog_supdex/caddy_thruster/modules/proxy_runner"
)

func main() {
	caddycmd.Main()
}
