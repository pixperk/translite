package internal

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type GenericServer struct {
	transport Transport
	listener  Listener
	addr      string
	opts      *ListenOptions

	//Event handlers
	connectHandler    ConnectHandler
	messageHandler    MessageHandler
	errorHandler      ErrorHandler
	disconnectHandler DisconnectHandler

	//State
	running   int32
	wg        sync.WaitGroup
	cancel    context.CancelFunc
	mu        sync.RWMutex
	conns     map[string]Conn
	connCount int64
}

func NewServer(transport Transport, addr string, opts *ListenOptions) *GenericServer {
	if opts == nil {
		opts = &ListenOptions{}
	}
	return &GenericServer{
		transport: transport,
		addr:      addr,
		opts:      opts,
		conns:     make(map[string]Conn),
	}
}

func (s *GenericServer) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return fmt.Errorf("server is already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	lis, err := s.transport.Listen(ctx, s.addr, s.opts)
	if err != nil {
		atomic.StoreInt32(&s.running, 0)
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.listener = lis
	s.wg.Add(1)
	go s.acceptLoop(ctx)

	return nil
}

func (s *GenericServer) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&s.running, 1, 0) {
		return fmt.Errorf("server is not running")
	}

	if s.cancel != nil {
		s.cancel()
	}

	if s.listener != nil {
		s.listener.Close()
	}

	s.mu.Lock()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Event handler setters
func (s *GenericServer) OnConnect(handler ConnectHandler) {
	s.connectHandler = handler
}

func (s *GenericServer) OnMessage(handler MessageHandler) {
	s.messageHandler = handler
}

func (s *GenericServer) OnDisconnect(handler DisconnectHandler) {
	s.disconnectHandler = handler
}

func (s *GenericServer) OnError(handler ErrorHandler) {
	s.errorHandler = handler
}

func (s *GenericServer) GetConnectionCount() int64 {
	return atomic.LoadInt64(&s.connCount)
}

func (s *GenericServer) acceptLoop(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if atomic.LoadInt32(&s.running) == 0 {
				return
			}
			if s.errorHandler != nil {
				s.errorHandler(fmt.Errorf("accept error: %w", err))
			}
			continue
		}

		s.addConnection(conn)
		s.wg.Add(1)
		go s.handleConnection(ctx, conn)
	}
}

func (s *GenericServer) handleConnection(ctx context.Context, conn Conn) {
	defer s.wg.Done()
	defer s.removeConnection(conn)
	defer conn.Close()

	if s.connectHandler != nil {
		s.connectHandler(conn)
	}

	buffer := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, err := conn.Read(buffer)
		if err != nil {
			if s.disconnectHandler != nil {
				s.disconnectHandler(conn, err)
			}
			return
		}

		if n > 0 && s.messageHandler != nil {
			data := make([]byte, n)
			copy(data, buffer[:n])
			s.messageHandler(conn, data)
		}
	}
}

func (s *GenericServer) addConnection(conn Conn) {
	s.mu.Lock()
	s.conns[conn.ID()] = conn
	s.mu.Unlock()
	atomic.AddInt64(&s.connCount, 1)
}

func (s *GenericServer) removeConnection(conn Conn) {
	s.mu.Lock()
	delete(s.conns, conn.ID())
	s.mu.Unlock()
	atomic.AddInt64(&s.connCount, -1)
}
