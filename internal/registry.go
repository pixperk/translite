package internal

import (
	"fmt"
	"strings"
	"sync"
)

type Registry struct {
	mu         sync.RWMutex
	transports map[string]Transport
}

func NewRegistry() *Registry {
	return &Registry{
		transports: make(map[string]Transport),
	}
}

func (r *Registry) Register(name string, transport Transport) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = strings.ToLower(name)
	if _, exists := r.transports[name]; exists {
		return fmt.Errorf("transport %s is already registered", name)
	}

	r.transports[name] = transport
	return nil
}

func (r *Registry) Get(name string) (Transport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name = strings.ToLower(name)
	transport, exists := r.transports[name]
	if !exists {
		return nil, fmt.Errorf("transport %s is not registered", name)
	}

	return transport, nil
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.transports))
	for name := range r.transports {
		names = append(names, name)
	}

	return names
}

func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = strings.ToLower(name)
	if _, exists := r.transports[name]; !exists {
		return fmt.Errorf("transport %s is not registered", name)
	}

	delete(r.transports, name)
	return nil
}

var defaultRegistry = NewRegistry()

func RegisterTransport(name string, transport Transport) error {
	return defaultRegistry.Register(name, transport)
}

func GetTransport(name string) (Transport, error) {
	return defaultRegistry.Get(name)
}

func ListTransports() []string {
	return defaultRegistry.List()
}
