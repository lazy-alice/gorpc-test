package transport

import (
	"context"
	"errors"
	"net"
)

var clientTransportMap = make(map[string]*ClientTransport)

type ClientTransport struct {
	opts *ClientTransportOptions
}

func NewClientTransport(opts ...ClientOption) *ClientTransport {
	clientTransport := &ClientTransport{opts: &ClientTransportOptions{}}
	for _, o := range opts {
		o(clientTransport.opts)
	}
	return clientTransport
}

func GetClientTransport(targetServiceName string) *ClientTransport {
	if clientTransport, ok := clientTransportMap[targetServiceName]; ok {
		return clientTransport
	}
	clientTransport := &ClientTransport{opts: &ClientTransportOptions{}}
	clientTransportMap[targetServiceName] = clientTransport
	return clientTransport
}

func (c *ClientTransport) Send(ctx context.Context, reqFrame []byte, opts ...ClientOption) ([]byte, error) {
	for _, o := range opts {
		o(c.opts)
	}
	addr, err := c.opts.selector.Select(c.opts.targetServiceName)
	if err != nil {
		return nil, err
	}
	if addr == "" {
		return nil, errors.New("target address is nil")
	}

	//conn, err := c.opts.pool.Get(ctx, addr)
	//println("client conn....", conn.LocalAddr())
	conn, err := net.Dial("tcp", "127.0.0.1:49999")
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sendNum := 0
	num := 0
	for sendNum < len(reqFrame) {
		num, err = conn.Write(reqFrame[sendNum:])
		if err != nil {
			return nil, err
		}
		sendNum += num
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	wC := wrapConn(conn)
	rspFrame, err := wC.frame.ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	return rspFrame, nil
}
