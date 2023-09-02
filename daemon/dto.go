// Copyright 2023 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: AGPL

package daemon

import (
	"github.com/pb33f/libopenapi-validator/errors"
	"net/textproto"
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
	Timestamp       int64                  `json:"timestamp,omitempty"`
	URL             string                 `json:"url,omitempty"`
	Method          string                 `json:"method,omitempty"`
	Host            string                 `json:"host,omitempty"`
	Path            string                 `json:"path,omitempty"`
	OriginalPath    string                 `json:"originalPath,omitempty"`
	DroppedHeaders  []string               `json:"droppedHeaders,omitempty"`
	InjectedHeaders map[string]string      `json:"injectedHeaders,omitempty"`
	Query           string                 `json:"query,omitempty"`
	Headers         map[string]any         `json:"headers,omitempty"`
	Body            string                 `json:"requestBody,omitempty"`
	Cookies         map[string]*HttpCookie `json:"cookies,omitempty"`
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
