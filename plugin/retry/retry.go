package retry

import (
	"alice_gorpc/interceptor"
	"context"
	"log"
	"time"
)

func RetryClientInterceptor(servicePath string, maxRetries int, retryInterval time.Duration) interceptor.ClientInterceptor {

	return func(ctx context.Context, req, rsp interface{}, ivk interceptor.Invoker) error {
		var err error
		for i := 0; i < maxRetries; i++ {
			err = ivk(ctx, req, rsp, servicePath)
			if err == nil {
				return nil
			}
			log.Printf("rpc call failed:%s;retrying....\n", err)
			time.Sleep(retryInterval)
		}
		return err
	}
}
