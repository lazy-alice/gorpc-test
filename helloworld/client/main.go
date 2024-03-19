package main

import (
	"alice_gorpc/client"
	"alice_gorpc/plugin/auth"
	"alice_gorpc/plugin/circuit"
	"alice_gorpc/plugin/consul"
	"alice_gorpc/plugin/jaeger"
	"alice_gorpc/plugin/retry"
	"alice_gorpc/test_data"
	"context"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"time"
)

func main() {
	consul.Init("127.0.0.1:8500")

	tracer, err := jaeger.Init("localhost:6831")
	if err != nil {
		panic(err)
	}
	servicePath := "/greeter/SayHello"
	opts := []client.Option{
		client.WithServiceName("greeter"),
		client.WithTimeout(2000 * time.Millisecond),
		client.WithPerRPCAuth(auth.NewOAuth2ByToken("testToken")),
		client.WithInterceptor(jaeger.OpenTracingClientInterceptor(tracer, "/greeter/SayHello")),
		client.WithInterceptor(circuit.CircuitClientInterceptor("greeter", hystrix.CommandConfig{
			Timeout:                5000,
			MaxConcurrentRequests:  5,
			RequestVolumeThreshold: 1,
			SleepWindow:            5000,
			ErrorPercentThreshold:  1,
		}, "/greeter/SayHello")),
		client.WithInterceptor(retry.RetryClientInterceptor(servicePath, 1, 1*time.Millisecond)),
	}
	c := client.NewClient()
	req1 := test_data.HelloRequest{Msg: "hello"}
	req2 := test_data.HelloRequest{Msg: "zzzzz"}

	rsp1 := &test_data.HelloReply{}
	rsp2 := &test_data.HelloReply{}
	err = c.Call(context.Background(), "/greeter/SayHello", req1, rsp1, opts...)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("rsp1:", rsp1)
	//time.Sleep(1 * time.Second)
	err = c.Call(context.Background(), "/greeter/SayHello", req2, rsp2, opts...)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("rsp2:", rsp2)
	//
	req3 := test_data.HelloRequest{Msg: "hello"}
	rsp3 := &test_data.HelloReply{}
	time.Sleep(1 * time.Second)
	err = c.Call(context.Background(), "/greeter/SayHello", req3, rsp3, opts...)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("rsp3:", rsp3)
}
