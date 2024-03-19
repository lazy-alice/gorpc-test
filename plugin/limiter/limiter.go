package limiter

import (
	"alice_gorpc/interceptor"
	"context"
	"errors"
	"golang.org/x/time/rate"
	"log"
)

// LimiterServerInterceptor limit-令牌的填充速度 burst-令牌桶的容量
func LimiterServerInterceptor(serviceName string, limit rate.Limit, burst int) interceptor.ServerInterceptor {
	limiters := make(map[string]map[string]*rate.Limiter)
	return func(ctx context.Context, req interface{}, handler interceptor.Handler, methodName string) (interface{}, error) {
		serviceLimiters, ok := limiters[serviceName]
		if !ok {
			serviceLimiters = make(map[string]*rate.Limiter)
			limiters[serviceName] = serviceLimiters
		}
		limiter, ok := serviceLimiters[methodName]
		if !ok {
			limiter = rate.NewLimiter(limit, burst)
			limiters[serviceName][methodName] = limiter
		}
		if !limiter.Allow() {
			log.Printf("limiting...\n")
			return nil, errors.New("limiting....")
		}
		return handler(ctx, req)
	}
}
