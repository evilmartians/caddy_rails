package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/evilmartians/caddy_thruster/internal/app"
	_ "github.com/evilmartians/caddy_thruster/internal/commands"
)

func main() {
	caddycmd.Main()
}
