// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gozilla

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	validator "gopkg.in/go-playground/validator.v9"

	"golang.org/x/net/trace"
)

type Context struct {
	*http.Request
	http.ResponseWriter
}

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

// Codec creates a CodecRequest to process each request.
type Codec interface {
	NewRequest(*http.Request) CodecRequest
}

// CodecRequest decodes a request and encodes a response using a specific
// serialization scheme.
type CodecRequest interface {
	// Reads the request and returns the RPC method name.
	Method() (string, error)
	// Reads the request filling the RPC method args.
	ReadRequest(interface{}) error
	// Writes the response using the RPC method reply.
	WriteResponse(http.ResponseWriter, interface{})
	// Writes an error produced by the server.
	WriteError(w http.ResponseWriter, status int, err error)
}

// ----------------------------------------------------------------------------
// Server
// ----------------------------------------------------------------------------

// NewServer returns a new RPC server.
func NewServer(opt *Options) *Server {
	s := &Server{
		opt:      opt,
		codecs:   make(map[string]Codec),
		services: new(serviceMap),
	}
	if opt.EnableValidator {
		s.validator = validator.New()
	}
	return s
}

// Server serves registered RPC services using registered codecs.
type Server struct {
	opt        *Options
	loghandler http.Handler
	codecs     map[string]Codec
	services   *serviceMap
	validator  *validator.Validate
}

// RegisterCodec adds a new codec to the server.
//
// Codecs are defined to process a given serialization scheme, e.g., JSON or
// XML. A codec is chosen based on the "Content-Type" header from the request,
// excluding the charset definition.
func (s *Server) RegisterCodec(codec Codec, contentType string) {
	s.codecs[strings.ToLower(contentType)] = codec
}

// RegisterService adds a new service to the server.
//
// The name parameter is optional: if empty it will be inferred from
// the receiver type name.
//
// Methods from the receiver will be extracted if these rules are satisfied:
//
//    - The receiver is exported (begins with an upper case letter) or local
//      (defined in the package registering the service).
//    - The method name is exported.
//    - The method has two arguments: *Context, *args.
//    - All two arguments are pointers.
//    - The second argument are exported or local.
//    - The method has two returns: reply and error.
//
// All other methods are ignored.
func (s *Server) RegisterService(receiver interface{}, name string) error {
	return s.services.register(receiver, name)
}

// HasMethod returns true if the given method is registered.
//
// The method uses a dotted notation as in "Service.Method".
func (s *Server) HasMethod(method string) bool {
	if _, _, err := s.services.get(method); err == nil {
		return true
	}
	return false
}

func (s *Server) newTrace(method string) trace.Trace {
	ss := strings.Split(method, ".")
	if len(ss) != 2 {
		panic("bad method")
	}
	return trace.New(ss[0], ss[1])
}

// ServeHTTP
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	idx := strings.Index(contentType, ";")
	if idx != -1 {
		contentType = contentType[:idx]
	}
	var codec Codec
	if contentType == "" && len(s.codecs) == 1 {
		// If Content-Type is not set and only one codec has been registered,
		// then default to that codec.
		for _, c := range s.codecs {
			codec = c
		}
	} else if codec = s.codecs[strings.ToLower(contentType)]; codec == nil {
		WriteError(w, 415, "rpc: unrecognized Content-Type: "+contentType)
		return
	}
	// Create a new codec request.
	codecReq := codec.NewRequest(r)
	// Get service method to be called.
	method, errMethod := codecReq.Method()
	if errMethod != nil {
		codecReq.WriteError(w, 400, errMethod)
		return
	}
	tr := s.newTrace(method)
	defer tr.Finish()

	serviceSpec, methodSpec, errGet := s.services.get(method)
	if errGet != nil {
		tr.LazyPrintf("method not found:%s", method)
		tr.SetError()
		codecReq.WriteError(w, 404, errGet)
		return
	}
	// Decode the args.
	args := reflect.New(methodSpec.argsType)
	if errRead := codecReq.ReadRequest(args.Interface()); errRead != nil {
		tr.LazyPrintf("read request error:%s", errRead)
		tr.SetError()
		codecReq.WriteError(w, 400, errRead)
		return
	}

	// Validate the args
	if s.opt.EnableValidator && args.Elem().Kind() == reflect.Struct {
		err := s.validator.Struct(args.Interface())
		if err != nil {
			tr.LazyPrintf("validate request error:%s", err)
			tr.SetError()
			codecReq.WriteError(w, 400, err)
			return
		}
	}

	// Catch method panic and return 500 error
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			s.opt.ErrorLog.Printf("gozilla: panic serving %v: %v\n%s", r.RemoteAddr, err, buf)
			codecReq.WriteError(w, 500, fmt.Errorf("%v", err))
		}
	}()

	r = r.WithContext(trace.NewContext(r.Context(), tr))
	ctx := &Context{Request: r, ResponseWriter: w}
	retValue := methodSpec.method.Func.Call([]reflect.Value{
		serviceSpec.rcvr,
		reflect.ValueOf(ctx),
		args,
	})
	// Cast the result to error if needed.
	var errResult error
	errInter := retValue[1].Interface()
	if errInter != nil {
		errResult = errInter.(error)
	}
	// Prevents Internet Explorer from MIME-sniffing a response away
	// from the declared content-type
	w.Header().Set("x-content-type-options", "nosniff")
	// Encode the response.
	if errResult == nil {
		codecReq.WriteResponse(w, retValue[0].Interface())
	} else {
		tr.LazyPrintf("call error:%s", errResult)
		tr.SetError()
		codecReq.WriteError(w, 400, errResult)
	}
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, msg)
}

var (
	// DefaultLogOpt是默认的日志配置项，通过修改来自定义日志配置
	DefaultLogOpt = DefaultLogOptions()

	// DefaultOpt是默认的server配置，通过修改来自定义server行为
	DefaultOpt = DefaultOptions()

	DefaultServer = NewServer(DefaultOpt)
)

func RegisterService(srv interface{}, name string) error {
	return DefaultServer.RegisterService(srv, name)
}

func ListenAndServe(addr string) error {
	loghander := NewLoggerHandler(DefaultLogOpt, DefaultServer)
	http.Handle("/", loghander)
	return http.ListenAndServe(addr, nil)
}

func init() {
	// for POST JSON body
	DefaultServer.RegisterCodec(JSONCodec{}, "application/json")

	// for GET
	DefaultServer.RegisterCodec(FormCodec{}, "")

	// for POST form
	DefaultServer.RegisterCodec(FormCodec{}, "application/x-www-form-urlencoded")

	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}
}
