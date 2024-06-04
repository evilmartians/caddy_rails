package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/evilmartians/caddy_anycable"
	_ "github.com/evilmartians/caddy_rails/internal/app"
	_ "github.com/evilmartians/caddy_rails/internal/commands"
)

func main() {
	caddycmd.Main()
}
