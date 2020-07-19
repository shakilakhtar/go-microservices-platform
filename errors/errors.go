package errors

import (
	logger "github.com/sirupsen/logrus"
	"encoding/json"
	"errors"
	"net/http"
)

type Errors struct {
	// We use an array of *Error to set common errors in variables. More on that later.
	Errors []*Error `json:"errors"`
}

type Error struct {
	Type        string `json:"error"`
	Id          string `json:"id"`
	Status      int    `json:"status"`
	Description string `json:"error_description"`
}

const (
	BAD_REQUEST              = "bad_request"
	UNAUTHORIZED             = "unauthorized"
	MSG_UNAUTHORIZED         = "The authorization token does not seem to get you access at the moment. Please contact admin"
	NO_ACCESS_TOKEN_PROVIDED = "no_authorization_token_provided"
	INTERNAL_SERVER_ERROR    = "internal_server_error"
)

var (
	// ErrUnknown is used when a requested resource could not be found.
	ErrUnknown = NewError("unknown resource")

	// ErrInvalidArgument is returned when one or more arguments are invalid.
	ErrInvalidArgument = NewError("invalid argument")
	ErrBadRequest      = &Error{Id: BAD_REQUEST, Status: http.StatusBadRequest, Description: NO_ACCESS_TOKEN_PROVIDED}
	ErrInternalServer  = &Error{Id: INTERNAL_SERVER_ERROR, Status: 500, Description: "Internal Server Error.Something went wrong."}
	ErrUnauthorized    = &Error{Id: UNAUTHORIZED, Status: http.StatusUnauthorized, Description: MSG_UNAUTHORIZED}
)

// HandleError creates an errors.error type with a given string, logs the error and returns it
func HandleError(errMessage string) error {
	err := errors.New(errMessage)
	logger.Error(err.Error())
	return err
}

//Creates an standard error for a given error message
func NewError(msg string) error {
	return errors.New(msg)
}

//Write errors to HTTP response
func WriteError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	json.NewEncoder(w).Encode(Errors{[]*Error{err}})
}
