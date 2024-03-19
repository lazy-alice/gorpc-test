package transport

import (
	"alice_gorpc/pool/conn_pool"
	"alice_gorpc/selector"
)

type ClientTransportOptions struct {
	targetServiceName string
	selector          *selector.Selector
	pool              *conn_pool.Pool
}

type ClientOption func(options *ClientTransportOptions)

func WithTargetServiceName(name string) ClientOption {
	return func(c *ClientTransportOptions) {
		c.targetServiceName = name
	}
}

func WithSelector(selector *selector.Selector) ClientOption {
	return func(c *ClientTransportOptions) {
		c.selector = selector
	}
}

func WithPool(p *conn_pool.Pool) ClientOption {
	return func(c *ClientTransportOptions) {
		c.pool = p
	}
}
