package service

import "alice_gorpc/interceptor"

type Options struct {
	interceptors []interceptor.ServerInterceptor
	Name         string
}

type Option func(o *Options)

func WithInterceptors(interceptors ...interceptor.ServerInterceptor) Option {
	return func(o *Options) {
		o.interceptors = append(o.interceptors, interceptors...)
	}
}

func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}
