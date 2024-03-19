package client

import (
	"alice_gorpc/interceptor"
	"alice_gorpc/plugin/auth"
	"time"
)

type Options struct {
	timeout     time.Duration
	serviceName string
	//method      string
	target       string
	selectorName string
	interceptors []interceptor.ClientInterceptor
	perRPCAuth   []*auth.OAuth2
}

type Option func(options *Options)

func WithTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.timeout = t
	}
}

func WithServiceName(name string) Option {
	return func(o *Options) {
		o.serviceName = name
	}
}

func WithTarget(address string) Option {
	return func(o *Options) {
		o.target = address
	}
}

func WithSelectorName(name string) Option {
	return func(o *Options) {
		o.selectorName = name
	}
}

func WithInterceptor(interceptors ...interceptor.ClientInterceptor) Option {
	return func(o *Options) {
		o.interceptors = append(o.interceptors, interceptors...)
	}
}

func WithPerRPCAuth(rpcAuth *auth.OAuth2) Option {
	return func(o *Options) {
		o.perRPCAuth = append(o.perRPCAuth, rpcAuth)
	}
}
