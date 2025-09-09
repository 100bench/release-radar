package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repo struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	ETag          string    `json:"etag"`
	LastCheckedAt time.Time `json:"last_checked_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Subscription struct {
	ID        uuid.UUID `json:"id"`
	RepoID    uuid.UUID `json:"repo_id"`
	UserID    uuid.UUID `json:"user_id"`
	Channel   string    `json:"channel"` // Telegram chat ID or similar
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Release struct {
	ID          uuid.UUID `json:"id"`
	RepoID      uuid.UUID `json:"repo_id"`
	Tag         string    `json:"tag"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	Hash        string    `json:"hash"` // Hash of release content for idempotency
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Delivery struct {
	ID        uuid.UUID `json:"id"`
	ReleaseID uuid.UUID `json:"release_id"`
	UserID    uuid.UUID `json:"user_id"`
	Channel   string    `json:"channel"`
	Status    string    `json:"status"` // e.g., "pending", "sent", "failed"
	Attempt   int       `json:"attempt"`
	LastError string    `json:"last_error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
