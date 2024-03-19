package jaeger

import (
	"alice_gorpc/interceptor"
	"alice_gorpc/metadata"
	"alice_gorpc/plugin"
	"context"
	"errors"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go/config"
)

type Jaeger struct {
	opts *plugin.Options
}

const Name = "jaeger"
const JaegerClientName = "gorpc-client-jaeger"
const JaegerServerName = "gorpc-server-jaeger"

var JaegerSvr = &Jaeger{
	opts: &plugin.Options{},
}

func init() {
	plugin.Register(Name, JaegerSvr)
}

func (j *Jaeger) Init(opts ...plugin.Option) (opentracing.Tracer, error) {

	for _, o := range opts {
		o(j.opts)
	}

	if j.opts.TracingSvrAddr == "" {
		return nil, errors.New("jaeger init error, traingSvrAddr is empty")
	}

	return initJaeger(j.opts.TracingSvrAddr, JaegerServerName, opts...)

}

func initJaeger(tracingSvrAddr string, jaegerServiceName string, opts ...plugin.Option) (opentracing.Tracer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const", // Fixed sampling
			Param: 1,       // 1= full sampling, 0= no sampling
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: tracingSvrAddr,
		},
		ServiceName: jaegerServiceName,
	}

	tracer, _, err := cfg.NewTracer()
	if err != nil {
		fmt.Println("NewTracer err: ", err)
		return nil, err
	}

	opentracing.SetGlobalTracer(tracer)

	return tracer, nil
}

func OpenTracingServerInterceptor(tracer opentracing.Tracer, spanName string) interceptor.ServerInterceptor {

	return func(ctx context.Context, req interface{}, handler interceptor.Handler, methodName string) (interface{}, error) {

		serverMd := metadata.ServerMetadata(ctx)
		spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, serverMd)
		if err != nil {
			fmt.Printf("sever extract trace information error:%v\n", err)
		}

		serverSpan := tracer.StartSpan(spanName, ext.SpanKindRPCServer, ext.RPCServerOption(spanCtx))
		defer serverSpan.Finish()

		ctx = opentracing.ContextWithSpan(ctx, serverSpan)

		serverSpan.LogFields(log.String("spanName", spanName))

		return handler(ctx, req)
	}
}

func OpenTracingClientInterceptor(tracer opentracing.Tracer, spanName string) interceptor.ClientInterceptor {

	return func(ctx context.Context, req, rsp interface{}, ivk interceptor.Invoker) error {

		//var parentCtx opentracing.SpanContext
		//
		//if parent := opentracing.SpanFromContext(ctx); parent != nil {
		//	parentCtx = parent.Context()
		//}

		//clientSpan := tracer.StartSpan(spanName, ext.SpanKindRPCClient, opentracing.ChildOf(parentCtx))
		//md := metadata.CliMetadata{}
		md := metadata.ClientMetadata(ctx)
		//span := opentracing.SpanFromContext(ctx)
		//if span!=nil {
		//
		//}

		//ctx = metadata.WithClientMetadata(ctx,md)
		clientSpan := tracer.StartSpan(spanName, ext.SpanKindRPCClient)
		defer clientSpan.Finish()

		if err := tracer.Inject(clientSpan.Context(), opentracing.TextMap, md); err != nil {
			clientSpan.LogFields(log.String("event", "Tracer.Inject() failed"), log.Error(err))
		}

		ctx = metadata.WithClientMetadata(ctx, md)

		clientSpan.LogFields(log.String("spanName", spanName))
		fmt.Println("tracing 中间件执行")
		return ivk(ctx, req, rsp, spanName)

	}
}

func Init(tracingSvrAddr string, opts ...plugin.Option) (opentracing.Tracer, error) {
	return initJaeger(tracingSvrAddr, JaegerClientName, opts...)
}
