# wiretap


A local and pipeline based tool to sniff API request and responses from clients and servers
to detect OpenAPI contract violations and compliance.

A shift left tool, for those who want to know if their applications
are actually compliant with an API.

This is an early tool and in active development.

Probably best to leave this one alone for now, come back later
when it's a little more baked.


![](https://github.com/pb33f/wiretap/blob/main/.github/assets/wiretap-preview.gif)

## Configuring paths & rewriting them. 

Provide a configuration for path rewriting via the wiretap config.

This uses the same syntax for path rewriting as the [http-proxy-middleware](https://github.com/chimurai/http-proxy-middleware)

Paths can be matched by globs and then individual segments can be matched and re-written.

```yaml
paths:
  /pb33f/*/test/**:
    target: localhost:80
    pathRewrite:
      '^/pb33f/(\w+)/test/': ''
```

## Dropping certain headers

To prevent certain headers from being proxies, you can drop them using the `headers` config property and the `drop` property
which is an array of headers to drop from all outbound requests..

```yaml
headers:
  drop:
    - Origin
```    

## Configuring static paths

Running a single page application? Need certain paths to always be caught and forwarded to your SPA? Configure
static paths to be caught and forwarded to your SPA.

```yaml
staticDir: ui/dist
staticIndex: index.html
staticPaths:
  - /my-app/*
  - /another-app/somewhere/*
```

## Command Line Interface

### Available Flags

| Shortcut | Flag             | Description                                                                                |
|----------|------------------|--------------------------------------------------------------------------------------------|
| `-u`     | `--url`          | Redirect URL for wiretap to send traffic to.                                               |
| `-s`     | `--spec`         | Path to the OpenAPI Specification to use.                                                  |
| `-p`     | `--port`         | Port on which to listen for API traffic. (default is `9090`)                               |
| `-m`     | `--monitor-port` | Port on which to serve the monitor UI. (default is `9091`)                                 |
| `-d`     | `--delay`        | Set a global delay for all API requests in milliseconds. (default is `0`)                  |
| `-c`     | `--config`       | Location of wiretap configuration file to use (default is `.wiretap` in current directory) |
| `-t`     | `--static`       | Location of static files to serve along with API requests (simulate real app deployment)   |
| `-i`     | `--static-index` | Index file to serve for root static requests and all static paths (default is index.html)  |
