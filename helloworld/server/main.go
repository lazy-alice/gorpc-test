package main

import (
	"alice_gorpc/plugin/auth"
	"alice_gorpc/plugin/jaeger"
	"alice_gorpc/plugin/limiter"
	"alice_gorpc/server"
	"alice_gorpc/service"
	"alice_gorpc/test_data"
)

func main() {
	opts := []server.Option{
		server.WithIp("127.0.0.1"),
		server.WithPort(49999),
		server.WithPlugin("consul"),
		server.WithSelectorSvrAddr("127.0.0.1:8500"),
		server.WithTracingSvrAddr("localhost:6831"),
		server.WithTracingSpanName("greeter"),
		server.WithPlugin(jaeger.Name),
	}
	greeterServiceName := "greeter"
	serviceOpts := []service.Option{
		service.WithInterceptors(auth.BuildAuthInterceptor()),
		service.WithInterceptors(limiter.LimiterServerInterceptor(greeterServiceName, 1, 1)),
	}
	s := server.NewServer(opts...)
	if err := s.RegisterService(greeterServiceName, new(test_data.Service), serviceOpts); err != nil {
		panic(err)
	}
	s.Start()
}
