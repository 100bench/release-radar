package persistence

import (
	"context"
	"fmt"
	"time"

	"errors"

	"github.com/google/uuid"
	"github.com/mackb/releaseradar/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresStore struct {
	db *gorm.DB
}

func NewPostgresStore(dsn string) (Store, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &PostgresStore{db: db}, nil

}

func (p *PostgresStore) WithinTransaction(ctx context.Context, txFunc func(ctx context.Context) error) error {
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return txFunc(context.WithValue(ctx, txKey, tx))
	})
}

type txKeyType struct{}

var txKey = txKeyType{}

func getDB(ctx context.Context, p *PostgresStore) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return p.db
}

// --- User Repository Implementations ---

func (p *PostgresStore) CreateUser(ctx context.Context, user *domain.User) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Create(user).Error
}

func (p *PostgresStore) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	db := getDB(ctx, p)
	var user domain.User
	if err := db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

func (p *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	db := getDB(ctx, p)
	var user domain.User
	if err := db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}

// --- Repo Repository Implementations ---

func (p *PostgresStore) CreateRepo(ctx context.Context, repo *domain.Repo) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Create(repo).Error
}

func (p *PostgresStore) GetRepoByID(ctx context.Context, id uuid.UUID) (*domain.Repo, error) {
	db := getDB(ctx, p)
	var repo domain.Repo
	if err := db.WithContext(ctx).First(&repo, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &repo, nil
}

func (p *PostgresStore) GetRepoByOwnerAndName(ctx context.Context, owner, name string) (*domain.Repo, error) {
	db := getDB(ctx, p)
	var repo domain.Repo
	if err := db.WithContext(ctx).Where("owner = ? AND name = ?", owner, name).First(&repo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &repo, nil
}

func (p *PostgresStore) UpdateRepo(ctx context.Context, repo *domain.Repo) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Save(repo).Error
}

func (p *PostgresStore) ListRepos(ctx context.Context, userID uuid.UUID) ([]domain.Repo, error) {
	db := getDB(ctx, p)
	var repos []domain.Repo
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&repos).Error; err != nil {
		return nil, err
	}
	return repos, nil
}

// --- Subscription Repository Implementations ---

func (p *PostgresStore) CreateSubscription(ctx context.Context, sub *domain.Subscription) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Create(sub).Error
}

func (p *PostgresStore) GetSubscription(ctx context.Context, repoID, userID uuid.UUID, channel string) (*domain.Subscription, error) {
	db := getDB(ctx, p)
	var sub domain.Subscription
	if err := db.WithContext(ctx).Where("repo_id = ? AND user_id = ? AND channel = ?", repoID, userID, channel).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (p *PostgresStore) DeleteSubscription(ctx context.Context, repoID, userID uuid.UUID, channel string) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Where("repo_id = ? AND user_id = ? AND channel = ?", repoID, userID, channel).Delete(&domain.Subscription{}).Error
}

func (p *PostgresStore) ListSubscriptionsByRepoID(ctx context.Context, repoID uuid.UUID) ([]domain.Subscription, error) {
	db := getDB(ctx, p)
	var subs []domain.Subscription
	if err := db.WithContext(ctx).Where("repo_id = ?", repoID).Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (p *PostgresStore) ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	db := getDB(ctx, p)
	var subs []domain.Subscription
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

// --- Release Repository Implementations ---

func (p *PostgresStore) CreateRelease(ctx context.Context, release *domain.Release) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Create(release).Error
}

func (p *PostgresStore) GetReleaseByID(ctx context.Context, id uuid.UUID) (*domain.Release, error) {
	db := getDB(ctx, p)
	var release domain.Release
	if err := db.WithContext(ctx).First(&release, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &release, nil
}

func (p *PostgresStore) GetReleaseByRepoIDAndTag(ctx context.Context, repoID uuid.UUID, tag string) (*domain.Release, error) {
	db := getDB(ctx, p)
	var release domain.Release
	if err := db.WithContext(ctx).Where("repo_id = ? AND tag = ?", repoID, tag).First(&release).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &release, nil
}

func (p *PostgresStore) ListReleasesByRepoID(ctx context.Context, repoID uuid.UUID) ([]domain.Release, error) {
	db := getDB(ctx, p)
	var releases []domain.Release
	if err := db.WithContext(ctx).Where("repo_id = ?", repoID).Find(&releases).Error; err != nil {
		return nil, err
	}
	return releases, nil
}

// --- Delivery Repository Implementations ---

func (p *PostgresStore) CreateDelivery(ctx context.Context, delivery *domain.Delivery) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Create(delivery).Error
}

func (p *PostgresStore) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, status, lastError string, attempt int) error {
	db := getDB(ctx, p)
	return db.WithContext(ctx).Model(&domain.Delivery{}).Where("id = ?", id).Updates(map[string]interface{}{"status": status, "last_error": lastError, "attempt": attempt, "updated_at": time.Now()}).Error
}

func (p *PostgresStore) ListPendingDeliveries(ctx context.Context) ([]domain.Delivery, error) {
	db := getDB(ctx, p)
	var deliveries []domain.Delivery
	if err := db.WithContext(ctx).Where("status = ?", "pending").Find(&deliveries).Error; err != nil {
		return nil, err
	}
	return deliveries, nil
}

func (p *PostgresStore) GetDelivery(ctx context.Context, releaseID, userID uuid.UUID, channel string) (*domain.Delivery, error) {
	db := getDB(ctx, p)
	var delivery domain.Delivery
	if err := db.WithContext(ctx).Where("release_id = ? AND user_id = ? AND channel = ?", releaseID, userID, channel).First(&delivery).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &delivery, nil
}
