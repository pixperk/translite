package internal

import (
	"context"
	"net"
	"time"
)

type Transport interface {
	Listen(ctx context.Context, addr string, opts *ListenOptions) (Listener, error)
	Dial(ctx context.Context, addr string, opts *DialOptions) (Conn, error)
	Name() string
}

type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() net.Addr
}

type Conn interface {
	ID() string
	RemoteAddr() net.Addr
	LocalAddr() net.Addr
	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	OnConnect(handler ConnectHandler)
	OnMessage(handler MessageHandler)
	OnDisconnect(handler DisconnectHandler)
	OnError(handler ErrorHandler)
}

type ConnectHandler func(conn Conn)
type MessageHandler func(conn Conn, data []byte)
type DisconnectHandler func(conn Conn, err error)
type ErrorHandler func(err error)

type ListenOptions struct {
	TLSConfig       any
	KeepAlive       time.Duration
	ReadBufferSize  int
	WriteBufferSize int
	Extra           map[string]any
}

type DialOptions struct {
	Timeout   time.Duration
	TLSConfig any
	KeepAlive time.Duration
	Extra     map[string]any
}
