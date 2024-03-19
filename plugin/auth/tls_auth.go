package auth

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"syscall"
)

type tlsAuth struct {
	config *tls.Config
	state  tls.ConnectionState
}

func NewServerTLSAuthFromFile(certFile, keyFile string) (*tlsAuth, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.New("load cert error:" + err.Error())
	}
	conf := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return &tlsAuth{config: conf}, nil
}

func (t *tlsAuth) ServerHandshake(rawConn net.Conn) (net.Conn, error) {
	conn := tls.Server(rawConn, t.config)
	if err := conn.Handshake(); err != nil {
		return nil, err
	}
	return WrapConn(rawConn, conn), nil
}

func WrapConn(rawConn, newConn net.Conn) net.Conn {
	sysConn, ok := rawConn.(syscall.Conn)
	if !ok {
		return newConn
	}
	return &wrapperConn{
		Conn:    newConn,
		sysConn: sysConn,
	}
}

type sysConn = syscall.Conn

type wrapperConn struct {
	net.Conn
	// sysConn is a type alias of syscall.Conn. It's necessary because the name
	// `Conn` collides with `net.Conn`.
	sysConn
}

func NewClientTLSAuthFromFile(certFile, serverName string) (*tlsAuth, error) {
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(cert) {
		return nil, errors.New("append cert error:" + err.Error())
	}
	conf := &tls.Config{
		ServerName: serverName,
		RootCAs:    cp,
	}
	return &tlsAuth{config: conf}, nil
}
