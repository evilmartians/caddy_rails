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
	"go.uber.org/zap"
	"log"
)

func StartCaddyReverseProxy(fs cmd.Flags) error {
	route := createGroupedRoutes(fs)
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

	br, err := caddy.GetModule("http.encoders.br")
	if err != nil {
		log.Fatalf("Failed to load br module: %v", err)
	}

	encodeRoute := caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(encode.Encode{
			EncodingsRaw: caddy.ModuleMap{
				"zstd": caddyconfig.JSON(zstd.New(), nil),
				"gzip": caddyconfig.JSON(gzip.New(), nil),
				"br":   caddyconfig.JSON(br.New(), nil),
			},
			Prefer: []string{"zstd", "br", "gzip"},
		}, "handler", "encode", nil)},
	}

	return encodeRoute
}

func createReverseProxyRoute(fs cmd.Flags) caddyhttp.Route {
	reverseProxyHandler := reverseproxy.Handler{
		TransportRaw: caddyconfig.JSONModuleObject(reverseproxy.HTTPTransport{}, "protocol", "http", nil),
		Upstreams: reverseproxy.UpstreamPool{
			{Dial: fmt.Sprintf("%s:%s", fs.String("listen"), fs.String("target_port"))},
		},
		Headers: &headers.Handler{
			Request: &headers.HeaderOps{
				Set: map[string][]string{"Host": {"{http.reverse_proxy.upstream.hostport}"}},
			},
		},
	}
	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(reverseProxyHandler, "handler", "reverse_proxy", nil)},
	}
}

func createFileServerRoute(fs cmd.Flags) caddyhttp.Route {
	fileServerHandler := fileserver.FileServer{
		Root: fs.String("files_path"),
	}

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(fileServerHandler, "handler", "file_server", nil)},
	}
}

func createHTTPApp(fs cmd.Flags, route caddyhttp.Route) caddyhttp.App {
	httpServer := createServer(fs.String("http_port"), caddyhttp.RouteList{route}, fs)
	httpsServer := createServer(fs.String("https_port"), caddyhttp.RouteList{route}, fs)

	return caddyhttp.App{
		Servers: map[string]*caddyhttp.Server{
			"http_server":  httpServer,
			"https_server": httpsServer,
		},
	}
}

func createServer(port string, routes caddyhttp.RouteList, fs cmd.Flags) *caddyhttp.Server {
	server := &caddyhttp.Server{
		Listen:       []string{fmt.Sprintf(":%s", port)},
		Routes:       routes,
		IdleTimeout:  caddy.Duration(fs.Duration("http_idle_timeout")),
		ReadTimeout:  caddy.Duration(fs.Duration("http_read_timeout")),
		WriteTimeout: caddy.Duration(fs.Duration("http_write_timeout")),
	}
	if fs.Bool("access_log") {
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

	if !fs.Bool("no-compress") {
		routes = append(routes, createEncodeRoute())
	}

	routes = append(routes,
		createReverseProxyRoute(fs),
		createFileServerRoute(fs),
	)

	subroute := caddyhttp.Subroute{Routes: routes}

	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(subroute, "handler", "subroute", nil)},
	}
}
