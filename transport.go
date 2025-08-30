package transport

import (
	"context"
	"fmt"

	"github.com/pixperk/translite/internal"
	"github.com/pixperk/translite/internal/transports"
)

// Re-export core interfaces and types
type (
	Transport         = internal.Transport
	Listener          = internal.Listener
	Conn              = internal.Conn
	Server            = internal.Server
	ListenOptions     = internal.ListenOptions
	DialOptions       = internal.DialOptions
	ConnectHandler    = internal.ConnectHandler
	MessageHandler    = internal.MessageHandler
	DisconnectHandler = internal.DisconnectHandler
	ErrorHandler      = internal.ErrorHandler
)

func init() {
	internal.RegisterTransport("tcp", transports.NewTCPTransport())
	internal.RegisterTransport("websocket", transports.NewWSTransport())
}

func NewServer(transportName, addr string, opts *ListenOptions) (Server, error) {
	transport, err := internal.GetTransport(transportName)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport %s: %w", transportName, err)
	}

	return internal.NewServer(transport, addr, opts), nil
}

func Dial(ctx context.Context, transportName, addr string, opts *DialOptions) (Conn, error) {
	transport, err := internal.GetTransport(transportName)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport %s: %w", transportName, err)
	}

	return transport.Dial(ctx, addr, opts)
}

func Listen(ctx context.Context, transportName, addr string, opts *ListenOptions) (Listener, error) {
	transport, err := internal.GetTransport(transportName)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport %s: %w", transportName, err)
	}

	return transport.Listen(ctx, addr, opts)
}
func RegisterTransport(name string, transport Transport) error {
	return internal.RegisterTransport(name, transport)
}

func GetTransport(name string) (Transport, error) {
	return internal.GetTransport(name)
}

func ListTransports() []string {
	return internal.ListTransports()
}
