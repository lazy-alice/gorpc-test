package conn_pool

import "time"

type Options struct {
	initialCap  int
	maxCap      int
	idleTimeout time.Duration
	maxIdle     int
	dialTimeout time.Duration
}

type Option func(*Options)

func WithInitialCap(initialCap int) Option {
	return func(o *Options) {
		o.initialCap = initialCap
	}
}

func WithMaxCap(maxCap int) Option {
	return func(o *Options) {
		o.maxCap = maxCap
	}
}

func WithMaxIdle(maxIdle int) Option {
	return func(o *Options) {
		o.maxIdle = maxIdle
	}
}

func WithIdleTimeout(idleTimeout time.Duration) Option {
	return func(o *Options) {
		o.idleTimeout = idleTimeout
	}
}

func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(o *Options) {
		o.dialTimeout = dialTimeout
	}
}
