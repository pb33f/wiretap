# AGENTS.md

This guide is for any automated agent working in this repository. It explains how to operate the `wiretap` CLI from the current codebase and what to check before changing CLI behavior or documentation.

## Source of truth

- CLI flags are registered in `cmd/root_command.go`.
- YAML configuration fields are defined in `shared/config.go`.
- Startup wiring is in `cmd/run_service.go`, `cmd/handle_http_traffic.go`, and `cmd/booted_message.go`.
- Live proxy and mock routing decisions are in `daemon/handle_request.go` and `daemon/prep.go`.
- OpenAPI response mocking is implemented in `mock/mock_engine.go`.
- Static file mocks are implemented in `static-mock/`.
- Multi-spec discovery and conflict analysis live in `specs/`.
- HAR replay and validation live in `har/`.

Before updating CLI docs, verify the current binary behavior with:

```shell
./wiretap --help
```

If the local binary may be stale, use:

```shell
go run . --help
```

## Default ports

`wiretap` starts three local endpoints unless configured otherwise:

| Purpose | Default | Flag | Config key |
| --- | --- | --- | --- |
| API gateway | `http://localhost:9090` | `--port`, `-p` | `port` |
| Monitor UI | `http://localhost:9091` | `--monitor-port`, `-m` | `monitorPort` |
| Monitor websocket | `localhost:9092` | `--ws-port`, `-w` | `webSocketPort` |

When `--cert` and `--key` are provided, the gateway and monitor URLs use `https`.

## Common operating modes

### Proxy and validate live traffic

Use this when an upstream API exists and traffic should be forwarded while requests and responses are checked against OpenAPI.

```shell
wiretap -u https://api.example.com -s ./openapi.yaml
```

Incoming requests go to `localhost:9090`, are validated, proxied to the upstream URL, and displayed in the monitor UI.

### Full OpenAPI mock mode

Use this when the backend does not exist, should not be called, or a client needs a deterministic contract-driven simulation.

```shell
wiretap -s ./openapi.yaml --mock-mode
```

Mock mode requires at least one OpenAPI spec. It does not require `--url`, because no upstream API is called. Every API response is generated from the loaded OpenAPI contract.

Useful controls:

| Control | Effect |
| --- | --- |
| `Preferred: ExampleName` request header | Selects a named OpenAPI example when one exists for the matched operation response. |
| `wiretap-status-code: 404` request header | Overrides the returned status code after a successful mock body is generated. |
| `--hard-validation` | Rejects invalid mock-mode requests instead of serving a mock. |
| `wiretap-bypass-validation: true` request header | Lets a malformed request still receive the preferred/status-code mock when hard validation is enabled. |
| `--mock-bypass-validation` | Applies that bypass to all mock-mode requests. |
| `--strict-mode` | Reports undeclared properties, parameters, headers, and cookies as validation errors. |
| `--enable-all-mock-response-fields=false` | Limits generated mock response fields to required properties only. The flag defaults to `true`. |

By default, the mock engine chooses the lowest declared 2xx response for the matched operation, prefers the requested media type, and falls back to `application/json` when possible.

### Path-scoped OpenAPI mocking

Use this when some routes should be mocked and others should still proxy to an upstream API.

```yaml
redirectURL: https://api.example.com
contract: ./openapi.yaml
mockModeList:
  - /checkout/**
  - /beta/*
```

Start with:

```shell
wiretap -c ./wiretap.yaml
```

Requests matching `mockModeList` return generated mock responses. Other requests are proxied to `redirectURL`.

### Static mock definitions

Use this when exact request/response fixtures should override normal proxy behavior.

```shell
wiretap -u https://api.example.com --static-mock-dir ./mocks
```

Expected directory layout:

```text
./mocks/
  mock-definitions/
    get-product.json
  body-jsons/
    product.json
```

Each file in `mock-definitions/` can contain one definition or an array of definitions:

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

Static mock matching supports host, method, URL path, headers, query parameters, and JSON or form bodies. String fields can be exact values or regex-like patterns. Response bodies can use request-derived template values such as `${queryParams.view}`.

If no static mock matches, Wiretap falls back to the normal request handler, so static mocks can be combined with live proxying, OpenAPI validation, scoped mock mode, static hosting, and the monitor UI. Definition files are watched and reloaded when JSON files are created, updated, removed, or renamed.

### Multi-spec validation

Use this when a gateway or system is described by multiple OpenAPI contracts.

```shell
wiretap -u https://api.example.com --specs ./users.yaml,./orders.yaml
```

```shell
wiretap -u https://api.example.com --spec-dir ./contracts
```

Discovery accepts explicit files, globs, remote URLs, and directories. Directory discovery scans `.yaml`, `.yml`, and `.json` files containing OpenAPI or Swagger markers.

Run discovery and conflict analysis without starting servers:

```shell
wiretap --spec-dir ./contracts --dry-run
```

Use `--ignore` to exclude specs from discovery and `--ignore-clashing-operationid` when duplicate `operationId` values are intentional.

