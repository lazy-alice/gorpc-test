package test_data

import (
	"context"
)

type Service struct {
}

type HelloRequest struct {
	Msg string
}

type HelloReply struct {
	Msg string
}

func (s *Service) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	//fmt.Println("request:", req)
	rsp := &HelloReply{
		Msg: "world",
	}

	return rsp, nil
}
