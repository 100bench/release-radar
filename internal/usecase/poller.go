package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mackb/releaseradar/internal/adapter/github"
	"github.com/mackb/releaseradar/internal/adapter/persistence"
	"github.com/mackb/releaseradar/internal/domain"
	"github.com/mackb/releaseradar/pkg/idempotency"
	"github.com/mackb/releaseradar/pkg/logger"
)

type pollerUseCase struct {
	repoStore     persistence.RepoRepository
	releaseStore  persistence.ReleaseRepository
	subStore      persistence.SubscriptionRepository
	userStore     persistence.UserRepository     // Добавлено
	deliveryStore persistence.DeliveryRepository // Добавлено
	githubClient  github.Client
	transactor    persistence.Transactor
}

func NewPollerUseCase(repoStore persistence.RepoRepository, releaseStore persistence.ReleaseRepository, subStore persistence.SubscriptionRepository, userStore persistence.UserRepository, deliveryStore persistence.DeliveryRepository, githubClient github.Client, transactor persistence.Transactor) PollerUseCase {
	return &pollerUseCase{
		repoStore:     repoStore,
		releaseStore:  releaseStore,
		subStore:      subStore,
		userStore:     userStore,     // Добавлено
		deliveryStore: deliveryStore, // Добавлено
		githubClient:  githubClient,
		transactor:    transactor,
	}
}

func (p *pollerUseCase) PollReleases(ctx context.Context) error {
	const op = "PollerUseCase.PollReleases"
	logger.L().Sugar().Debugf("%s: starting release polling cycle", op)

	// In a real application, this would iterate through all repos that need polling
	// For a stub, we'll simulate polling for one existing repo (or create one)

	// --- STUB: Fetch a dummy repo or create one if none exists ---
	var repo *domain.Repo
	repos, err := p.repoStore.ListRepos(ctx, uuid.Nil) // Assuming a method to list all repos or a specific user's repos
	if err != nil && !errors.Is(err, persistence.ErrNotFound) {
		return fmt.Errorf("%s: failed to list repos: %w", op, err)
	}

	if len(repos) == 0 {
		// Create a dummy user and repo if none exist
		logger.L().Sugar().Warnf("%s: no repos found, creating a dummy user and repo for polling", op)
		user := &domain.User{
			ID:        uuid.New(),
			Email:     "dummy@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = p.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := p.userStore.CreateUser(txCtx, user); err != nil { // Исправлено
				return fmt.Errorf("%s: failed to create dummy user: %w", op, err)
			}
			dummyRepo := &domain.Repo{
				ID:            uuid.New(),
				UserID:        user.ID,
				Owner:         "golang",
				Name:          "go",
				ETag:          "",
				LastCheckedAt: time.Now().Add(-24 * time.Hour), // Set to past to trigger initial check
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := p.repoStore.CreateRepo(txCtx, dummyRepo); err != nil {
				return fmt.Errorf("%s: failed to create dummy repo: %w", op, err)
			}
			repo = dummyRepo
			return nil
		})
		if err != nil {
			return fmt.Errorf("%s: failed to set up dummy repo for polling: %w", op, err)
		}
	} else {
		repo = &repos[0] // Just pick the first one for the stub
	}
	// --- END STUB ---

	logger.L().Sugar().Infof("%s: polling repo %s/%s (ID: %s)", op, repo.Owner, repo.Name, repo.ID)

	githubRelease, newETag, err := p.githubClient.GetLatestRelease(ctx, repo.Owner, repo.Name, repo.ETag)
	if err != nil {
		return fmt.Errorf("%s: failed to get latest release from GitHub for %s/%s: %w", op, repo.Owner, repo.Name, err)
	}

	if githubRelease == nil {
		logger.L().Sugar().Infof("%s: no new release or content not modified for %s/%s", op, repo.Owner, repo.Name)
	} else {
		// Calculate hash of release content (e.g., body + tag + title + url)
		hash := sha256.New()
		hash.Write([]byte(githubRelease.Body + githubRelease.Tag + githubRelease.Title + githubRelease.URL))
		releaseHash := hex.EncodeToString(hash.Sum(nil))

		existingRelease, err := p.releaseStore.GetReleaseByRepoIDAndTag(ctx, repo.ID, githubRelease.Tag)
		if err != nil && !errors.Is(err, persistence.ErrNotFound) {
			return fmt.Errorf("%s: failed to check for existing release: %w", op, err)
		}

		// Only create a new release if it's genuinely new or content has changed
		if existingRelease == nil || existingRelease.Hash != releaseHash {
			newRelease := &domain.Release{
				ID:          uuid.New(),
				RepoID:      repo.ID,
				Tag:         githubRelease.Tag,
				Title:       githubRelease.Title,
				URL:         githubRelease.URL,
				PublishedAt: githubRelease.PublishedAt,
				Hash:        releaseHash,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if err := p.releaseStore.CreateRelease(ctx, newRelease); err != nil {
				return fmt.Errorf("%s: failed to create new release: %w", op, err)
			}
			logger.L().Sugar().Infof("%s: new release %s for %s/%s", op, newRelease.Tag, repo.Owner, repo.Name)

			// Enqueue deliveries for this new release
			if err := p.EnqueueDeliveries(ctx, newRelease); err != nil {
				logger.L().Sugar().Errorf("%s: failed to enqueue deliveries for release %s: %v", op, newRelease.ID, err)
			}
		} else {
			logger.L().Sugar().Infof("%s: release %s for %s/%s already exists with same content", op, githubRelease.Tag, repo.Owner, repo.Name)
		}
	}

	// Update repo ETag and LastCheckedAt
	if newETag != "" && newETag != repo.ETag {
		repo.ETag = newETag
	}
	repo.LastCheckedAt = time.Now()
	if err := p.repoStore.UpdateRepo(ctx, repo); err != nil {
		return fmt.Errorf("%s: failed to update repo %s ETag/LastCheckedAt: %w", op, repo.ID, err)
	}

	logger.L().Sugar().Debugf("%s: finished release polling cycle", op)
	return nil
}

