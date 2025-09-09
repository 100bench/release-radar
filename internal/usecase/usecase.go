package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/mackb/releaseradar/internal/domain"
)

type UserUseCase interface {
	SignUp(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type RepoUseCase interface {
	AddRepo(ctx context.Context, userID uuid.UUID, owner, name string) (*domain.Repo, error)
	ListRepos(ctx context.Context, userID uuid.UUID) ([]domain.Repo, error)
	GetRepoByID(ctx context.Context, repoID uuid.UUID) (*domain.Repo, error)
}

type SubscriptionUseCase interface {
	Subscribe(ctx context.Context, userID, repoID uuid.UUID, channel string) (*domain.Subscription, error)
	Unsubscribe(ctx context.Context, userID, repoID uuid.UUID, channel string) error
	ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error)
}

type PollerUseCase interface {
	PollReleases(ctx context.Context) error
	EnqueueDeliveries(ctx context.Context, release *domain.Release) error
}

type NotifierUseCase interface {
	Notify(ctx context.Context) error
}

// Usecases combines all use case interfaces
type Usecases struct { // Удалены неиспользуемые поля
	User         UserUseCase
	Repo         RepoUseCase
	Subscription SubscriptionUseCase
	Poller       PollerUseCase
	Notifier     NotifierUseCase
}
