# Caddy Rails

CaddyRails is a reverse proxy module for Ruby on Rails designed to integrate with Caddy, facilitating features like reverse proxying, automatic HTTPS, compression, and more.

## Features

- **Reverse Proxy:** Simplifies forwarding requests to your application.
- **Automatic HTTPS:** Automatically manages SSL/TLS certificates. (Currently in progress)
- **Compression:** Supports Gzip, and Zstd for reducing data transfer sizes.
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

### Using Default Rails Command

To start CaddyRails with the default Rails server command, simply run the following command. This will automatically find the `bin/rails` file and run the Rails server using the server command:
```bash
./caddy_rails serve --https_port 8443 --http_port 8012 --target_port 3000
```

### Specifying a Custom Command

If you need to specify a custom command or additional arguments, you can provide them directly to `serve`:

```bash
./caddy_rails serve bin/rails s --https_port 8443 --http_port 8012 --target_port 3000
```

By default, serve will locate the bin/rails file and run the Rails server using the server command. If no custom command is provided, this is the default behavior.

### Error Handling

If no custom command is provided and the `bin/rails` file does not exist, the command will raise an error indicating that the required Rails executable is missing.

### Command Line Arguments
- `--target-port`: The port that your server should run on.  caddy-server will set this value to the PORT environment variable. Default: `3000`.
- `--http-port`: The port to listen on for HTTP traffic. Default: `80`.
- `--https-port`: The port to listen on for HTTPS traffic. Default: `443`.
- `-l, --listen`: The address to which to bind the listener. Default: `localhost`.
- `--ssl-domain`: The domain name to use for SSL provisioning. If not set, SSL will be disabled.
- `-v, --debug`: Enable verbose debug logs.
- `--access-log`: Enable the access log. Default: `true`.
- `--no-compress`: Disable Zstandard and Gzip compression
- `--http-idle-timeout`: The maximum time a client can be idle before the connection is closed. Default: `60s`.
- `--http-read-timeout`: The maximum time a client can take to send the request headers. Default: `30s`.
- `--http-write-timeout`: The maximum time during which the client must read the response. Default: `30s`.
- `--pid-file`: Path to the PID file to control an existing process. Default is `tmp/pids/server.pid`
- `--phased-restart`: Using for hot reloading the existing process. Default is false
- `--server-type`: Using for restarting the existing process. Puma and Unicorn have different ways for hot reloading
- `--stop`: Using for stopping the existing process
- `--anycable-enabled`: Enable AnyCable. Default is false

## Configuration File Generation
Generate a customized Caddyfile by running:

```bash
./caddy_rails config-init --folder_path "./config" --https_enable --ssl-domain localhost
```

This command creates a Caddyfile in the specified directory, tailoring it with options for SSL, compression, and logging based on provided parameters.

### Command Line Arguments
- `--folder-path`: Directory to generate the Caddyfile in. Defaults to the current directory
- `--http-host`: Host address for the HTTP server. Default: `localhost`
- `--http-port`: The port for HTTP traffic. Default: `80`
- `--https-port`: The port for HTTP traffic. Default: `443`
- `--enable-debug`: Enable verbose debug logs
- `--access-log`: Enable the access log. Default: `true`
- `--ssl-domain`: The domain name for SSL. If empty, SSL is disabled
- `--backend-port`: THe port that the backend service listens on. Default is `8080`
- `--compression-enable`: Enable response compression using gzip and zstd
- `--anycable-enable`: Enable anycable

### Running the application

After the generation the `Caddyfile` you can run project by this command

```bash
./caddy_rails run
```

**Important:** The caddy_rails can not have enough permissions for ports 80 and 443.
So you can change these ports by `--http-port` and `--https-port`, or run it via sudo or use `setcap`

## Managing Application Lifecycle
You can manage the running Rails application in another console session using:

#### Phased/Hot Restart
- Stopping the Server: `./caddy_rails serve --stop`
- Restarting the Server: `./caddy_rails serve --phased-restart`
- Phased Restart for Puma or Hot Restart for Unicorn:
```bash
./caddy_rails serve --phased_restart --server-type puma
./caddy_rails serve --phased_restart --server-type unicorn 
```

These commands facilitate seamless application management, ensuring minimal downtime and flexible maintenance operations.
