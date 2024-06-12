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

	encodeRoute := caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(encode.Encode{
			EncodingsRaw: caddy.ModuleMap{
				"zstd": caddyconfig.JSON(zstd.New(), nil),
				"gzip": caddyconfig.JSON(gzip.New(), nil),
			},
			Prefer: []string{"zstd", "gzip"},
		}, "handler", "encode", nil)},
	}

	return encodeRoute
}

func createAnyCableRoute() caddyhttp.Route {
	anyCableHandler := caddy_anycable.AnyCableHandler{}
	anyCableOptions := os.Getenv("ANYCABLE_OPT")
	if anyCableOptions != "" {
		options := strings.Split(anyCableOptions, " ")
		anyCableHandler.Options = options
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
	reverseProxyHandler := reverseproxy.Handler{
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
	}

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(reverseProxyHandler, "handler", "reverse_proxy", nil)},
	}
}

func createFileServerRoute(fs cmd.Flags) caddyhttp.Route {
	fileServerHandler := fileserver.FileServer{
		Root: fs.String("files-path"),
	}

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(fileServerHandler, "handler", "file_server", nil)},
	}
}

func createHTTPApp(fs cmd.Flags, route caddyhttp.Route) caddyhttp.App {
	var actualPort int

	httpPort := fs.Int("http-port")
	httpsPort := fs.Int("https-port")

	if fs.String("ssl-domain") == "" {
		actualPort = httpPort
	} else {
		actualPort = httpsPort
	}

	httpServer := createServer(actualPort, caddyhttp.RouteList{route}, fs)

	return caddyhttp.App{
		HTTPPort:  httpPort,
		HTTPSPort: httpsPort,
		Servers: map[string]*caddyhttp.Server{
			"rails": httpServer,
		},
	}
}

func createServer(port int, routes caddyhttp.RouteList, fs cmd.Flags) *caddyhttp.Server {
	server := &caddyhttp.Server{
		Listen:       []string{fmt.Sprintf(":%d", port)},
		Routes:       routes,
		IdleTimeout:  caddy.Duration(fs.Duration("http-idle-timeout")),
		ReadTimeout:  caddy.Duration(fs.Duration("http-read-timeout")),
		WriteTimeout: caddy.Duration(fs.Duration("http-write-timeout")),
	}
	if fs.Bool("access-log") {
		server.Logs = &caddyhttp.ServerLogConfig{}
	}
	return server
}

func createCaddyConfig(httpApp caddyhttp.App, debug bool) *caddy.Config {
	cfg := &caddy.Config{
		Admin: &caddy.AdminConfig{
			Disabled: true,
			Config: &caddy.ConfigSettings{
				Persist: new(bool),
			},
		},
		AppsRaw: caddy.ModuleMap{"http": caddyconfig.JSON(httpApp, nil)},
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

	routes = append(routes,
		createFileServerRoute(fs),
	)

	subroute := caddyhttp.Subroute{Routes: routes}

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(subroute, "handler", "subroute", nil)},
	}
}
