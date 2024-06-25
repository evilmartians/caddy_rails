package utils

import (
	"encoding/json"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig"
	cmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/encode"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/headers"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/reverseproxy"
	"github.com/evilmartians/caddy_anycable"
	"github.com/evilmartians/caddy_rails/internal/app"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
)

func StartCaddyReverseProxy(fs cmd.Flags) error {
	route := createGroupedRoutes(fs)

	if fs.String("ssl-domain") != "" {
		route.MatcherSetsRaw = []caddy.ModuleMap{
			{
				"host": caddyconfig.JSON(caddyhttp.MatchHost{fs.String("ssl-domain")}, nil),
			},
		}
	}

	httpApp := createHTTPApp(fs, route)

	cfg := createCaddyConfig(httpApp, fs.Bool("debug"))
	return caddy.Run(cfg)
}

func createEncodeRoute() caddyhttp.Route {
	gzip, err := caddy.GetModule("http.encoders.gzip")
	if err != nil {
		log.Fatalf("Failed to load gzip module: %v", err)
	}

	zstd, err := caddy.GetModule("http.encoders.zstd")
	if err != nil {
		log.Fatalf("Failed to load zstd module: %v", err)
	}

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(encode.Encode{
			EncodingsRaw: caddy.ModuleMap{
				"zstd": caddyconfig.JSON(zstd.New(), nil),
				"gzip": caddyconfig.JSON(gzip.New(), nil),
			},
			Prefer: []string{"zstd", "gzip"},
		}, "handler", "encode", nil)},
	}
}

func createAnyCableRoute() caddyhttp.Route {
	anyCableHandler := caddy_anycable.AnyCableHandler{}
	anyCableOptions := os.Getenv("ANYCABLE_OPT")
	if anyCableOptions != "" {
		anyCableHandler.Options = strings.Split(anyCableOptions, " ")
	}

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(anyCableHandler, "handler", "anycable", nil)},
		MatcherSetsRaw: []caddy.ModuleMap{
			{
				"path": caddyconfig.JSON(caddyhttp.MatchPath{"/cable*"}, nil),
			},
		},
	}
}

func createReverseProxyRoute(fs cmd.Flags) caddyhttp.Route {
	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(reverseproxy.Handler{
			Upstreams: reverseproxy.UpstreamPool{
				{Dial: fmt.Sprintf("%s:%d", fs.String("listen"), fs.Int("target-port"))},
			},
			Headers: &headers.Handler{
				Request: &headers.HeaderOps{
					Set: map[string][]string{
						"Host": {"{http.reverse_proxy.upstream.hostport}"},
					},
				},
			},
		}, "handler", "reverse_proxy", nil)},
	}
}

func createFileServerRoute(fs cmd.Flags) caddyhttp.Route {
	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(fileserver.FileServer{
			Root: fs.String("files-path"),
		}, "handler", "file_server", nil)},
	}
}

func createHTTPApp(fs cmd.Flags, route caddyhttp.Route) caddyhttp.App {
	port := fs.Int("https-port")
	if fs.String("ssl-domain") == "" {
		port = fs.Int("http-port")
	}

	return caddyhttp.App{
		HTTPPort:  fs.Int("http-port"),
		HTTPSPort: fs.Int("https-port"),
		Servers: map[string]*caddyhttp.Server{
			"rails": {
				Listen:       []string{fmt.Sprintf(":%d", port)},
				Routes:       caddyhttp.RouteList{route},
				IdleTimeout:  caddy.Duration(fs.Duration("http-idle-timeout")),
				ReadTimeout:  caddy.Duration(fs.Duration("http-read-timeout")),
				WriteTimeout: caddy.Duration(fs.Duration("http-write-timeout")),
				Logs:         getServerLogConfig(fs.Bool("access-log")),
			},
		},
	}
}

func getServerLogConfig(enabled bool) *caddyhttp.ServerLogConfig {
	if enabled {
		return &caddyhttp.ServerLogConfig{}
	}
	return nil
}

func createCaddyConfig(httpApp caddyhttp.App, debug bool) *caddy.Config {
	cfg := &caddy.Config{
		Admin: &caddy.AdminConfig{
			Disabled: true,
			Config: &caddy.ConfigSettings{
				Persist: new(bool),
			},
		},
		AppsRaw: caddy.ModuleMap{
			"http":  caddyconfig.JSON(httpApp, nil),
			"serve": caddyconfig.JSON(app.CaddyRailsApp{}, nil),
		},
	}
	if debug {
		cfg.Logging = &caddy.Logging{
			Logs: map[string]*caddy.CustomLog{"default": {BaseLog: caddy.BaseLog{Level: zap.DebugLevel.CapitalString()}}},
		}
	}
	return cfg
}

func createGroupedRoutes(fs cmd.Flags) caddyhttp.Route {
	routes := caddyhttp.RouteList{}

	if fs.Bool("anycable-enabled") {
		routes = append(routes, createAnyCableRoute())
	}

	routes = append(routes, createReverseProxyRoute(fs))

	if !fs.Bool("no-compress") {
		routes = append(routes, createEncodeRoute())
	}

	routes = append(routes, createFileServerRoute(fs))

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(caddyhttp.Subroute{Routes: routes}, "handler", "subroute", nil)},
	}
}
