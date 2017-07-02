package gozilla

import (
	"encoding/json"
	"net/http"
)

type JSONCodec struct {
}

func (c JSONCodec) NewRequest(r *http.Request) CodecRequest {
	return _JSONCodecRequest{codecRequest{r}}
}

type _JSONCodecRequest struct {
	codecRequest
}

func (r _JSONCodecRequest) ReadRequest(x interface{}) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(x)
}
