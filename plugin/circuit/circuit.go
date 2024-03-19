package circuit

import (
	"alice_gorpc/interceptor"
	"context"
	"github.com/afex/hystrix-go/hystrix"
	"log"
)

//func init() {
//	config := hystrix.CommandConfig{
//		Timeout:                5000,
//		MaxConcurrentRequests:  5,
//		RequestVolumeThreshold: 1,
//		SleepWindow:            5000,
//		ErrorPercentThreshold:  1,
//	}
//	hystrix.ConfigureCommand("greeter", config)
//}

func CircuitClientInterceptor(serviceName string, config hystrix.CommandConfig, servicePath string) interceptor.ClientInterceptor {
	hystrix.ConfigureCommand(serviceName, config)
	return func(ctx context.Context, req, rsp interface{}, ivk interceptor.Invoker) error {
		return hystrix.Do(serviceName, func() error {
			return ivk(ctx, req, rsp, servicePath)
		}, func(err error) error {

			log.Printf("circut error:%v\n", err)
			return err
		})
	}
}
