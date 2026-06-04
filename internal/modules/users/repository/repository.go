package repository

import (
	"ai_testing/internal/modules/users/model"
	"context"

	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func New(db *bun.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.NewSelect().Model(&user).Where("email = ?", email).Scan(ctx); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Register(ctx context.Context, user *model.User) (*model.User, error) {
	if _, err := r.db.NewInsert().
		Model(user).
		Returning("id").
		Exec(ctx, &user.ID); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) RegisterUserProfile(ctx context.Context, userProfile *model.UserProfile) error {
	if _, err := r.db.NewInsert().Model(userProfile).Exec(ctx); err != nil {
		return err
	}
	return nil
}
