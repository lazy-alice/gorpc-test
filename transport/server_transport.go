package transport

import (
	"alice_gorpc/metadata"
	"alice_gorpc/msg"
	"alice_gorpc/protocol"
	"alice_gorpc/utils"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
)

type ServerTransport struct {
	opts *ServerTransportOptions
}

var defaultServerTransport = &ServerTransport{
	opts: &ServerTransportOptions{},
}

var serverTransPortMap = make(map[string]*ServerTransport)

func init() {
	serverTransPortMap["default"] = defaultServerTransport
}

func NewTransport(name string, opts *ServerTransportOptions) *ServerTransport {
	return &ServerTransport{
		opts: opts,
	}
}

func GetServerTransport(name string) *ServerTransport {
	if trans, ok := serverTransPortMap[name]; ok {
		return trans
	}
	return defaultServerTransport
}

func (t *ServerTransport) ListenAndServe(ctx context.Context, opts ...ServerOption) error {
	for _, opt := range opts {
		opt(t.opts)
	}
	lis, err := net.Listen("tcp", t.opts.Address)
	if err != nil {
		return err
	}
	go func() {
		if err = t.Serve(ctx, lis); err != nil {
			log.Default().Println("transport serve error:", err)
		}
	}()
	return nil
}

func (t *ServerTransport) Serve(ctx context.Context, lis net.Listener) error {
	tl, ok := lis.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn, err := tl.AcceptTCP()
		if err != nil {
			return err
		}
		if err = conn.SetKeepAlive(true); err != nil {
			return err
		}
		if t.opts.KeepAlivePeriod != 0 {
			_ = conn.SetKeepAlivePeriod(t.opts.KeepAlivePeriod)
		}

		//tlsAuth, err := auth.NewServerTLSAuthFromFile("./test_data/server.crt", "./test_data/server.key")
		//if err != nil {
		//	return err
		//}
		//tlsConn, err := tlsAuth.ServerHandshake(conn)
		//if err != nil {
		//	return err
		//}
		//println("conn in .....", tlsConn.LocalAddr())
		go func() {
			if err = t.handleConn(ctx, wrapConn(conn)); err != nil {
				//log.Printf("gorpc handle tcp conn error, %v", err)
			}
		}()

	}
}

type connWrapper struct {
	net.Conn
	frame *protocol.Frame
}

func wrapConn(conn net.Conn) *connWrapper {
	return &connWrapper{
		Conn:  conn,
		frame: protocol.NewFrame(),
	}
}

func (t *ServerTransport) handle(ctx context.Context, frame []byte) ([]byte, error) {
	codec := protocol.NewCodec()
	reqBody := codec.Decode(frame)

	request := &msg.Request{}
	if err := json.Unmarshal(reqBody, request); err != nil {
		return nil, err
	}

	serviceName, _, err := utils.ParseServicePath(request.ServicePath)
	if err != nil {
		return nil, err
	}
	service := t.opts.Handlers[serviceName]
	rsp, err := service.Handle(ctx, request)
	if err != nil {
		return nil, err
	}
	rspBuf, err := json.Marshal(rsp)
	if err != nil {
		return nil, err
	}
	rsp = addRspHeader(ctx, rspBuf, err)
	response, err := json.Marshal(rsp)
	if err != nil {
		return nil, err
	}
	response, err = codec.Encode(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func addRspHeader(ctx context.Context, payload []byte, err error) *msg.Response {
	rsp := &msg.Response{
		RetCode:  0,
		RetMsg:   "success",
		Metadata: metadata.ServerMetadata(ctx),
		Payload:  payload,
	}
	if err != nil {
		rsp.RetCode = 500
		rsp.RetMsg = "server internal error"
	}
	return rsp
}

func (t *ServerTransport) handleConn(ctx context.Context, conn *connWrapper) error {
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		frame, err := conn.frame.ReadFrame(conn)

		if err != nil {
			if err == io.EOF {
				continue
				//return nil
			}
			return err
		}
		rsp, err := t.handle(ctx, frame)
		if err != nil {
			return err
		}
		if _, err = conn.Write(rsp); err != nil {
			return err
		}
	}
}
