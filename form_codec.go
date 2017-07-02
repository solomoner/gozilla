package gozilla

import (
	"net/http"

	"github.com/gorilla/schema"
)

var formDecoder = schema.NewDecoder()

type FormCodec struct {
}

func (c FormCodec) NewRequest(r *http.Request) CodecRequest {
	return formCodecRequest{codecRequest{r}}
}

type formCodecRequest struct {
	codecRequest
}

func (r formCodecRequest) ReadRequest(x interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	return formDecoder.Decode(x, r.Form)
}
