package myerrors

import (
	"encoding/json"
	"errors"
)

type ErrSupport struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value,omitempty"`
	Msg      string `json:"msg"`
}

type ErrComplex struct {
	Errors     []ErrSupport `json:"errors,omitempty"`
	Msg        string       `json:"message,omitempty"`
	StatusCode int          `json:"-"`
}

func (e ErrComplex) Error() string {
	str, err := json.Marshal(e)
	if err != nil {
		return ErrInternalError.Error()
	}
	return string(str)
}

func MyUnwrap(e error) error {
	for errors.Unwrap(e) != nil {
		e = errors.Unwrap(e)
	}
	return e
}

var (
	ErrInternalError = errors.New("internal error")
	ErrBadRequest    = errors.New("bad request")
)
