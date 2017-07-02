package main

import "github.com/solomoner/gozilla"

type HelloRequest struct {
	Name string
}

type HelloReply struct {
	Reply string
}

type HelloService struct {
}

func (s *HelloService) Hello(ctx *gozilla.Context, r *HelloRequest) (*HelloReply, error) {
	rep := &HelloReply{Reply: "hello " + r.Name}
	return rep, nil
}

func (s *HelloService) HelloError(ctx *gozilla.Context, r *HelloRequest) (*HelloReply, error) {
	return nil, gozilla.NewError(440, "bad return")
}

func (s *HelloService) HelloErrorWithData(ctx *gozilla.Context, r *HelloRequest) (*HelloReply, error) {
	return nil, gozilla.NewErrorWithData(440, "bad return", r)
}

func main() {
	gozilla.RegisterService(new(HelloService), "hello")
	gozilla.ListenAndServe(":8000")
}
