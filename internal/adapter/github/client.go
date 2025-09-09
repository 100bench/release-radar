package github

import (
	"context"
	"time"
)

type Release struct {
	Tag         string
	Title       string
	URL         string
	PublishedAt time.Time
	Body        string // For calculating hash
}

type Client interface {
	GetLatestRelease(ctx context.Context, owner, repo string, etag string) (*Release, string, error)
}
