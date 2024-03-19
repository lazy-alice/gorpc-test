package plugin

type Plugin interface {
}

type ResolverPlugin interface {
	Init(...Option) error
}

var PluginMap = make(map[string]Plugin)

func Register(name string, p Plugin) {
	if PluginMap == nil {
		PluginMap = make(map[string]Plugin)
	}
	PluginMap[name] = p
}

type Options struct {
	SelectorSvrAddr string
	Services        []string
	SvrAddr         string
	TracingSvrAddr  string
}

type Option func(o *Options)

func WithSelectorSvrAddr(addr string) Option {
	return func(o *Options) {
		o.SelectorSvrAddr = addr
	}
}

func WithServices(services []string) Option {
	return func(o *Options) {
		o.Services = services
	}
}

func WithSvrAddr(addr string) Option {
	return func(o *Options) {
		o.SvrAddr = addr
	}
}

func WithTracingSvrAddr(addr string) Option {
	return func(o *Options) {
		o.TracingSvrAddr = addr
	}
}
