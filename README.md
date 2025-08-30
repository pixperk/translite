# Translite

A lightweight, transport-agnostic server/client library for Go that supports multiple protocols through a unified interface.

## Features

- **Multiple Transport Support**: TCP and WebSocket protocols
- **Unified Interface**: Same API for all transport types
- **Concurrent Handling**: Built-in goroutine management with proper synchronization
- **Event-Driven**: Handler-based architecture for connections, messages, and errors
- **Context Support**: Full context.Context integration for cancellation and timeouts
- **Thread-Safe**: Safe for concurrent use across multiple goroutines

## Architecture

The library follows a modular design with clear separation of concerns:

- **Transport Layer**: Abstract transport interface with concrete implementations (TCP, WebSocket)
- **Server/Client**: Generic server and client implementations that work with any transport
- **Connection Management**: Thread-safe connection tracking with proper cleanup
- **Event Handling**: Customizable handlers for connection lifecycle events

## Quick Start

### Creating a Server

```go
package main

import (
    "context"
    "log"
    "github.com/pixperk/translite/transport"
)

func main() {
    ctx := context.Background()
    
    // Create a TCP server
    server, err := transport.NewServer("tcp", ":8080", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Set up message handler
    server.OnMessage(func(conn transport.Conn, data []byte) {
        // Echo the message back
        conn.Write([]byte("echo: " + string(data)))
    })
    
    // Set up connection handler
    server.OnConnect(func(conn transport.Conn) {
        log.Printf("Client connected: %s", conn.ID())
    })
    
    // Start the server
    if err := server.Start(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Server is now running...
}
```

### Creating a Client

```go
package main

import (
    "context"
    "log"
    "github.com/pixperk/translite/transport"
)

func main() {
    ctx := context.Background()
    
    // Connect to a TCP server
    conn, err := transport.Dial(ctx, "tcp", "localhost:8080", nil)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    // Send a message
    _, err = conn.Write([]byte("hello world"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Read response
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Received: %s", string(buffer[:n]))
}
```

## Supported Transports

### TCP
```go
server, err := transport.NewServer("tcp", ":8080", nil)
conn, err := transport.Dial(ctx, "tcp", "localhost:8080", nil)
```

### WebSocket
```go
server, err := transport.NewServer("websocket", ":8080", nil)
conn, err := transport.Dial(ctx, "websocket", "localhost:8080", nil)
```

## Event Handlers

The server supports several event handlers:

```go
// Called when a new client connects
server.OnConnect(func(conn transport.Conn) {
    log.Printf("New connection: %s", conn.ID())
})

// Called when a message is received
server.OnMessage(func(conn transport.Conn, data []byte) {
    // Handle the message
})

// Called when a client disconnects
server.OnDisconnect(func(conn transport.Conn, err error) {
    log.Printf("Client disconnected: %s, reason: %v", conn.ID(), err)
})

// Called when an error occurs
server.OnError(func(err error) {
    log.Printf("Server error: %v", err)
})
```

## Concurrency and Synchronization

The library uses two key synchronization primitives:

- **`sync.WaitGroup`**: Tracks all active goroutines (accept loop and connection handlers) to ensure graceful shutdown
- **`sync.RWMutex`**: Protects the connections map for thread-safe access when adding/removing connections

This design ensures:
- No race conditions when managing connections
- Proper cleanup of all goroutines during shutdown
- Safe concurrent access to shared state

## Running Tests

```bash
go test ./...
```

The test suite includes comprehensive tests for both TCP and WebSocket transports, covering server/client communication scenarios.

## Dependencies

- Go 1.23.6+
- `github.com/gorilla/websocket` for WebSocket support


This is a practice project for learning Go concurrency patterns and network programming.
