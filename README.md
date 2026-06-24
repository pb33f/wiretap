# wiretap

![logo](.github/wiretap-hero.png)

[![discord](https://img.shields.io/discord/923258363540815912)](https://discord.gg/x7VACVuEGP)
[![GitHub downloads](https://img.shields.io/github/downloads/pb33f/wiretap/total?label=github%20downloads&style=flat-square)](https://github.com/pb33f/wiretap/releases)
[![npm](https://img.shields.io/npm/dm/@pb33f/wiretap?style=flat-square&label=npm%20downloads)](https://www.npmjs.com/package/@pb33f/wiretap)
[![Docker Pulls](https://img.shields.io/docker/pulls/pb33f/wiretap?style=flat-square)](https://hub.docker.com/r/pb33f/wiretap)

`wiretap` is an API diagnostics, validation, proxying, and mocking tool for teams building against OpenAPI contracts.

Run it locally or in a pipeline to inspect real HTTP traffic, validate requests and responses, serve static frontends, replay HAR files, and simulate APIs from an OpenAPI description before the backend exists.

> This is an early tool and in active development. Try it out and tell us what breaks, what helps, and what you want next.

![](https://github.com/pb33f/wiretap/blob/main/.github/assets/wiretap-preview.gif)

## Quick start

[Quick Start Guide](https://pb33f.io/wiretap/quickstart/)

## Install

### Homebrew

```shell
brew install pb33f/taps/wiretap
```

### npm

```shell
npm install -g @pb33f/wiretap
```

### cURL

```shell
curl -fsSL https://pb33f.io/wiretap/install.sh | sh
```

### Docker

Images are published to both Docker Hub and GitHub Container Registry:

```shell
docker pull pb33f/wiretap
docker pull ghcr.io/pb33f/wiretap
```

```shell
docker run -p 9090:9090 -p 9091:9091 -p 9092:9092 --rm -v \
  $PWD:/work:rw pb33f/wiretap -u https://somehostoutthere.com
docker run -p 9090:9090 -p 9091:9091 -p 9092:9092 --rm -v \
  $PWD:/work:rw ghcr.io/pb33f/wiretap -u https://somehostoutthere.com
```

The default ports are `9090` for the API gateway, `9091` for the monitor UI, and `9092` for the monitor websocket.

### Windows

Download the Windows build for your CPU from the [latest GitHub release](https://github.com/pb33f/wiretap/releases).

## Run as a validation proxy

Send traffic through `wiretap` and forward it to a real upstream API:

```shell
wiretap -u https://api.pb33f.com
```

Add an OpenAPI contract to validate traffic:

```shell
wiretap -u https://api.pb33f.com -s ./openapi.yaml
```

Your client calls `http://localhost:9090`; `wiretap` validates and proxies the request, validates the response, and streams everything to the monitor UI.

## Mock APIs from OpenAPI

`wiretap` is not only a passive proxy. It can run as a contract-driven API simulator.

Use full mock mode when the backend does not exist, is unstable, or should not be called during client tests:

```shell
wiretap -s ./openapi.yaml --mock-mode
```

In mock mode, no traffic is sent upstream. `wiretap` matches the incoming request to the OpenAPI operation, validates the request, chooses an appropriate response, and generates the response body from schemas and examples.

Mocking capabilities include:

- Generate all responses from an OpenAPI contract with `--mock-mode`.
- Mock only selected paths with `mockModeList` in `wiretap.yaml`, while other paths still proxy to an upstream API.
- Select named OpenAPI examples by sending a `Preferred: ExampleName` request header.
- Override the returned status code with `wiretap-status-code: 404` after a successful mock body is generated.
- Use `--hard-validation` to reject invalid mock requests, or `--mock-bypass-validation` / `wiretap-bypass-validation: true` to keep serving chosen examples for malformed fixture traffic.
- Enable `--strict-mode` to detect undeclared properties, parameters, headers, and cookies.
- Return non-JSON examples such as `text/plain` or `text/html` without JSON string wrapping.
- Run with no upstream URL in full mock mode; an OpenAPI contract is the required input.

Path-scoped mocking example:

```yaml
redirectURL: https://api.pb33f.com
contract: ./openapi.yaml
mockModeList:
  - /checkout/**
  - /beta/*
```

```shell
wiretap -c ./wiretap.yaml
```

Requests under `/checkout/**` and `/beta/*` return generated mocks. Everything else is proxied and validated against `https://api.pb33f.com`.

## Static mock fixtures

For exact fixture-based mocks, point `wiretap` at a static mock directory:

```shell
wiretap -u https://api.pb33f.com --static-mock-dir ./mocks
```

Directory layout:

```text
./mocks/
  mock-definitions/
    get-product.json
  body-jsons/
    product.json
```

Definition example:

```json
{
  "request": {
    "method": "GET",
    "urlPath": "/products/123",
    "queryParams": {
      "view": "full"
    }
  },
  "response": {
    "statusCode": 200,
    "header": {
      "x-mock": "wiretap"
    },
    "bodyJsonFilename": "product.json"
  }
}
```

Static mocks can match method, host, path, headers, query parameters, and JSON or form request bodies. Response bodies can be inline or loaded from `body-jsons/`, and can use request data with template expressions such as `${queryParams.view}`.

If no static mock matches, `wiretap` falls back to the normal proxy/mock handler. Mock definition JSON files are watched and reloaded while the service is running.

## Multi-spec mode

Got more than one OpenAPI contract? Hand `wiretap` a list, or point it at a directory and let it discover them.

```shell
wiretap -u https://api.pb33f.com --specs ./users.yaml,./orders.yaml,./billing.yaml
```

```shell
wiretap -u https://api.pb33f.com --spec-dir ./contracts
```

`wiretap` routes each request to the spec that owns it and reports duplicate or ambiguous routes across contracts at startup.

Use `--dry-run` to discover, analyze, and print the conflict report without starting the proxy. It exits non-zero on conflicts or load errors, which makes it useful in CI.

```shell
wiretap --spec-dir ./contracts --dry-run
```

Full details: [Multi-spec mode](https://pb33f.io/wiretap/multi-spec-mode/).

## HAR and headless workflows

Load a HAR file into the monitor:

```shell
wiretap --har ./traffic.har
```

Validate a HAR file against OpenAPI without running the proxy:

```shell
wiretap --har ./traffic.har --har-validate -s ./openapi.yaml --har-allow /api
```

Stream validation violations to a report file while `wiretap` runs:

```shell
wiretap -u https://api.pb33f.com -s ./openapi.yaml \
  --stream-report \
  --report-filename ./wiretap-report.jsonl
```

## Documentation

- [Quick Start](https://pb33f.io/wiretap/quickstart/)
- [Installing](https://pb33f.io/wiretap/installing/)
- [Configuring](https://pb33f.io/wiretap/configuring/)
- [Mock responses](https://pb33f.io/wiretap/mock-responses/)
- [Multi-spec mode](https://pb33f.io/wiretap/multi-spec-mode/)
- [Monitor UI](https://pb33f.io/wiretap/monitor/)
- [Serving static content](https://pb33f.io/wiretap/static-content/)
- [HAR files](https://pb33f.io/wiretap/har-files/)
- [Headless mode](https://pb33f.io/wiretap/headless-mode/)
- [GiftShop example API](https://pb33f.io/wiretap/giftshop-api/)
- [Contributing](https://pb33f.io/wiretap/contributing/)
