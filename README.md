# Caddy Thruster

Caddy Thruster is a reverse proxy module designed to integrate seamlessly with Caddy, facilitating features like reverse proxying, automatic HTTPS, compression, and more.

## Features

- **Reverse Proxy:** Simplifies forwarding requests to your application.
- **Automatic HTTPS:** Automatically manages SSL/TLS certificates. (Currently in progress)
- **Compression:** Supports Gzip and Zstd for reducing data transfer sizes.
- **Access Logging:** Enables detailed logging of incoming requests.
- **Connection Timeouts:** Customizable read, write, and idle timeouts for connections.
- **Debugging:** Provides extensive debug logs to troubleshoot issues.

## Installation

To install Caddy Thruster, navigate to the project directory and compile the module with Caddy:

```bash
xcaddy build --with github.com/evilmartians/caddy_thruster
```

After compilation, copy the `caddy` binary to your Rails project directory.

## Usage

### Command Line Interface
Run Caddy Thruster directly from the command line by specifying your target server command and its arguments. 
Caddy Thruster sets up a reverse proxy to your application.

### Command Line Arguments

- `--target_port <port>`: The port that your server should run on.  ProxyRunner will set the PORT environment variable to this value. Default: `3000`.
- `--http_port <port>`: The port to listen on for HTTP traffic. Default: `80`.
- `--https_port <port>`: The port to listen on for HTTPS traffic. Default: `443`.
- `-l, --listen <address>`: The address to which to bind the listener. Default: `localhost`.
- `--ssl_domain <domain>`: The domain name to use for SSL provisioning. If not set, SSL will be disabled. (in progress)
- `-v, --debug`: Enable verbose debug logs.
- `--access_log <bool>`: Enable the access log. Default: `true`.
- `--no-compress`: Disable Zstandard and Gzip compression.
- `--http_idle_timeout <duration>`: The maximum time a client can be idle before the connection is closed. Default: `60s`.
- `--http_read_timeout <duration>`: The maximum time a client can take to send the request headers. Default: `30s`.
- `--http_write_timeout <duration>`: The maximum time during which the client must read the response. Default: `30s`.

### Examples

```bash
./caddy thruster bin/rails s --https_port 8443 --http_port 8012 --target_port 3000
```

### Configuration via Caddyfile

Configure Caddy Thruster using the Caddyfile as follows:

```caddyfile
{
  thruster bin/rails s -p {$CADDY_BACKEND_PORT}
}

http://{$CADDY_HOST}:{$CADDY_PORT} {
  root * ./public
  @notStatic {
    not {
      file {
        try_files {path}
      }
    }
  }

  encode gzip zstd

  reverse_proxy @notStatic {
    to localhost:{$CADDY_BACKEND_PORT}

    header_up X-Real-IP {remote_host}
    header_up X-Forwarded-Proto {scheme}
    header_up Access-Control-Allow-Origin *
    header_up Access-Control-Allow-Credentials true
    header_up Access-Control-Allow-Headers Cache-Control,Content-Type
    transport http {
      read_buffer 8192
    }
  }

  file_server
}
```

Run the command with environment variables:

```bash
    CADDY_HOST=localhost CADDY_PORT=3000 CADDY_BACKEND_PORT=4000 ./caddy run
```

