package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/mackb/releaseradar/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type RepoRepository interface {
	CreateRepo(ctx context.Context, repo *domain.Repo) error
	GetRepoByID(ctx context.Context, id uuid.UUID) (*domain.Repo, error)
	GetRepoByOwnerAndName(ctx context.Context, owner, name string) (*domain.Repo, error)
	UpdateRepo(ctx context.Context, repo *domain.Repo) error
	ListRepos(ctx context.Context, userID uuid.UUID) ([]domain.Repo, error)
}

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, sub *domain.Subscription) error
	GetSubscription(ctx context.Context, repoID, userID uuid.UUID, channel string) (*domain.Subscription, error)
	DeleteSubscription(ctx context.Context, repoID, userID uuid.UUID, channel string) error
	ListSubscriptionsByRepoID(ctx context.Context, repoID uuid.UUID) ([]domain.Subscription, error)
	ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error)
}

type ReleaseRepository interface {
	CreateRelease(ctx context.Context, release *domain.Release) error
	GetReleaseByID(ctx context.Context, id uuid.UUID) (*domain.Release, error)
	GetReleaseByRepoIDAndTag(ctx context.Context, repoID uuid.UUID, tag string) (*domain.Release, error)
	ListReleasesByRepoID(ctx context.Context, repoID uuid.UUID) ([]domain.Release, error)
}

type DeliveryRepository interface {
	CreateDelivery(ctx context.Context, delivery *domain.Delivery) error
	UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status, lastError string, attempt int) error
	ListPendingDeliveries(ctx context.Context) ([]domain.Delivery, error)
	GetDelivery(ctx context.Context, releaseID, userID uuid.UUID, channel string) (*domain.Delivery, error)
}

type Transactor interface {
	WithinTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error
}

// Store combines all repository interfaces and the transactor.
type Store interface {
	UserRepository
	RepoRepository
	SubscriptionRepository
	ReleaseRepository
	DeliveryRepository
	Transactor
}
