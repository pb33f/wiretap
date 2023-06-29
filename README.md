# wiretap

A local and pipeline based tool to sniff API request and responses from clients and servers
to detect OpenAPI contract violations and compliance.

A shift left tool, for those who want to know if their applications
are actually compliant with an API.

This is an early tool and in active development.

Probably best to leave this one alone for now, come back later
when it's a little more baked.

## Command Line Interface

### Available Flags

| Shortcut | Flag             | Description                                                                                |
| -------- | ---------------- | ------------------------------------------------------------------------------------------ |
| `-u`     | `--url`          | Redirect URL for wiretap to send traffic to.                                               |
| `-s`     | `--spec`         | Path to the OpenAPI Specification to use.                                                  |
| `-p`     | `--port`         | Port on which to listen for API traffic. (default is `9090`)                               |
| `-m`     | `--monitor-port` | Port on which to serve the monitor UI. (default is `9091`)                                 |
| `-d`     | `--delay`        | Set a global delay for all API requests in milliseconds. (default is `0`)                  |
| `-c`     | `--config`       | Location of wiretap configuration file to use (default is `.wiretap` in current directory) |