### HAR replay and validation

Load a HAR file into the monitor instead of sniffing live traffic:

```shell
wiretap --har ./traffic.har
```

Validate a HAR file against OpenAPI in headless mode:

```shell
wiretap --har ./traffic.har --har-validate -s ./openapi.yaml --har-allow /api
```

`--har-validate` requires `--har`, at least one spec, and at least one `--har-allow` path. It cannot be combined with mock mode.

Use `--har-replay-delay` to adjust replay pacing in milliseconds.

### Headless report streaming

Stream validation violations to a JSONL report while Wiretap runs:

```shell
wiretap -u https://api.example.com -s ./openapi.yaml --stream-report --report-filename ./wiretap-report.jsonl
```

### Static content hosting

Serve a local web app and proxy its API calls through the same Wiretap instance:

```shell
wiretap -u https://api.example.com --static ./dist --static-index index.html
```

For SPA catch-all paths, use `staticPaths` in `wiretap.yaml`.

### TLS, HTTPS, and HTTP/2

```shell
wiretap -u https://api.example.com --cert ./config/sample_cert.crt --key ./config/sample_key.pem
```

This switches Wiretap's local gateway and monitor to HTTPS.

## Complete flag checklist

| Flag | Short | Purpose |
| --- | --- | --- |
| `--url` | `-u` | Upstream redirect URL for proxy mode. |
| `--delay` | `-d` | Global API response delay in milliseconds. |
| `--port` | `-p` | API gateway port. |
| `--monitor-port` | `-m` | Monitor UI port. |
| `--ws-port` | `-w` | Monitor websocket port. |
| `--ws-host` | `-v` | Backend hostname used by the monitor. |
| `--spec` | `-s` | Primary OpenAPI spec path or URL. |
| `--specs` | `-S` | List of OpenAPI specs. |
| `--spec-dir` | | Directory roots to scan recursively for specs. |
| `--ignore` | | Glob patterns ignored during spec discovery. |
| `--dry-run` | | Discover, load, analyze, report, and exit. |
| `--ignore-clashing-operationid` | | Suppress duplicate `operationId` conflicts. |
| `--static` | `-t` | Static file directory to serve. |
| `--static-index` | `-i` | Index file name for static serving. |
| `--cert` | `-n` | TLS certificate path. |
| `--key` | `-k` | TLS private key path. |
| `--hard-validation` | `-e` | Return HTTP errors for non-compliant traffic. |
| `--hard-validation-code` | `-q` | Request validation error code. |
| `--hard-validation-return-code` | `-y` | Response validation error code. |
| `--hard-error-return-problem` | | Return RFC 9457 problem details for hard-validation errors. |
| `--static-mock-dir` | | Directory containing static mock definitions. |
| `--mock-mode` | `-x` | Generate all API responses from OpenAPI. |
| `--mock-bypass-validation` | | Serve preferred/status-code mocks even for malformed requests in hard-validation mock mode. |
| `--enable-all-mock-response-fields` | `-o` | Include examples for all properties in generated mocks; defaults to `true`. |
| `--config` | `-c` | Configuration file path. Defaults to `wiretap.yaml` discovery. |
| `--base` | `-b` | Base path or base URL for resolving references. |
| `--debug` | `-l` | Enable debug logging. |
| `--har` | `-z` | Load a HAR file instead of sniffing live traffic. |
| `--har-validate` | `-g` | Validate a HAR file against OpenAPI and exit. |
| `--har-allow` | `-j` | Add a HAR validation allow-list path. Repeatable. |
| `--har-replay-delay` | | Delay between replayed HAR events. |
| `--report-filename` | `-f` | Report output filename. |
| `--stream-report` | `-a` | Stream violations to the report file while running. |
| `--strict-redirect-location` | `-r` | Rewrite redirect `Location` headers to the Wiretap gateway host. |
| `--strict-mode` | | Detect undeclared request and response elements. |

## Config-only features

Some features are only available through `wiretap.yaml`:

- `variables` for reusable values in path, header, and auth configuration.
- `paths` for per-path targets, path rewrites, basic auth, rewrite identifiers, and header injection or dropping.
- `pathDelays` for per-path response delays.
- `headers` for global header injection or dropping.
- `staticPaths` for SPA catch-all static serving.
- `ignoreRedirects` and `redirectAllowList` for redirect control.
- `ignoreValidation` and `validationAllowList` for validation scoping.
- `hardValidationList` for path-scoped hard validation.
- `mockModeList` for path-scoped OpenAPI mocking.
- `websockets` for websocket proxy configuration.

## Verification before changing docs

Run focused checks that match the change:

```shell
go test ./cmd ./config ./mock ./static-mock ./specs ./har ./daemon/...
```

If tests fail because the sandbox cannot bind local ports, report that explicitly and run narrower package tests that do not require listeners.

Always preserve unrelated worktree changes. At the time this guide was added, generated or local files can exist in the repository root; do not remove them unless the user asks.
