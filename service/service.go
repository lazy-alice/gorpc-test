package service

import (
	"alice_gorpc/interceptor"
	"alice_gorpc/metadata"
	"alice_gorpc/msg"
	"alice_gorpc/utils"
	"context"
	"errors"
	"time"
)

type Service struct {
	Svr      interface{}
	Ctx      context.Context
	Cancel   context.CancelFunc
	Timeout  time.Duration
	Name     string
	Handlers map[string]Handler
	Opts     *Options
}

// SvrDescription is a detailed description of a service
type SvrDescription struct {
	Svr         interface{}
	ServiceName string
	Methods     []*MethodDescription
	HandlerType interface{}
}

// MethodDescription is a detailed description of a method
type MethodDescription struct {
	MethodName string
	Handler    Handler
}

type Handler func(context.Context, interface{}, []interceptor.ServerInterceptor) (interface{}, error)

//func (s *Service) Serve() {
//	serverTransport := transport.GetServerTransport("tcp")
//	s.Ctx, s.Cancel = context.WithCancel(context.Background())
//
//	if err := serverTransport.ListenAndServe(s.Ctx); err != nil {
//		log.Default().Printf("tcp listen and serve error %v", err)
//		return
//	}
//
//	fmt.Printf("%s service listen at %s:%d", s.Name, s.)
//
//	<-s.Ctx.Done()
//}

func (s *Service) Handle(ctx context.Context, request *msg.Request) (interface{}, error) {
	ctx = metadata.WithServerMetadata(ctx, request.Metadata)
	if s.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		s.Cancel = cancel
		defer cancel()
	}

	_, method, err := utils.ParseServicePath(request.ServicePath)
	if err != nil {
		return nil, err
	}

	handle := s.Handlers[method]
	if handle == nil {
		return nil, errors.New("handle is nil")
	}

	rsp, err := handle(ctx, request, s.Opts.interceptors)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (s *Service) ServiceName() string {
	return s.Name
}
