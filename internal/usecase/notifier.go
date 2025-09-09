package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/mackb/releaseradar/internal/adapter/persistence"
	"github.com/mackb/releaseradar/internal/adapter/telegram"
	"github.com/mackb/releaseradar/pkg/idempotency"
	"github.com/mackb/releaseradar/pkg/logger"
)

type notifierUseCase struct {
	deliveryStore      persistence.DeliveryRepository
	releaseStore       persistence.ReleaseRepository
	userStore          persistence.UserRepository
	telegramClient     telegram.Client
	idempotencyManager *idempotency.Manager
	transactor         persistence.Transactor
}

func NewNotifierUseCase(deliveryStore persistence.DeliveryRepository, releaseStore persistence.ReleaseRepository, userStore persistence.UserRepository, telegramClient telegram.Client, idempotencyManager *idempotency.Manager, transactor persistence.Transactor) NotifierUseCase {
	return &notifierUseCase{
		deliveryStore:      deliveryStore,
		releaseStore:       releaseStore,
		userStore:          userStore,
		telegramClient:     telegramClient,
		idempotencyManager: idempotencyManager,
		transactor:         transactor,
	}
}

func (n *notifierUseCase) Notify(ctx context.Context) error {
	const op = "NotifierUseCase.Notify"
	logger.L().Sugar().Debugf("%s: starting notification cycle", op)

	deliveries, err := n.deliveryStore.ListPendingDeliveries(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to list pending deliveries: %w", op, err)
	}

	if len(deliveries) == 0 {
		logger.L().Sugar().Debugf("%s: no pending deliveries", op)
		return nil
	}

	logger.L().Sugar().Infof("%s: found %d pending deliveries", op, len(deliveries))

	for _, delivery := range deliveries {
		// Use idempotency manager to ensure delivery is processed only once
		idempotencyKey := fmt.Sprintf("notify:%s:%s:%s", delivery.ReleaseID, delivery.UserID, delivery.Channel)

		err := n.idempotencyManager.Do(ctx, idempotencyKey, 10*time.Minute, func() error {
			// Fetch associated release and user details
			release, err := n.releaseStore.GetReleaseByID(ctx, delivery.ReleaseID)
			if err != nil {
				return fmt.Errorf("%s: failed to get release %s for delivery %s: %w", op, delivery.ReleaseID, delivery.ID, err)
			}
			if release == nil {
				logger.L().Sugar().Warnf("%s: release %s not found for delivery %s, skipping", op, delivery.ReleaseID, delivery.ID)
				return n.deliveryStore.UpdateDeliveryStatus(ctx, delivery.ID, "skipped", "release not found", delivery.Attempt+1) // Update status to skipped
			}

			user, err := n.userStore.GetUserByID(ctx, delivery.UserID)
			if err != nil {
				return fmt.Errorf("%s: failed to get user %s for delivery %s: %w", op, delivery.UserID, delivery.ID, err)
			}
			if user == nil {
				logger.L().Sugar().Warnf("%s: user %s not found for delivery %s, skipping", op, delivery.UserID, delivery.ID)
				return n.deliveryStore.UpdateDeliveryStatus(ctx, delivery.ID, "skipped", "user not found", delivery.Attempt+1) // Update status to skipped
			}

			message := fmt.Sprintf("New release for %s/%s: <b>%s</b> (%s)\n%s", release.RepoID, "", release.Title, release.Tag, release.URL) // Placeholder for repo name
			// In a real scenario, you'd get repo details from release.RepoID to display owner/name

			logger.L().Sugar().Infof("%s: sending message for release %s to user %s on channel %s", op, release.ID, user.ID, delivery.Channel)
			sendErr := n.telegramClient.SendMessage(ctx, delivery.Channel, message)
			if sendErr != nil {
				logger.L().Sugar().Errorf("%s: failed to send telegram message for delivery %s: %v", op, delivery.ID, sendErr)
				// Mark as failed and retry later
				return n.deliveryStore.UpdateDeliveryStatus(ctx, delivery.ID, "failed", sendErr.Error(), delivery.Attempt+1)
			}

			logger.L().Sugar().Infof("%s: successfully sent telegram message for delivery %s", op, delivery.ID)
			return n.deliveryStore.UpdateDeliveryStatus(ctx, delivery.ID, "sent", "", delivery.Attempt+1)
		})

		if err != nil {
			logger.L().Sugar().Errorf("%s: failed to process delivery %s: %v", op, delivery.ID, err)
		}
	}

	logger.L().Sugar().Debugf("%s: finished notification cycle", op)
	return nil
}
