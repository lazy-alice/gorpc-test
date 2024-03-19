package transport

import (
	"alice_gorpc/service"
	"time"
)

type ServerTransportOptions struct {
	Address         string
	KeepAlivePeriod time.Duration
	Handlers        map[string]*service.Service
}

type ServerOption func(*ServerTransportOptions)

func WithAddress(address string) ServerOption {
	return func(o *ServerTransportOptions) {
		o.Address = address
	}
}

func WithKeepAlivePeriod(period time.Duration) ServerOption {
	return func(o *ServerTransportOptions) {
		o.KeepAlivePeriod = period
	}
}

func WithHandler(handler map[string]*service.Service) ServerOption {
	return func(o *ServerTransportOptions) {
		o.Handlers = handler
	}
}
