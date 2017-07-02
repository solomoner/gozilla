package gozilla

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

type echoService struct {
}

type EchoRequest struct {
	Body string `validate:"required"`
}

func (s echoService) Echo(ctx *Context, req *EchoRequest) (*EchoRequest, error) {
	return req, nil
}

func TestServerValidate(t *testing.T) {
	rpcServer := NewServer(DefaultOptions())
	rpcServer.RegisterService(echoService{}, "echo")
	rpcServer.RegisterCodec(JSONCodec{}, "application/json")

	body, _ := json.Marshal(&EchoRequest{""})
	r := httptest.NewRequest("POST", "/echo/Echo", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	rpcServer.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("bad code:%v", w.Code)
	}
}

func BenchmarkServerPOSTJSON(b *testing.B) {
	rpcServer := NewServer(new(Options))
	rpcServer.RegisterService(echoService{}, "echo")
	rpcServer.RegisterCodec(JSONCodec{}, "application/json")

	body, _ := json.Marshal(&EchoRequest{"hello"})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := httptest.NewRequest("POST", "/echo/Echo", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			rpcServer.ServeHTTP(w, r)
			if w.Code > 200 {
				b.Fatalf("bad code:%v", w.Code)
			}
		}
	})
}

func BenchmarkServerPOSTForm(b *testing.B) {
	rpcServer := NewServer(new(Options))
	rpcServer.RegisterService(echoService{}, "echo")
	rpcServer.RegisterCodec(FormCodec{}, "application/x-www-form-urlencoded")

	value := url.Values{}
	value.Add("echo", "hello")
	body := value.Encode()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := httptest.NewRequest("POST", "/echo/Echo", strings.NewReader(body))
			w := httptest.NewRecorder()
			rpcServer.ServeHTTP(w, r)
			if w.Code > 200 {
				b.Fatalf("bad code:%v", w.Code)
			}
		}
	})
}

func BenchmarkServerPostFormLogging(b *testing.B) {
	rpcServer := NewServer(new(Options))
	rpcServer.RegisterService(echoService{}, "echo")
	rpcServer.RegisterCodec(FormCodec{}, "application/x-www-form-urlencoded")
	dir, err := ioutil.TempDir("", "gozilla")
	if err != nil {
		b.Fatal(err)
	}
	DefaultLogOpt.BaseDir = dir
	defer os.RemoveAll(dir)

	logger := NewLoggerHandler(DefaultLogOpt, rpcServer)

	value := url.Values{}
	value.Add("echo", "hello")
	body := value.Encode()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := httptest.NewRequest("POST", "/echo/Echo", strings.NewReader(body))
			w := httptest.NewRecorder()
			logger.ServeHTTP(w, r)
			if w.Code > 200 {
				b.Fatalf("bad code:%v", w.Code)
			}
		}
	})
}
