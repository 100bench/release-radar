package idempotency

import (
	"context"
	"fmt"
	"time"
)

type Storage interface {
	CheckAndSet(ctx context.Context, key string, expiration time.Duration) (bool, error)
}

type Manager struct {
	storage Storage
}

func NewManager(storage Storage) *Manager {
	return &Manager{storage: storage}
}

func (m *Manager) Do(ctx context.Context, key string, expiration time.Duration, fn func() error) error {
	set, err := m.storage.CheckAndSet(ctx, key, expiration)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}

	if !set {
		// Operation already in progress or completed
		return nil // Or return a specific idempotency error if needed
	}

	return fn()
}
