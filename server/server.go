package server

import (
	"alice_gorpc/interceptor"
	"alice_gorpc/plugin"
	"alice_gorpc/plugin/consul"
	"alice_gorpc/plugin/jaeger"
	"alice_gorpc/service"
	"alice_gorpc/transport"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

type Server struct {
	opts     *Options
	Ctx      context.Context
	Cancel   context.CancelFunc
	Services map[string]*service.Service
	Plugins  []plugin.Plugin
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		opts:     &Options{},
		Services: map[string]*service.Service{},
	}
	for _, opt := range opts {
		opt(s.opts)
	}
	for pluginName, plugin := range plugin.PluginMap {
		if !containPlugin(s.opts.pluginNames, pluginName) {
			continue
		}
		s.Plugins = append(s.Plugins, plugin)
	}
	return s
}

func (s *Server) InitPlugins() error {
	for _, p := range s.Plugins {
		switch val := p.(type) {
		case consul.Consul:
			serviceNames := make([]string, 0)
			for _, s := range s.Services {
				serviceNames = append(serviceNames, s.ServiceName())
			}
			opts := []plugin.Option{
				plugin.WithSelectorSvrAddr(s.opts.selectorSvrAddr),
				plugin.WithServices(serviceNames),
				plugin.WithSvrAddr(fmt.Sprintf("%s:%d", s.opts.ip, s.opts.port)),
			}
			if err := val.Init(opts...); err != nil {
				log.Printf("resolver init error, %v", err)
				return err
			}
		case jaeger.Jaeger:

			pluginOpts := []plugin.Option{
				plugin.WithTracingSvrAddr(s.opts.tracingSvrAddr),
			}

			tracer, err := val.Init(pluginOpts...)
			if err != nil {
				log.Printf("tracing init error, %v", err)
				return err
			}

			for _, ser := range s.Services {
				option := service.WithInterceptors(jaeger.OpenTracingServerInterceptor(tracer, s.opts.tracingSpanName))
				option(ser.Opts)
				//service.Opts.interceptors = append(s.opts.interceptors, jaeger.OpenTracingServerInterceptor(tracer, s.opts.tracingSpanName))
			}

		default:
		}
	}
	return nil
}

func (s *Server) Start() {
	//for _, s := range s.Services {
	//	s.Serve()
	//}
	if err := s.InitPlugins(); err != nil {
		panic(err)
	}
	s.Serve()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGSEGV)
	<-ch
}

func (s *Server) Serve() {
	serviceMp := make(map[string]*service.Service)
	for _, s := range s.Services {
		serviceMp[s.Name] = s
	}
	transportOption := []transport.ServerOption{
		transport.WithAddress(fmt.Sprintf("%s:%d", s.opts.ip, s.opts.port)),
		transport.WithHandler(serviceMp),
	}

	serverTransport := transport.GetServerTransport("tcp")
	s.Ctx, s.Cancel = context.WithCancel(context.Background())

	if err := serverTransport.ListenAndServe(s.Ctx, transportOption...); err != nil {
		log.Default().Printf("tcp listen and serve error %v", err)
		return
	}

	fmt.Printf("sever listen at %s:%d\n", s.opts.ip, s.opts.port)

	<-s.Ctx.Done()
}

func (s *Server) RegisterService(serviceName string, svr interface{}, opts []service.Option) error {
	if svr == nil {
		return nil
	}
	serviceType := reflect.TypeOf(svr)
	serviceValue := reflect.ValueOf(svr)
	//svrDescription := &service.SvrDescription{
	//	serviceName: serviceName,
	//	HandlerType: (*interface{})(nil),
	//	Svr:         svr,
	//}
	methods, err := getServiceMethods(serviceType, serviceValue)
	if err != nil {
		log.Printf("get method error:%s", err.Error())
		return err
	}
	//svrDescription.Methods = methods

	//handlerType := reflect.TypeOf(svrDescription.HandlerType).Elem()
	//if !serviceType.Implements(handlerType) {
	//	log.Fatalf("handlerType %v not match service : %v ", handlerType, serviceType)
	//}

	ser := &service.Service{
		Ctx:      context.Background(),
		Svr:      svr,
		Name:     serviceName,
		Handlers: map[string]service.Handler{},
		Opts:     &service.Options{},
	}
	for _, method := range methods {
		ser.Handlers[method.MethodName] = method.Handler
	}
	for _, o := range opts {
		o(ser.Opts)
	}
	s.Services[serviceName] = ser
	return nil
}

func getServiceMethods(serviceType reflect.Type, serviceValue reflect.Value) ([]*service.MethodDescription, error) {
	var methods []*service.MethodDescription
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		if err := checkMethod(method.Type); err != nil {
			return nil, err
		}

		methodHandler := func(ctx context.Context, svr interface{}, ceps []interceptor.ServerInterceptor) (interface{}, error) {
			reqType := method.Type.In(2)
			req := reflect.New(reqType.Elem()).Interface()

			handler := func(ctx context.Context, reqBody interface{}) (interface{}, error) {
				values := method.Func.Call([]reflect.Value{serviceValue, reflect.ValueOf(ctx), reflect.ValueOf(reqBody)})
				return values[0].Interface(), nil
			}
			return interceptor.ServerIntercept(ctx, req, ceps, handler, method.Name)
		}

		methods = append(methods, &service.MethodDescription{
			MethodName: method.Name,
			Handler:    methodHandler,
		})

	}
	return methods, nil
}

func checkMethod(method reflect.Type) error {
	if method.NumIn() < 3 {
		return fmt.Errorf("method %s invalid, the number of input params < 2", method.Name())
	}

	if method.NumOut() != 2 {
		return fmt.Errorf("method %s invalid, the number of return values != 2", method.Name())
	}

	ctxType := method.In(1)
	var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	if !ctxType.Implements(contextType) {
		return fmt.Errorf("method %s invalid, first param is not context", method.Name())
	}

	argType := method.In(2)
	if argType.Kind() != reflect.Ptr {
		return fmt.Errorf("method %s invalid, req type is not a pointer", method.Name())
	}

	replyType := method.Out(0)
	if replyType.Kind() != reflect.Ptr {
		return fmt.Errorf("method %s invalid, reply type is not a pointer", method.Name())
	}

	errType := method.Out(1)
	var errorType = reflect.TypeOf((*error)(nil)).Elem()
	if !errType.Implements(errorType) {
		return fmt.Errorf("method %s invalid, returns %s , not error", method.Name(), errType.Name())
	}

	return nil
}

func containPlugin(plugins []string, name string) bool {
	for _, p := range plugins {
		if p == name {
			return true
		}
	}
	return false
}
