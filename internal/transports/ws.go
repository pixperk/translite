package transports

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pixperk/translite/internal"
)

type WSTransport struct {
	upgrader websocket.Upgrader
}

func NewWSTransport() *WSTransport {
	return &WSTransport{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true //allow from any origin
			},
		},
	}
}

func (t *WSTransport) Name() string {
	return "websocket"
}

func (t *WSTransport) Listen(ctx context.Context, addr string, opts *internal.ListenOptions) (internal.Listener, error) {
	if opts == nil {
		opts = &internal.ListenOptions{}
	}

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s : %w", addr, err)
	}

	wsListener := &WSListener{
		listener:  lis,
		transport: t,
		opts:      opts,
		connChan:  make(chan internal.Conn, 10),
		errChan:   make(chan error, 1),
	}

	go wsListener.serve()

	return wsListener, nil
}

func (t *WSTransport) Dial(ctx context.Context, addr string, opts *internal.DialOptions) (internal.Conn, error) {
	if opts == nil {
		opts = &internal.DialOptions{
			Timeout: 30 * time.Second,
		}
	}

	u, err := url.Parse(addr)
	if err != nil {
		//assuming it is host:port
		u = &url.URL{
			Scheme: "ws",
			Host:   addr,
			Path:   "/",
		}
	}

	if u.Scheme == "" {
		u.Scheme = "ws"
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: opts.Timeout,
	}

	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to dial WebSocket %s: %w", addr, err)
	}

	return newWebSocketConn(conn), nil
}

type WSListener struct {
	listener  net.Listener
	transport *WSTransport
	opts      *internal.ListenOptions
	connChan  chan internal.Conn
	errChan   chan error
	server    *http.Server
}

func (l *WSListener) serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", l.handleWebSocket)

	l.server = &http.Server{
		Handler: mux,
	}

	if err := l.server.Serve(l.listener); err != nil && err != http.ErrServerClosed {
		select {
		case l.errChan <- err:
		default:
		}
	}
}

func (l *WSListener) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := l.transport.upgrader.Upgrade(w, r, nil)
	if err != nil {
		select {
		case l.errChan <- fmt.Errorf("WebSocket upgrade failed: %w", err):
		default:
		}
		return
	}

	wsConn := newWebSocketConn(conn)
	select {
	case l.connChan <- wsConn:
	default:
		conn.Close()
	}
}

func (l *WSListener) Accept() (internal.Conn, error) {
	select {
	case conn := <-l.connChan:
		return conn, nil
	case err := <-l.errChan:
		return nil, err
	}
}
func (l *WSListener) Close() error {
	if l.server != nil {
		return l.server.Close()
	}
	return l.listener.Close()
}

func (l *WSListener) Addr() net.Addr {
	return l.listener.Addr()
}

type WebSocketConn struct {
	conn *websocket.Conn
	id   string
}

func newWebSocketConn(conn *websocket.Conn) *WebSocketConn {
	id := generateWSConnID()
	return &WebSocketConn{
		conn: conn,
		id:   id,
	}
}

func (c *WebSocketConn) ID() string {
	return c.id
}

func (c *WebSocketConn) Read(b []byte) (int, error) {
	_, data, err := c.conn.ReadMessage()
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			return 0, io.EOF
		}
		return 0, err
	}
	n := copy(b, data)
	if n < len(data) {
		return n, fmt.Errorf("buffer too small: need %d bytes, got %d", len(data), len(b))
	}

	return n, nil

}

func (c *WebSocketConn) Write(b []byte) (int, error) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *WebSocketConn) Close() error {
	return c.conn.Close()
}

func (c *WebSocketConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *WebSocketConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *WebSocketConn) SetDeadline(t time.Time) error {
	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}
	return c.conn.SetWriteDeadline(t)
}

func (c *WebSocketConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *WebSocketConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func generateWSConnID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("ws_%x", b)
}
