package staticMock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pb33f/wiretap/shared"
)

func (sms *StaticMockService) getBodyFromMockDefinition(matchedMockDefinition StaticMockDefinition, request *http.Request) string {
	bodyStr := matchedMockDefinition.Response.Body

	// If the BodyJsonPath is defined then set the body to contents of the file
	if matchedMockDefinition.Response.BodyJsonFilename != "" {
		bodyJsonFilePath := sms.wiretapService.StaticMockDir + "/body-jsons/" + matchedMockDefinition.Response.BodyJsonFilename

		file, err := os.ReadFile(bodyJsonFilePath)
		if err != nil {
			panic(err)
		}

		bodyStr = string(file)
	}

	requestObjectWithIncomingRequestValues := StaticMockDefinitionRequest{
		Method:  request.Method,
		UrlPath: request.URL.Path,
		Host:    request.Host,
	}
	if (request.Body != nil) && (request.Body != http.NoBody) {
		requestObjectWithIncomingRequestValues.Body = sms.getBodyFromHttpRequest(request)
	}
	queryParams := make(map[string]any)
	if request.URL.Query() != nil {
		for k, v := range request.URL.Query() {
			// If there is only one value in the slice, store it as a string
			if len(v) == 1 {
				queryParams[k] = v[0]
			} else {
				// Otherwise, store the slice as is
				queryParams[k] = v
			}
		}
	}
	requestObjectWithIncomingRequestValues.QueryParams = &queryParams

	typeStrippedRequestJson, err := json.Marshal(requestObjectWithIncomingRequestValues)

	if err != nil {
		panic(err)
	}

	var typeStrippedRequest interface{}
	err = json.Unmarshal(typeStrippedRequestJson, &typeStrippedRequest)
	if err != nil {
		panic(err)
	}

	templateReplacedBodyStr, err := shared.ReplaceTemplateVars(bodyStr, typeStrippedRequest)
	if err != nil {
		panic(err)
	}

	return templateReplacedBodyStr
}

func (sms *StaticMockService) getHeadersFromMockDefinition(matchedMockDefinition StaticMockDefinition) http.Header {
	header := http.Header{}
	// wiretap needs to work from anywhere, so allow everything.
	headers := make(map[string][]string)
	shared.SetCORSHeaders(headers)
	headers["Content-Type"] = []string{"application/json"}

	// Add cors and content-type headers
	for k, v := range headers {
		header.Add(k, fmt.Sprint(v))
	}

	// Add headers from mock definition JSON
	for k, v := range matchedMockDefinition.Response.Header {
		header.Add(k, fmt.Sprint(v))
	}

	return header
}

func (sms *StaticMockService) getStaticMockResponse(matchedMockDefinition StaticMockDefinition, request *http.Request) *http.Response {
	body := sms.getBodyFromMockDefinition(matchedMockDefinition, request)

	buff := bytes.NewBuffer([]byte(body))

	response := &http.Response{
		StatusCode: matchedMockDefinition.Response.StatusCode,
		Body:       io.NopCloser(buff),
	}
	response.Header = sms.getHeadersFromMockDefinition(matchedMockDefinition)

	return response
}
