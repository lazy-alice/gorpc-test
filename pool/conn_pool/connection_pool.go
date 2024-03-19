package conn_pool

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	opts  *Options
	conns *sync.Map
}

type ChannelPool struct {
	net.Conn
	initialCap  int
	maxCap      int
	maxIdle     int
	idleTimeout time.Duration
	dialTimeout time.Duration
	Dial        func(ctx context.Context) (net.Conn, error)
	conns       chan *Conn
	mu          sync.RWMutex
}

type Conn struct {
	net.Conn
	mu          sync.RWMutex
	unusable    bool
	createTime  time.Time
	dialTimeout time.Duration
	c           *ChannelPool
}

var poolMap = make(map[string]*Pool)

var DefaultPool = NewConnPool()

func init() {
	registerPool("default", DefaultPool)
}

func registerPool(poolName string, pool *Pool) {
	poolMap[poolName] = pool
}

func GetPool(poolName string) *Pool {
	if v, ok := poolMap[poolName]; ok {
		return v
	}
	return DefaultPool
}

func NewConnPool(opt ...Option) *Pool {
	// default options
	opts := &Options{
		maxCap:      1000,
		idleTimeout: 1 * time.Minute,
		dialTimeout: 200 * time.Millisecond,
	}
	m := &sync.Map{}

	p := &Pool{
		conns: m,
		opts:  opts,
	}
	for _, o := range opt {
		o(p.opts)
	}

	return p
}

func (p *Pool) Get(ctx context.Context, address string) (net.Conn, error) {
	if v, ok := p.conns.Load(address); ok {
		if cp, ok := v.(*ChannelPool); ok {
			conn, err := cp.Get(ctx)
			return conn, err
		}
	}

	cp, err := p.NewChannelPool(ctx, address)
	if err != nil {
		return nil, err
	}
	p.conns.Store(address, cp)
	return cp.Get(ctx)
}

func (p *Pool) NewChannelPool(ctx context.Context, address string) (*ChannelPool, error) {
	c := &ChannelPool{
		initialCap:  p.opts.initialCap,
		maxCap:      p.opts.maxCap,
		maxIdle:     p.opts.maxIdle,
		idleTimeout: p.opts.idleTimeout,
		dialTimeout: p.opts.dialTimeout,
		conns:       make(chan *Conn, p.opts.maxCap),
		Dial: func(ctx context.Context) (net.Conn, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			timeout := p.opts.dialTimeout
			if t, ok := ctx.Deadline(); ok {
				timeout = t.Sub(time.Now())
			}
			rawConn, err := net.DialTimeout("tcp", address, timeout)
			if err != nil {
				return nil, err
			}
			return rawConn, err
			//tAuth, err := auth.NewClientTLSAuthFromFile("./test_data/server.crt", "alice")
			//if err != nil {
			//	println("client tauth err:", err)
			//	return nil, err
			//}
			//tlsConn, err := tAuth.ServerHandshake(rawConn)
			//if err != nil {
			//	println("client tlsconn err:", err)
			//	return nil, err
			//}
			//return tlsConn, nil
		},
	}

	if p.opts.initialCap == 0 {
		// default initialCap is 1
		p.opts.initialCap = 1
	}

	for i := 0; i < len(c.conns); i++ {
		conn, err := c.Dial(ctx)
		if err != nil {
			return nil, err
		}
		err = c.Put(c.WrapConn(conn))
		if err != nil {
			return nil, err
		}
	}
	c.RegisterChecker(3*time.Second, func(conn *Conn) bool {
		if conn.createTime.Add(c.idleTimeout).Before(time.Now()) {
			return false
		}
		return true
	})
	return c, nil
}

func (p *ChannelPool) Put(conn *Conn) error {
	if conn == nil {
		return errors.New("connection close")
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.conns == nil {
		conn.MarkUnusable()
		_ = conn.Close()
		return errors.New("connection close")
	}

	select {
	case p.conns <- conn:
		//fmt.Printf("put conn:%s\n", conn.LocalAddr().String())
		return nil
	default:
		return conn.Close()
	}
}

func (p *ChannelPool) Get(ctx context.Context) (net.Conn, error) {
	if p.conns == nil {
		return nil, errors.New("connection close")
	}
	select {
	case c := <-p.conns:
		if c == nil {
			return nil, errors.New("connection close")
		}
		if c.unusable {
			return nil, errors.New("connection close")
		}
		//fmt.Printf("get conn:%s\n", c.Conn.LocalAddr().String())
		return c, nil
	default:
		conn, err := p.Dial(ctx)
		if err != nil {
			return nil, err
		}
		return p.WrapConn(conn), nil
	}
}

func (p *ChannelPool) WrapConn(conn net.Conn) *Conn {
	return &Conn{
		c:           p,
		Conn:        conn,
		unusable:    false,
		createTime:  time.Now(),
		dialTimeout: p.dialTimeout,
	}
}

func (p *ChannelPool) RegisterChecker(internal time.Duration, checker func(conn *Conn) bool) {
	if internal <= 0 || checker == nil {
		return
	}
	go func() {

		time.Sleep(internal)

		for i := 0; i < len(p.conns); i++ {
			select {
			case conn := <-p.conns:
				if !checker(conn) {
					conn.MarkUnusable()
					conn.Close()
					break
				} else {
					p.Put(conn)
				}
			default:
				break
			}
		}
	}()
}

func (c *Conn) MarkUnusable() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.unusable = true
}

func (c *Conn) Close() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.unusable {
		if c.Conn != nil {
			return c.Conn.Close()
		}
	}
	c.Conn.SetDeadline(time.Time{})
	return c.c.Put(c)
}
