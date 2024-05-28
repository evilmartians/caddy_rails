# Caddy Thruster

CaddyRails is a reverse proxy module for Ruby on Rails designed to integrate with Caddy, facilitating features like reverse proxying, automatic HTTPS, compression, and more.

## Features

- **Reverse Proxy:** Simplifies forwarding requests to your application.
- **Automatic HTTPS:** Automatically manages SSL/TLS certificates. (Currently in progress)
- **Compression:** Supports Gzip, Brotli and Zstd for reducing data transfer sizes.
- **Access Logging:** Enables detailed logging of incoming requests.
- **Connection Timeouts:** Customizable read, write, and idle timeouts for connections.
- **Debugging:** Provides extensive debug logs to troubleshoot issues.

## Installation

Install CaddyRails by navigating to your project directory and compiling the module with the following command:

```bash
make build
```

Once compiled, ensure that the `caddy_rails` binary is located within your Rails project directory for easy access.

## Usage Instructions

## Starting the Server
Initiate CaddyRails directly from the command line by specifying your Rails server command alongside necessary arguments. 
The tool sets up a reverse proxy automatically.

```bash
./caddy_rails serve-rails bin/rails s --https_port 8443 --http_port 8012 --target_port 3000
```

### Command Line Arguments
- `--target_port <port>`: The port that your server should run on.  ProxyRunner will set the PORT environment variable to this value. Default: `3000`.
- `--http_port <port>`: The port to listen on for HTTP traffic. Default: `80`.
- `--https_port <port>`: The port to listen on for HTTPS traffic. Default: `443`.
- `-l, --listen <address>`: The address to which to bind the listener. Default: `localhost`.
- `--ssl_domain <domain>`: The domain name to use for SSL provisioning. If not set, SSL will be disabled. (in progress)
- `-v, --debug`: Enable verbose debug logs.
- `--access_log <bool>`: Enable the access log. Default: `true`.
- `--no-compress`: Disable Brotli, Zstandard and Gzip compression
- `--http_idle_timeout <duration>`: The maximum time a client can be idle before the connection is closed. Default: `60s`.
- `--http_read_timeout <duration>`: The maximum time a client can take to send the request headers. Default: `30s`.
- `--http_write_timeout <duration>`: The maximum time during which the client must read the response. Default: `30s`.

## Configuration File Generation
Generate a customized Caddyfile by running:

```bash
./caddy_rails config-init --folder_path "./config" --http_host "myapp.local" --https_enable
```

This command creates a Caddyfile in the specified directory, tailoring it with options for SSL, compression, and logging based on provided parameters.

### Command Line Arguments
- `--folder_path <string>`: Directory to generate the Caddyfile in. Defaults to the current directory
- `--http_host <string>`: Host address for the HTTP server. Default: `localhost`
- `--http_port <string>`: The port for HTTP traffic. Default: `80`
- `--https_port <string>`: The port for HTTP traffic. Default: `443`
- `--enable_debug <bool>`: Enable verbose debug logs
- `--access_log <bool>`: Enable the access log. Default: `true`
- `--ssl_domain <string>`: The domain name for SSL. If empty, SSL is disabled
- `--backend_port <string>`: THe port that the backend service listens on. Default is `8080`
- `--https_enable <bool>`: Enable HTTPS configuration. Default `false`
- `--enable_compression`: Enable response compression using gzip, brotli and zstd

### Running the application

After the generation the `Caddyfile` you can run project by this command

```bash
    ./caddy_rails run
```

**important:** The caddy_rails can do not have enough permissions for ports 80 and 443 on locally.
So you can change these ports by `--http_port` and `--https_port` 

## Managing Application Lifecycle
You can manage the running Rails application in another console session using:

#### Phased/Hot Restart

- Stopping the Server: `./caddy_rails serve-rails --stop`
- Restarting the Server: `./caddy_rails serve-rails --restart`
- Phased Restart for Puma or Hot Restart for Unicorn:
```bash
./caddy_rails serve-rails --phased_restart --server-type puma
./caddy_rails serve-rails --phased_restart --server-type unicorn 
```

These commands facilitate seamless application management, ensuring minimal downtime and flexible maintenance operations.
