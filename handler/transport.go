package handler

import (
	"bytes"
	dterrors "shakilakhtar/go-microservices-platform/errors"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

type errorer interface {
	error() error
}

const (
	ContentTypeHeader = "Content-Type"
	JsonMediaType     = "application/json; charset=utf-8"
)

type ErrorResponse struct {
	Status  string
	Message string
}

//Encode response and send results back to client
func EncodeResponse(w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		EncodeError(e.error(), w)
		return nil
	}
	w.Header().Set(ContentTypeHeader, JsonMediaType)
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func EncodeError(err error, w http.ResponseWriter) {
	switch err {
	case dterrors.ErrUnknown:
		w.WriteHeader(http.StatusNotFound)
	case dterrors.ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set(ContentTypeHeader, JsonMediaType)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// EncodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// JSON-encodes any request to the request body. Primarily useful in a client.
func EncodeHTTPGenericRequest(r *http.Request, request interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	defer r.Body.Close()
	return nil
}

// EncodeHTTPGenericResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func EncodeHTTPGenericResponse(w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func EncodeJSONRequest(req *http.Request, request interface{}) error {
	// All resource requests are encoded in the same way:
	// simple JSON serialization to the request body.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	defer req.Body.Close()
	return nil
}

func EncodeJSONResponse(w http.ResponseWriter, response interface{}) error {
	w.Header().Set(ContentTypeHeader, JsonMediaType)
	return json.NewEncoder(w).Encode(response)
}

type errorWrapper struct {
	Error string `json:"error"`
}

//Encode error and add any error message to show
func ErrorEncoder(err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	msg := err.Error()

	switch err {
	case dterrors.ErrUnknown:
		code = http.StatusNotFound
	case dterrors.ErrInvalidArgument:
		code = http.StatusBadRequest
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorWrapper{Error: msg})
}

//Decodes error and make a  new error from wrapper
func errorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return dterrors.NewError(w.Error)
}

func IsHeaderPresent(r *http.Request, header string) bool {
	isPresent := r.Header.Get(header)
	if isPresent != "" {
		return true
	}
	return false
}

// Determine whether the request `content-type` includes a
// server-acceptable mime-type
// Failure should yield an HTTP 415 (`http.StatusUnsupportedMediaType`)
func HasContentType(r *http.Request, mimeType string) bool {
	contentType := r.Header.Get(ContentTypeHeader)
	if contentType == "" {
		return mimeType == JsonMediaType
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimeType {
			return true
		}
	}
	return false
}

//Decodes the request object
func Decode(r *http.Request, v interface{}) error {

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}
