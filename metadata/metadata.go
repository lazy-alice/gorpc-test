package metadata

import (
	"context"
	"strings"
)

type clientMD struct{}

type serverMD struct{}

type CliMetadata map[string][]byte

type SvrMetadata map[string][]byte

func (m SvrMetadata) Set(key, val string) {
	key = strings.ToLower(key)
	m[key] = []byte(val)
}

func (m SvrMetadata) ForeachKey(handler func(key, val string) error) error {
	for k, v := range m {
		handler(k, string(v))
	}
	return nil
}

func (m CliMetadata) Set(key, val string) {
	key = strings.ToLower(key)
	m[key] = []byte(val)
}

func (m CliMetadata) ForeachKey(handler func(key, val string) error) error {
	for k, v := range m {
		handler(k, string(v))
	}
	return nil
}

func ClientMetadata(ctx context.Context) CliMetadata {
	if md, ok := ctx.Value(clientMD{}).(CliMetadata); ok {
		return md
	}
	md := make(CliMetadata)
	//WithClientMetadata(ctx, md)
	return md
}

func WithClientMetadata(ctx context.Context, md CliMetadata) context.Context {
	return context.WithValue(ctx, clientMD{}, md)
}

func ServerMetadata(ctx context.Context) SvrMetadata {
	if md, ok := ctx.Value(serverMD{}).(SvrMetadata); ok {
		return md
	}
	md := make(SvrMetadata)
	//WithServerMetadata(ctx, md)
	return md
}

func WithServerMetadata(ctx context.Context, md SvrMetadata) context.Context {
	return context.WithValue(ctx, serverMD{}, md)
}
