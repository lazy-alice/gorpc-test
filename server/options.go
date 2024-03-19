package server

type Options struct {
	ip              string
	port            int
	selectorSvrAddr string //service discovery address
	pluginNames     []string
	tracingSvrAddr  string
	tracingSpanName string
}

type Option func(*Options)

func WithIp(ip string) Option {
	return func(o *Options) {
		o.ip = ip
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

func WithSelectorSvrAddr(address string) Option {
	return func(o *Options) {
		o.selectorSvrAddr = address
	}
}

func WithPlugin(pluginName ...string) Option {
	return func(o *Options) {
		o.pluginNames = append(o.pluginNames, pluginName...)
	}
}

func WithTracingSvrAddr(addr string) Option {
	return func(o *Options) {
		o.tracingSvrAddr = addr
	}
}

func WithTracingSpanName(name string) Option {
	return func(o *Options) {
		o.tracingSpanName = name
	}
}