func (p *pollerUseCase) EnqueueDeliveries(ctx context.Context, release *domain.Release) error {
	const op = "PollerUseCase.EnqueueDeliveries"
	logger.L().Sugar().Debugf("%s: enqueuing deliveries for release %s", op, release.ID)

	subscriptions, err := p.subStore.ListSubscriptionsByRepoID(ctx, release.RepoID)
	if err != nil {
		return fmt.Errorf("%s: failed to get subscriptions for repo %s: %w", op, release.RepoID, err)
	}

	for _, sub := range subscriptions {
		// Check for idempotency for this specific delivery
		deliveryKey := fmt.Sprintf("delivery:%s:%s:%s", release.ID, sub.UserID, sub.Channel)
		idempotencyManager := idempotency.NewManager(nil) // Needs a Redis-backed storage

		// --- STUB: Replace with actual RedisIdempotencyStorage when available ---
		// For now, simulate idempotency check by always returning true for simplicity
		// and directly creating the delivery.
		// In a real scenario, this would interact with Redis.
		_ = deliveryKey        // Avoid unused variable warning
		_ = idempotencyManager // Avoid unused variable warning
		// --- END STUB ---

		delivery := &domain.Delivery{
			ID:        uuid.New(),
			ReleaseID: release.ID,
			UserID:    sub.UserID,
			Channel:   sub.Channel,
			Status:    "pending",
			Attempt:   0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := p.deliveryStore.CreateDelivery(ctx, delivery); err != nil { // Исправлено
			logger.L().Sugar().Errorf("%s: failed to create delivery for release %s, user %s, channel %s: %v", op, release.ID, sub.UserID, sub.Channel, err)
		}
		logger.L().Sugar().Debugf("%s: enqueued delivery %s for release %s, user %s, channel %s", op, delivery.ID, release.ID, sub.UserID, sub.Channel)
	}

	logger.L().Sugar().Debugf("%s: finished enqueuing deliveries for release %s", op, release.ID)
	return nil
}
