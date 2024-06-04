package main

import (
	"flag"
	"fmt"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	_ "github.com/evilmartians/caddy_anycable"
	_ "github.com/evilmartians/caddy_rails/internal/app"
	_ "github.com/evilmartians/caddy_rails/internal/commands"
	"github.com/evilmartians/caddy_rails/version"
	"os"
)

func main() {
	versionFlag := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	caddycmd.Main()
}
