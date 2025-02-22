# Static Mocking in Wiretap

## Table of Contents
- [Overview](#overview)
- [How to Enable Static Mocking](#how-to-enable-static-mocking)
- [Mock Definitions](#mock-definitions)
  - [Request Definition](#request-definition)
  - [Response Definition](#response-definition)
- [Response Generation Using Request Data](#response-generation-using-request-data)
- [Directory Structure](#directory-structure)
- [Example](#example)
- [Notes](#notes)

## Overview

This feature allows static mocking of APIs in the Wiretap service by defining mock definitions in JSON files. It enables the server to match incoming requests against predefined mock definitions and return corresponding mock responses. If no match is found, the request is forwarded to the Wiretap's httpRequestHandler for further processing.

## How to Enable Static Mocking

To enable static mocking, you need to set the `--static-mock-dir` argument to a directory path when starting Wiretap, or configure it in the Wiretap configuration file.

Example:
```bash
wiretap --static-mock-dir /path/to/mocks -u http://some.api-host-outhere.some-domain.com
```

When this path is set, Wiretap will expect mock definitions and response body JSON files in the following structure:

- `/path/to/mocks/mock-definitions/` — Contains the mock definition JSON files.
- `/path/to/mocks/body-jsons/` — Contains the response body JSON files.

The static mock service will start and load all the mock definitions found in `/path/to/mocks/mock-definitions`.

## Mock Definitions

Mock definitions are JSON objects or arrays of objects that define the request and response structure. Each JSON object should contain the following keys:

- **request** — Specifies the conditions for the request.
- **respose** — Specifies the response that should be returned when the request matches the conditions.

### Request Definition

The request definition is parsed into the following Go type:

```go
type StaticMockDefinitionRequest struct {
	Method      string          `json:"method,omitempty"`
	UrlPath     string          `json:"urlPath,omitempty"`
	Host        string          `json:"host,omitempty"`
	Header      *map[string]any `json:"header,omitempty"`
	Body        interface{}     `json:"body,omitempty"`
	QueryParams *map[string]any `json:"queryParams,omitempty"`
}
```

Each field can use either a string or a regex string to match the actual request. For example, the `header`, `body`, and `queryParams` fields can contain regex patterns to match the incoming request.

#### Example Request Definition:

```json
{
	"method": "GET",
	"urlPath": "/test",
	"header": {
		"Content-Type": "application.*"
	},
	"queryParams": {
		"test": "ok",
		"arr": ["1", "2"]
	},
	"body": {
		"test": "o.*"
	}
}
```

### Response Definition

The response definition is parsed into the following Go type:

```go
type StaticMockDefinitionResponse struct {
	Header           map[string]any `json:"header,omitempty"`
	StatusCode       int            `json:"statusCode,omitempty"`
	Body             string         `json:"body,omitempty"`
	BodyJsonFilename string         `json:"bodyJsonFilename,omitempty"`
}
```

- `BodyJsonFilename`: The name of a file in the `body-jsons` folder, which contains the response body JSON. If this is specified, Wiretap will return the content of that file instead of using the `body` field.

#### Example Response Definition with Inline Body:

```json
{
	"statusCode": 200,
	"header": {
		"something-header": "test-ok"
	},
	"body": "{\"test\": \"${queryParams.arr.[1]}\"}"
}
```

In this example, the response body uses a reference to the request's query parameters.

#### Example Response Definition with Body from File:

```json
{
	"statusCode": 200,
	"header": {
		"something-header": "test-ok"
	},
	"bodyJsonFilename": "test.json"
}
```

In this example, Wiretap will look for a file named `test.json` in the `body-jsons` folder and return its content as the response body.

## Response Generation Using Request Data

The response body can dynamically generate values based on the request. This is done by using the request's fields (such as `queryParams`, `body`, etc.) in the response body.

For example:
```json
{
	"statusCode": 200,
	"header": {
		"something-header": "test-ok"
	},
	"body": "{\"test\": \"${queryParams.arr.[1]}\"}"
}
```

In this case, the response body will include the second element from the `arr` query parameter in the incoming request. The `${}` syntax is used to refer to the request's fields.

## Directory Structure

The `--static-mock-dir` should point to a directory that contains the following subdirectories and files:

```
/path/to/mocks/
  ├── mock-definitions/
  │     ├── mock1.json
  │     ├── mock2.json
  │     └── ...
  └── body-jsons/
        ├── test.json
        └── ...
```

- **mock-definitions/**: Contains the mock request and response definitions.
- **body-jsons/**: Contains the actual response body JSON files referenced by the mock definitions.

## Example

1. **Directory Structure:**

```
/mocks/
  ├── mock-definitions/
  │     ├── get-test-mock.json
  └── body-jsons/
        ├── test.json
```

2. **get-test-mock.json**:

```json
{
  "request": {
    "method": "GET",
    "urlPath": "/test",
    "header": {
      "Content-Type": "application.*"
    },
    "queryParams": {
      "test": "ok",
      "arr": ["1", "2"]
    },
    "body": {
      "test": "o.*"
    }
  },
  "response": {
    "statusCode": 200,
    "header": {
      "something-header": "test-ok"
    },
    "bodyJsonFilename": "test.json"
  }
}
```

3. **test.json**:

```json
{
  "test": "${queryParams.arr.[1]}"
}
```

4. **Run wiretap**:

```bash
wiretap -u "http://localhost:8089" --static-mock-dir "/path/to/mocks" --port 80
```

Note: Any requests not matching a static mock will be proxied to the URL specified by the `-u` parameter.

5. **Make a request via curl**:

```bash
curl --location --request GET 'http://localhost/test?test=ok&arr=1&arr=2' \
--header 'Content-Type: application/json' \
--data '{
    "test": "ok"
}'
```

With this configuration, when Wiretap receives a `GET` request to `/test`, it will respond with the following JSON which represents the content of `test.json` after substituting the appropriate query parameter value:

```json
{
        "test": "2"
}
```

## Notes

- If no mock definition is found that matches an incoming request, Wiretap will forward the request to the wiretap's request handler and let it return a response.
- The mock definitions can contain either a single object or an array of objects. In the case of an array, each object represents a separate mock definition.
g