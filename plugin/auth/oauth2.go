package auth

import (
	"alice_gorpc/interceptor"
	"alice_gorpc/metadata"
	"context"
	"errors"
	"golang.org/x/oauth2"
)

type OAuth2 struct {
	token *oauth2.Token
}

func NewOAuth2ByToken(token string) *OAuth2 {
	return &OAuth2{
		token: &oauth2.Token{
			AccessToken: token,
			TokenType:   "bearer",
		},
	}
}

func BuildAuthInterceptor() interceptor.ServerInterceptor {
	af := func(ctx context.Context) (context.Context, error) {
		md := metadata.ServerMetadata(ctx)

		if len(md) == 0 {
			return ctx, errors.New("token nil")
		}
		v := md["authorization"]
		if string(v) != "Bearer testToken" {
			return ctx, errors.New("token invalid")
		}
		return ctx, nil
	}
	return func(ctx context.Context, req interface{}, handler interceptor.Handler, methodName string) (interface{}, error) {
		newCtx, err := af(ctx)

		if err != nil {
			return nil, errors.New("oauth2 err:" + err.Error())
		}

		return handler(newCtx, req)
	}
}

func (o *OAuth2) GetMetadata(ctx context.Context, uri ...string) (map[string]string, error) {

	if o.token == nil {
		return nil, errors.New("get metadata err:token nil")
	}

	return map[string]string{
		"authorization": o.token.Type() + " " + o.token.AccessToken,
	}, nil
}
