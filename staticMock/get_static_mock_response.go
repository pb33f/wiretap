package staticMock

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/pb33f/wiretap/shared"
)

func (sms *StaticMockService) getStaticMockResponse(matchedMockDefinition StaticMockDefinition) *http.Response {
	bodyStr := matchedMockDefinition.Response.Body

	// If the BodyJsonPath is defined then set the body to contents of the file
	if matchedMockDefinition.Response.BodyJsonPath != "" {
		bodyJsonFilePath := sms.wiretapService.StaticMockDir + "/body-jsons" + matchedMockDefinition.Response.BodyJsonPath

		file, err := os.ReadFile(bodyJsonFilePath)
		if err != nil {
			panic(err)
		}

		bodyStr = string(file)
	}

	typeStrippedMockRequestJson, err := json.Marshal(matchedMockDefinition.Request)
	if err != nil {
		panic(err)
	}

	var typeStrippedMockRequest interface{}
	err = json.Unmarshal(typeStrippedMockRequestJson, &typeStrippedMockRequest)
	if err != nil {
		panic(err)
	}

	templateReplacedBodyStr, err := shared.ReplaceTemplateVars(bodyStr, typeStrippedMockRequest)
	if err != nil {
		panic(err)
	}

	buff := bytes.NewBuffer([]byte(templateReplacedBodyStr))

	response := &http.Response{
		StatusCode: matchedMockDefinition.Response.StatusCode,
		Body:       io.NopCloser(buff),
	}
	header := http.Header{}
	response.Header = header

	return response
}
