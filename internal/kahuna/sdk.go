/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/kowabunga-cloud/common/klog"
	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

const (
	SdkBaseRoute = "/api/v1"
)

type RequestArg struct {
	Name  string
	Value interface{}
}

func (a *RequestArg) String() string {
	return fmt.Sprintf("%s:%#v", a.Name, a.Value)
}

func RA(name string, value interface{}) RequestArg {
	return RequestArg{
		Name:  name,
		Value: value,
	}
}

func isDebug() bool {
	cfg := GetCfg()
	if cfg == nil {
		return false
	}
	return strings.EqualFold(cfg.Global.LogLevel, "DEBUG")
}

func LogHttpRequest(args ...RequestArg) {
	if !isDebug() {
		return
	}

	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return
	}

	funcname := runtime.FuncForPC(pc).Name()
	fn := funcname[strings.LastIndex(funcname, ".")+1:]
	var params strings.Builder
	for _, a := range args {
		params.WriteString(a.String())
		params.WriteString(" ")
	}
	msg := fmt.Sprintf("%s() request params - %s", fn, params.String())
	klog.Debug(msg)
}

func LogHttpResponse(body interface{}) {
	if !isDebug() {
		return
	}

	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return
	}

	funcname := runtime.FuncForPC(pc).Name()
	fn := funcname[strings.LastIndex(funcname, ".")+1:]
	msg := fmt.Sprintf("%s() response body - %+v", fn, body)
	klog.Debug(msg)
}

// ApiError is the JSON body of any API error response, as defined by the
// OpenAPI specification ApiError* schemas.
type ApiError struct {
	Status int32  `json:"status"`
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

// NewApiError returns the API error body for an HTTP status code. Server-side
// failures details are logged, not disclosed.
func NewApiError(code int, err error) ApiError {
	e := ApiError{
		Status: int32(code), // #nosec G115 -- HTTP status codes fit
		Error:  http.StatusText(code),
	}
	if err != nil {
		if code >= http.StatusInternalServerError && code != http.StatusInsufficientStorage {
			klog.Errorf("%s: %v", e.Error, err)
		} else {
			e.Reason = err.Error()
		}
	}
	return e
}

// ApiErrorHandler encodes request errors as JSON ApiError bodies, used by all
// API controllers in place of the SDK default one, which returns plain strings.
func ApiErrorHandler(w http.ResponseWriter, r *http.Request, err error, result *sdk.ImplResponse) {
	code := http.StatusInternalServerError
	var parsingErr *sdk.ParsingError
	var requiredErr *sdk.RequiredError
	switch {
	case errors.As(err, &parsingErr), errors.As(err, &requiredErr):
		code = http.StatusBadRequest
	case result != nil && result.Code != 0:
		code = result.Code
	}
	body := NewApiError(code, err)
	klog.Debugf("%s %s: %d %s", r.Method, r.URL.Path, code, body.Reason)
	_ = sdk.EncodeJSONResponse(body, &code, w)
}

// WriteApiError writes a JSON ApiError response, for requests rejected before
// reaching API controllers (e.g. by authentication middlewares).
func WriteApiError(w http.ResponseWriter, code int, err error) {
	_ = sdk.EncodeJSONResponse(NewApiError(code, err), &code, w)
}

func httpError(code int, err error) (sdk.ImplResponse, error) {
	return sdk.Response(code, NewApiError(code, err)), err
}

func HttpOK(body interface{}) (sdk.ImplResponse, error) {
	return sdk.Response(http.StatusOK, body), nil
}

func HttpUnauthorized(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusUnauthorized, err)
}

func HttpForbidden(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusForbidden, err)
}

func HttpCreated(body interface{}) (sdk.ImplResponse, error) {
	return sdk.Response(http.StatusCreated, body), nil
}

func HttpNoContent() (sdk.ImplResponse, error) {
	return sdk.Response(http.StatusNoContent, nil), nil
}

func HttpBadParams(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusBadRequest, err)
}

func HttpNotFound(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusNotFound, err)
}

func HttpConflict(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusConflict, err)
}

// HttpServerError reports a request the server can't process (e.g. a
// resource operation that failed, with its reason).
func HttpServerError(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusUnprocessableEntity, err)
}

// HttpInternalError reports an unexpected server-side failure, whose details
// are logged but not disclosed.
func HttpInternalError(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusInternalServerError, err)
}

func HttpQuota(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusInsufficientStorage, err)
}

func HttpNotImplemented(err error) (sdk.ImplResponse, error) {
	return httpError(http.StatusNotImplemented, err)
}
