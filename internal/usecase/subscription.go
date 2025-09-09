package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mackb/releaseradar/internal/adapter/persistence"
	"github.com/mackb/releaseradar/internal/domain"
	"github.com/mackb/releaseradar/pkg/logger"
)

type subscriptionUseCase struct {
	subscriptionStore persistence.SubscriptionRepository
	userStore         persistence.UserRepository
	repoStore         persistence.RepoRepository
	transactor        persistence.Transactor
}

func NewSubscriptionUseCase(subscriptionStore persistence.SubscriptionRepository, userStore persistence.UserRepository, repoStore persistence.RepoRepository, transactor persistence.Transactor) SubscriptionUseCase {
	return &subscriptionUseCase{
		subscriptionStore: subscriptionStore,
		userStore:         userStore,
		repoStore:         repoStore,
		transactor:        transactor,
	}
}

func (s *subscriptionUseCase) Subscribe(ctx context.Context, userID, repoID uuid.UUID, channel string) (*domain.Subscription, error) {
	const op = "SubscriptionUseCase.Subscribe"
	logger.L().Sugar().Debugf("%s: attempting to subscribe user %s to repo %s on channel %s", op, userID, repoID, channel)

	var subscription *domain.Subscription
	err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		_, err := s.userStore.GetUserByID(txCtx, userID)
		if err != nil {
			return fmt.Errorf("%s: user %s not found: %w", op, userID, err)
		}

		_, err = s.repoStore.GetRepoByID(txCtx, repoID)
		if err != nil {
			return fmt.Errorf("%s: repo %s not found: %w", op, repoID, err)
		}

		existingSub, err := s.subscriptionStore.GetSubscription(txCtx, repoID, userID, channel)
		if err != nil && err != persistence.ErrNotFound {
			return fmt.Errorf("%s: failed to check for existing subscription: %w", op, err)
		}
		if existingSub != nil {
			return fmt.Errorf("%s: user %s already subscribed to repo %s on channel %s", op, userID, repoID, channel)
		}

		newSub := &domain.Subscription{
			ID:        uuid.New(),
			RepoID:    repoID,
			UserID:    userID,
			Channel:   channel,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.subscriptionStore.CreateSubscription(txCtx, newSub); err != nil {
			return fmt.Errorf("%s: failed to create subscription: %w", op, err)
		}
		subscription = newSub
		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.L().Sugar().Infof("%s: successfully subscribed user %s to repo %s on channel %s", op, userID, repoID, channel)
	return subscription, nil
}

func (s *subscriptionUseCase) Unsubscribe(ctx context.Context, userID, repoID uuid.UUID, channel string) error {
	const op = "SubscriptionUseCase.Unsubscribe"
	logger.L().Sugar().Debugf("%s: attempting to unsubscribe user %s from repo %s on channel %s", op, userID, repoID, channel)

	err := s.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		existingSub, err := s.subscriptionStore.GetSubscription(txCtx, repoID, userID, channel)
		if err != nil {
			return fmt.Errorf("%s: failed to check for existing subscription: %w", op, err)
		}
		if existingSub == nil {
			return fmt.Errorf("%s: subscription not found for user %s, repo %s, channel %s", op, userID, repoID, channel)
		}

		if err := s.subscriptionStore.DeleteSubscription(txCtx, repoID, userID, channel); err != nil {
			return fmt.Errorf("%s: failed to delete subscription: %w", op, err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	logger.L().Sugar().Infof("%s: successfully unsubscribed user %s from repo %s on channel %s", op, userID, repoID, channel)
	return nil
}

func (s *subscriptionUseCase) ListSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	const op = "SubscriptionUseCase.ListSubscriptions"
	logger.L().Sugar().Debugf("%s: attempting to list subscriptions for user %s", op, userID)

	subs, err := s.subscriptionStore.ListSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to list subscriptions for user %s: %w", op, userID, err)
	}

	return subs, nil
}
