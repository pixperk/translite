package transports

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/pixperk/translite/internal"
)

type TCPTransport struct{} //implements Transport Interface

func NewTCPTransport() *TCPTransport {
	return &TCPTransport{}
}

func (t *TCPTransport) Name() string {
	return "tcp"
}

func (t *TCPTransport) Listen(ctx context.Context, addr string, opts *internal.ListenOptions) (internal.Listener, error) {
	if opts == nil {
		opts = &internal.ListenOptions{}
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s : %w", addr, err)
	}

	if opts.TLSConfig != nil {
		if tlsConfig, ok := opts.TLSConfig.(*tls.Config); ok {
			listener = tls.NewListener(listener, tlsConfig)
		}
	}

	return &TCPListener{
		listener: listener,
		opts:     opts,
	}, nil
}

func (t *TCPTransport) Dial(ctx context.Context, addr string, opts *internal.DialOptions) (internal.Conn, error) {
	if opts == nil {
		opts = &internal.DialOptions{
			Timeout: 30 * time.Second,
		}
	}

	dialer := &net.Dialer{
		Timeout:   opts.Timeout,
		KeepAlive: opts.KeepAlive,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s : %w", addr, err)
	}

	if opts.TLSConfig != nil {
		if tlsConfig, ok := opts.TLSConfig.(*tls.Config); ok {
			tlsConn := tls.Client(conn, tlsConfig)
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				conn.Close()
				return nil, fmt.Errorf("TLS handshake failed: %w", err)
			}
			conn = tlsConn
		}
	}

	return newTCPConn(conn), nil

}

type TCPListener struct {
	listener net.Listener
	opts     *internal.ListenOptions
}

func (l *TCPListener) Accept() (internal.Conn, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok && l.opts.KeepAlive > 0 {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(l.opts.KeepAlive)
	}

	return newTCPConn(conn), nil
}

func (l *TCPListener) Close() error {
	return l.listener.Close()
}

func (l *TCPListener) Addr() net.Addr {
	return l.listener.Addr()
}

type TCPConn struct {
	conn net.Conn
	id   string
}

func newTCPConn(conn net.Conn) *TCPConn {
	id := generateConnID()
	return &TCPConn{
		conn: conn,
		id:   id,
	}
}

func (c *TCPConn) ID() string {
	return c.id
}

func (c *TCPConn) Read(b []byte) (int, error) {
	return c.conn.Read(b)
}

func (c *TCPConn) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *TCPConn) Close() error {
	return c.conn.Close()
}

func (c *TCPConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *TCPConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *TCPConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *TCPConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *TCPConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func generateConnID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("tcp_%x", b)
}
