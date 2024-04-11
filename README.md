# Caddy Thruster

Caddy Thruster is a reverse proxy tool. Something like a Thruster

## Features

- **Reverse Proxy:** Easily forward requests to your application
- **Automatic HTTPS:** (Not yet, It Needs to realize generating certs like Thruster)
- **Compression:** Supports Gzip and Zstd compression
- **Access Logging:** Keep track of incoming requests with optional access logging.
- **Timeouts:** Configure read, write, and idle timeouts for connections.
- **Debugging:** Verbose debug logs.

## Installation

Navigate to the Caddy Thruster project directory, then compile the project

```bash
go build
```

After compilation, copy the caddy_thruster binary to your Rails project directory.

## Usage

To run Caddy Thruster, use the `proxy-runner` command followed by the target server command and any necessary arguments. Caddy Thruster will set up a reverse proxy to your application, handling HTTP(S) traffic according to the provided configuration.

### Command Line Arguments

- `--target_port <port>`: The port that your server should run on.  ProxyRunner will set the PORT environment variable to this value. Default: `3000`.
- `--http_port <port>`: The port to listen on for HTTP traffic. Default: `80`.
- `--https_port <port>`: The port to listen on for HTTPS traffic. Default: `443`.
- `-l, --listen <address>`: The address to which to bind the listener. Default: `localhost`.
- `--ssl_domain <domain>`: The domain name to use for SSL provisioning. If not set, SSL will be disabled. (Also in progress)
- `-v, --debug`: Enable verbose debug logs.
- `--access_log <bool>`: Enable the access log. Default: `true`.
- `--no-compress`: Disable Zstandard and Gzip compression.
- `--http_idle_timeout <duration>`: The maximum time a client can be idle before the connection is closed. Default: `60s`.
- `--http_read_timeout <duration>`: The maximum time a client can take to send the request headers. Default: `30s`.
- `--http_write_timeout <duration>`: The maximum time during which the client must read the response. Default: `30s`.

### Examples

```bash
./caddy_thruster proxy-runner bin/rails s --https_port 8443 --http_port 8012 --target_port 3001 
```
