// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package daemon

import (
	"bytes"
	"encoding/json"
	"github.com/pb33f/libopenapi-validator/errors"
	"github.com/pb33f/ranch/bus"
	"github.com/pb33f/ranch/model"
	"github.com/pb33f/wiretap/controls"
	"github.com/pb33f/wiretap/shared"
	"github.com/pterm/pterm"
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

type HttpCookie struct {
	Value   string `json:"value,omitempty"`
	Path    string `json:"path,omitempty"`
	Domain  string `json:"domain,omitempty"`
	Expires string `json:"expires,omitempty"`
	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int  `json:"maxAge,omitempty"`
	Secure   bool `json:"secure,omitempty"`
	HttpOnly bool `json:"httpOnly,omitempty"`
}

type HttpRequest struct {
	Timestamp int64                  `json:"timestamp,omitempty"`
	URL       string                 `json:"url,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Query     string                 `json:"query,omitempty"`
	Headers   map[string]any         `json:"headers,omitempty"`
	Body      string                 `json:"requestBody,omitempty"`
	Cookies   map[string]*HttpCookie `json:"cookies,omitempty"`
}

type HttpResponse struct {
	Timestamp  int64                  `json:"timestamp,omitempty"`
	Headers    map[string]any         `json:"headers,omitempty"`
	StatusCode int                    `json:"statusCode,omitempty"`
	Body       string                 `json:"responseBody,omitempty"`
	Cookies    map[string]*HttpCookie `json:"cookies,omitempty"`
	Time       time.Time              `json:"-"`
}

type HttpTransaction struct {
	Request            *HttpRequest              `json:"httpRequest,omitempty"`
	RequestValidation  []*errors.ValidationError `json:"requestValidation,omitempty"`
	Response           *HttpResponse             `json:"httpResponse,omitempty"`
	ResponseValidation []*errors.ValidationError `json:"responseValidation,omitempty"`
	Id                 string                    `json:"id,omitempty"`
}

func buildResponse(r *model.Request, response *http.Response) *HttpTransaction {
	code := 500
	headers := make(map[string]any)
	cookies := make(map[string]*HttpCookie)
	var respBody []byte

	if response != nil {
		code = response.StatusCode
		for k, v := range response.Header {
			headers[k] = v[0]
		}

		for _, c := range response.Cookies() {
			cookies[c.Name] = &HttpCookie{
				Value:    c.Value,
				Path:     c.Path,
				Domain:   c.Domain,
				Expires:  c.RawExpires,
				MaxAge:   c.MaxAge,
				Secure:   c.Secure,
				HttpOnly: c.HttpOnly,
			}
		}

		// sniff and replace response body.
		respBody, _ = io.ReadAll(response.Body)
		_ = response.Body.Close()
		response.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}
	return &HttpTransaction{
		Id: r.Id.String(),
		Response: &HttpResponse{
			Timestamp:  time.Now().UnixMilli(),
			Headers:    headers,
			StatusCode: code,
			Body:       string(respBody),
			Cookies:    cookies,
		},
	}
}

func buildRequest(r *model.Request) *HttpTransaction {

	storeManager := bus.GetBus().GetStoreManager()
	controlsStore := storeManager.CreateStore(controls.ControlServiceChan)
	config, _ := controlsStore.Get(shared.ConfigKey)

	newReq := cloneRequest(r.HttpRequest,
		config.(*shared.WiretapConfiguration).RedirectProtocol,
		config.(*shared.WiretapConfiguration).RedirectHost,
		config.(*shared.WiretapConfiguration).RedirectPort)

	var requestBody []byte

	headers := make(map[string]any)
	for k, v := range newReq.Header {
		headers[k] = v[0]
	}

	cookies := make(map[string]*HttpCookie)
	for _, c := range newReq.Cookies() {
		cookies[c.Name] = &HttpCookie{
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			Expires:  c.RawExpires,
			MaxAge:   c.MaxAge,
			Secure:   c.Secure,
			HttpOnly: c.HttpOnly,
		}
	}

	// check if request is a multipart form
	if ct, ok := headers["Content-Type"].(string); ok {
		if strings.Contains(ct, "multipart/form-data") {
			err := newReq.ParseMultipartForm(32 << 2)
			if err != nil {
				pterm.Error.Println(err.Error())
			}
		}
	}

	if newReq.MultipartForm != nil {
		var parts []FormPart
		for i := range newReq.MultipartForm.Value {
			parts = append(parts, FormPart{
				Name:  i,
				Value: newReq.MultipartForm.Value[i],
			})
		}
		for k, fHeaders := range newReq.MultipartForm.File {

			var formFiles []*FormFile

			for z := range fHeaders {
				ff := &FormFile{
					Name:    fHeaders[z].Filename,
					Headers: fHeaders[z].Header,
				}
				formFiles = append(formFiles, ff)
			}

			parts = append(parts, FormPart{
				Name:  k,
				Files: formFiles,
			})
		}
		requestBody, _ = json.Marshal(parts)
	} else {

		requestBody, _ = io.ReadAll(newReq.Body)

	}

	return &HttpTransaction{
		Id: r.Id.String(),
		Request: &HttpRequest{
			URL:       r.HttpRequest.URL.String(),
			Method:    r.HttpRequest.Method,
			Path:      r.HttpRequest.URL.Path,
			Query:     r.HttpRequest.URL.RawQuery,
			Cookies:   cookies,
			Headers:   headers,
			Body:      string(requestBody),
			Timestamp: time.Now().UnixMilli(),
		},
	}
}

type FormPart struct {
	Name  string      `json:"name,omitempty"`
	Value []string    `json:"value,omitempty"`
	Files []*FormFile `json:"files,omitempty"`
}

type FormFile struct {
	Name    string               `json:"name,omitempty"`
	Headers textproto.MIMEHeader `json:"headers,omitempty"`
	Data    string               `json:"data,omitempty"`
}
