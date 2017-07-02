package gozilla

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type reply struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type codecRequest struct {
	*http.Request
}

func (r codecRequest) Method() (string, error) {
	fs := strings.Split(r.URL.Path, "/")
	if len(fs) < 3 {
		return "", errors.New("bad url")
	}
	return fs[1] + "." + fs[2], nil
}

func (r codecRequest) ReadRequest(x interface{}) error {
	return errors.New("not implemention")
}

// Writes the response using the RPC method reply.
func (r codecRequest) WriteResponse(w http.ResponseWriter, x interface{}) {
	w.Header().Set("Content-Type", "application/json")
	rep := reply{
		Code: 200,
		Data: x,
	}
	enc := json.NewEncoder(w)
	enc.Encode(&rep)
}

// Writes an error produced by the server.
func (r codecRequest) WriteError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	rep := reply{
		Code: status,
		Msg:  err.Error(),
	}

	switch e := err.(type) {
	case *Error:
		rep.Code = e.Code
	case *ErrorWithData:
		rep.Code = e.Code
		rep.Data = e.Data
	}

	w.WriteHeader(rep.Code)
	enc := json.NewEncoder(w)
	enc.Encode(&rep)
}
