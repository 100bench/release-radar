package github

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gh "github.com/google/go-github/v63/github"
	"github.com/mackb/releaseradar/pkg/logger"
	"github.com/mackb/releaseradar/pkg/retry"
)

type githubClient struct {
	client *gh.Client
}

func NewGitHubClient(httpClient *http.Client) Client {
	return &githubClient{
		client: gh.NewClient(httpClient),
	}
}

func (g *githubClient) GetLatestRelease(ctx context.Context, owner, repo string, etag string) (*Release, string, error) {
	var latestRelease *Release
	var newETag string

	err := retry.Do(3, 2*time.Second, func() error {
		rel, resp, err := g.client.Repositories.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotModified {
				// Content not modified, return nil release and original etag
				return nil
			}
			logger.L().Sugar().Errorf("failed to get latest release for %s/%s: %v", owner, repo, err)
			return fmt.Errorf("github client error: %w", err)
		}

		if resp != nil {
			newETag = resp.Header.Get("Etag")
		}

		if rel != nil {
			latestRelease = &Release{
				Tag:         rel.GetTagName(),
				Title:       rel.GetName(),
				URL:         rel.GetHTMLURL(),
				PublishedAt: rel.GetPublishedAt().Time,
				Body:        rel.GetBody(),
			}
		}
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return latestRelease, newETag, nil
}
