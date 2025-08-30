package transport

import (
	"context"
	"testing"
	"time"
)

func TestTransportRegistration(t *testing.T) {
	transports := ListTransports()

	expectedTransports := map[string]bool{
		"tcp":       false,
		"websocket": false,
	}

	for _, transport := range transports {
		if _, exists := expectedTransports[transport]; exists {
			expectedTransports[transport] = true
		}
	}

	for name, found := range expectedTransports {
		if !found {
			t.Errorf("Expected transport %s to be registered", name)
		}
	}
}

func TestTCPServerAndClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server, err := NewServer("tcp", ":0", nil)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	received := make(chan string, 1)

	server.OnMessage(func(conn Conn, data []byte) {
		received <- string(data)
		conn.Write([]byte("echo: " + string(data)))
	})

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	server.Stop(ctx)

	server, err = NewServer("tcp", ":18080", nil)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	server.OnMessage(func(conn Conn, data []byte) {
		received <- string(data)
		conn.Write([]byte("echo: " + string(data)))
	})

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	conn, err := Dial(ctx, "tcp", "localhost:18080", nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	testMessage := "hello world"
	if _, err := conn.Write([]byte(testMessage)); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	select {
	case msg := <-received:
		if msg != testMessage {
			t.Errorf("Expected %q, got %q", testMessage, msg)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message")
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	expected := "echo: " + testMessage
	if string(buffer[:n]) != expected {
		t.Errorf("Expected %q, got %q", expected, string(buffer[:n]))
	}
}

func TestWSServerAndClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	server, err := NewServer("websocket", ":0", nil)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	received := make(chan string, 1)

	server.OnMessage(func(conn Conn, data []byte) {
		received <- string(data)
		conn.Write([]byte("echo: " + string(data)))
	})

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	server.Stop(ctx)

	server, err = NewServer("websocket", ":18081", nil)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	server.OnMessage(func(conn Conn, data []byte) {
		received <- string(data)
		conn.Write([]byte("echo: " + string(data)))
	})

	if err := server.Start(ctx); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop(ctx)

	time.Sleep(100 * time.Millisecond)

	conn, err := Dial(ctx, "websocket", "ws://localhost:18081", nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	testMessage := "hello ws"
	if _, err := conn.Write([]byte(testMessage)); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	select {
	case msg := <-received:
		if msg != testMessage {
			t.Errorf("Expected %q, got %q", testMessage, msg)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message")
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	expected := "echo: " + testMessage
	if string(buffer[:n]) != expected {
		t.Errorf("Expected %q, got %q", expected, string(buffer[:n]))
	}
}
