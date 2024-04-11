package proxy_runner

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
)

func startCaddyReverseProxy(fs cmd.Flags) error {
	routes := createRoutes(fs)
	httpApp := createHTTPApp(fs, routes)

	cfg := createCaddyConfig(httpApp, fs.Bool("debug"))
	return caddy.Run(cfg)
}

func createRoutes(fs cmd.Flags) caddyhttp.RouteList {
	routes := caddyhttp.RouteList{}
	if !fs.Bool("no-compress") {
		routes = append(routes, createEncodeRoute())
	}
	routes = append(routes, createReverseProxyRoute(fs))
	//routes = append(routes, createFileServerRoute(fs))

	return routes
}

func createEncodeRoute() caddyhttp.Route {
	zstd, _ := caddy.GetModule("http.encoders.zstd")
	gzip, _ := caddy.GetModule("http.encoders.gzip")

	encodeHandler := encode.Encode{
		EncodingsRaw: caddy.ModuleMap{
			"gzip": caddyconfig.JSON(gzip.New(), nil),
			"zstd": caddyconfig.JSON(zstd.New(), nil),
		},
		Prefer: []string{"zstd", "gzip"},
	}
	return caddyhttp.Route{
		HandlersRaw: []json.RawMessage{caddyconfig.JSONModuleObject(encodeHandler, "handler", "encode", nil)},
	}
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

func createHTTPApp(fs cmd.Flags, routes caddyhttp.RouteList) caddyhttp.App {
	httpServer := createServer(fs.String("http_port"), routes, fs)
	httpsServer := createServer(fs.String("https_port"), routes, fs)

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
