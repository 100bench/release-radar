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

type userUseCase struct {
	repo  persistence.UserRepository
	store persistence.Transactor
}

func NewUserUseCase(repo persistence.UserRepository, store persistence.Transactor) UserUseCase {
	return &userUseCase{repo: repo, store: store}
}

func (u *userUseCase) SignUp(ctx context.Context, email string) (*domain.User, error) {
	const op = "UserUseCase.SignUp"
	logger.L().Sugar().Debugf("%s: attempting to sign up user with email %s", op, email)

	var user *domain.User
	err := u.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		existingUser, err := u.repo.GetUserByEmail(txCtx, email)
		if err != nil {
			return fmt.Errorf("%s: failed to get user by email: %w", op, err)
		}
		if existingUser != nil {
			return fmt.Errorf("%s: user with email %s already exists", op, email)
		}

		newUser := &domain.User{
			ID:        uuid.New(),
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := u.repo.CreateUser(txCtx, newUser); err != nil {
			return fmt.Errorf("%s: failed to create user: %w", op, err)
		}
		user = newUser
		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.L().Sugar().Infof("%s: successfully signed up user %s", op, user.ID)
	return user, nil
}

func (u *userUseCase) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const op = "UserUseCase.GetUserByID"
	logger.L().Sugar().Debugf("%s: attempting to get user with ID %s", op, id)

	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get user by ID: %w", op, err)
	}

	return user, nil
}
