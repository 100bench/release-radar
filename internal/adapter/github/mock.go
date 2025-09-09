package github

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockGitHubClient struct {
	mock.Mock
}

func (m *MockGitHubClient) GetLatestRelease(ctx context.Context, owner, repo, etag string) (*Release, string, error) {
	args := m.Called(ctx, owner, repo, etag)
	var r *Release
	if args.Get(0) != nil {
		r = args.Get(0).(*Release)
	}
	return r, args.String(1), args.Error(2)
}
