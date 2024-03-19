package client

import (
	"alice_gorpc/interceptor"
	"alice_gorpc/metadata"
	"alice_gorpc/msg"
	"alice_gorpc/plugin/auth"
	"alice_gorpc/pool/conn_pool"
	"alice_gorpc/protocol"
	"alice_gorpc/selector"
	"alice_gorpc/transport"
	"alice_gorpc/utils"
	"context"
	"encoding/json"
	"errors"
)

type Client struct {
	opts *Options
}

func NewClient() *Client {
	return &Client{opts: &Options{
		perRPCAuth: make([]*auth.OAuth2, 0),
	}}
}

func (c *Client) Call(ctx context.Context, servicePath string, req interface{}, rsp interface{}, opts ...Option) error {
	return c.Invoke(ctx, servicePath, req, rsp, opts...)
}

func (c *Client) Invoke(ctx context.Context, servicePath string, req interface{}, rsp interface{}, opts ...Option) error {
	for _, o := range opts {
		o(c.opts)
	}

	if c.opts.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.opts.timeout)
		defer cancel()
	}

	serviceName, _, err := utils.ParseServicePath(servicePath)
	if err != nil {
		return err
	}
	c.opts.serviceName = serviceName
	//c.opts.method = method
	return interceptor.ClientIntercept(ctx, req, rsp, c.opts.interceptors, c.invoke, servicePath)
}

func (c *Client) invoke(ctx context.Context, req interface{}, rsp interface{}, servicePath string) error {
	reqBuf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	request := addReqHeader(ctx, servicePath, reqBuf, c)
	reqBuf, err = json.Marshal(request)
	if err != nil {
		return err
	}
	codec := protocol.NewCodec()
	frame, err := codec.Encode(reqBuf)
	if err != nil {
		return err
	}

	transportOptions := []transport.ClientOption{
		transport.WithTargetServiceName(c.opts.serviceName),
		transport.WithSelector(selector.GetSelector(c.opts.selectorName)),
		transport.WithPool(conn_pool.GetPool("default")),
	}
	clientTransport := transport.GetClientTransport(c.opts.serviceName)
	rspFrame, err := clientTransport.Send(ctx, frame, transportOptions...)
	if err != nil {
		return err
	}

	rspBodyBuf := codec.Decode(rspFrame)
	response := &msg.Response{}
	err = json.Unmarshal(rspBodyBuf, response)
	if response.RetCode != 0 {
		return errors.New("call error :" + response.RetMsg)
	}

	err = json.Unmarshal(response.Payload, rsp)
	if err != nil {
		return err
	}
	return nil
}

func addReqHeader(ctx context.Context, servicePath string, payload []byte, client *Client) *msg.Request {
	md := metadata.ClientMetadata(ctx)
	for _, pra := range client.opts.perRPCAuth {
		authMd, _ := pra.GetMetadata(ctx)
		for k, v := range authMd {
			md[k] = []byte(v)
		}
	}
	return &msg.Request{
		ServicePath: servicePath,
		Metadata:    md,
		Payload:     payload,
	}
}
