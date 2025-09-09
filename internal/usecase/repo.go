package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mackb/releaseradar/internal/adapter/github"
	"github.com/mackb/releaseradar/internal/adapter/persistence"
	"github.com/mackb/releaseradar/internal/domain"
	"github.com/mackb/releaseradar/pkg/logger"
)

type repoUseCase struct {
	repoStore    persistence.RepoRepository
	userStore    persistence.UserRepository
	githubClient github.Client
	transactor   persistence.Transactor
}

func NewRepoUseCase(repoStore persistence.RepoRepository, userStore persistence.UserRepository, githubClient github.Client, transactor persistence.Transactor) RepoUseCase {
	return &repoUseCase{
		repoStore:    repoStore,
		userStore:    userStore,
		githubClient: githubClient,
		transactor:   transactor,
	}
}

func (r *repoUseCase) AddRepo(ctx context.Context, userID uuid.UUID, owner, name string) (*domain.Repo, error) {
	const op = "RepoUseCase.AddRepo"
	logger.L().Sugar().Debugf("%s: attempting to add repo %s/%s for user %s", op, owner, name, userID)

	var repo *domain.Repo
	err := r.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := r.userStore.GetUserByID(txCtx, userID)
		if err != nil {
			return fmt.Errorf("%s: user %s not found: %w", op, userID, err)
		}

		existingRepo, err := r.repoStore.GetRepoByOwnerAndName(txCtx, owner, name)
		if err != nil {
			return fmt.Errorf("%s: failed to check for existing repo: %w", op, err)
		}
		if existingRepo != nil {
			return fmt.Errorf("%s: repo %s/%s already exists", op, owner, name)
		}

		newRepo := &domain.Repo{
			ID:            uuid.New(),
			UserID:        userID,
			Owner:         owner,
			Name:          name,
			ETag:          "", // Initial ETag
			LastCheckedAt: time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Optionally, fetch initial ETag from GitHub here or let the poller handle it on first run

		if err := r.repoStore.CreateRepo(txCtx, newRepo); err != nil {
			return fmt.Errorf("%s: failed to create repo: %w", op, err)
		}
		repo = newRepo
		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.L().Sugar().Infof("%s: successfully added repo %s/%s for user %s", op, owner, name, userID)
	return repo, nil
}

func (r *repoUseCase) ListRepos(ctx context.Context, userID uuid.UUID) ([]domain.Repo, error) {
	const op = "RepoUseCase.ListRepos"
	logger.L().Sugar().Debugf("%s: attempting to list repos for user %s", op, userID)

	repos, err := r.repoStore.ListRepos(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to list repos for user %s: %w", op, userID, err)
	}

	return repos, nil
}

func (r *repoUseCase) GetRepoByID(ctx context.Context, repoID uuid.UUID) (*domain.Repo, error) {
	const op = "RepoUseCase.GetRepoByID"
	logger.L().Sugar().Debugf("%s: attempting to get repo with ID %s", op, repoID)

	repo, err := r.repoStore.GetRepoByID(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get repo by ID %s: %w", op, repoID, err)
	}

	return repo, nil
}
