package repository

import (
	"ai_testing/internal/modules/tests/model"
	usersmodel "ai_testing/internal/modules/users/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Repository struct {
	db *bun.DB
}

func New(db *bun.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]model.Test, error) {
	var tests []model.Test
	if err := r.db.NewSelect().
		Model(&tests).
		Order("name ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return tests, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.Test, error) {
	var test model.Test
	if err := r.db.NewSelect().
		Model(&test).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return &test, nil
}

func (r *Repository) ListSectionsByTestIDs(ctx context.Context, testIDs []uuid.UUID) ([]model.TestSectionMapping, error) {
	var sections []model.TestSectionMapping
	if len(testIDs) == 0 {
		return sections, nil
	}

	if err := r.db.NewSelect().
		Model(&sections).
		Where("test_id IN (?)", bun.In(testIDs)).
		Order("name ASC").
		Scan(ctx); err != nil {
		return nil, err
	}

	return sections, nil
}

func (r *Repository) ListSectionsByTestID(ctx context.Context, testID uuid.UUID) ([]model.TestSectionMapping, error) {
	sections, err := r.ListSectionsByTestIDs(ctx, []uuid.UUID{testID})
	if err != nil {
		return nil, err
	}
	return sections, nil
}

func (r *Repository) GetSectionByID(ctx context.Context, id uuid.UUID) (*model.TestSectionMapping, error) {
	var section model.TestSectionMapping
	if err := r.db.NewSelect().
		Model(&section).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return &section, nil
}

func (r *Repository) ListUserTestMappingsByUserID(ctx context.Context, userID uuid.UUID) ([]usersmodel.UserTestMapping, error) {
	var mappings []usersmodel.UserTestMapping
	if err := r.db.NewSelect().
		Model(&mappings).
		Where("user_id = ?", userID).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []usersmodel.UserTestMapping{}, nil
		}
		return nil, err
	}
	return mappings, nil
}

func (r *Repository) CreateUserTestMapping(ctx context.Context, mapping *usersmodel.UserTestMapping) error {
	_, err := r.db.NewInsert().
		Model(mapping).
		Column("user_id", "test_id", "micro_phone_permission").
		Returning("id").
		Exec(ctx, &mapping.ID)
	return err
}

func (r *Repository) GetUserTestMappingByUserIDAndTestID(
	ctx context.Context,
	userID uuid.UUID,
	testID uuid.UUID,
) (*usersmodel.UserTestMapping, error) {
	var mapping usersmodel.UserTestMapping
	if err := r.db.NewSelect().
		Model(&mapping).
		Where("user_id = ?", userID).
		Where("test_id = ?", testID).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *Repository) GetUserTestMappingByIDAndUserID(
	ctx context.Context,
	id uuid.UUID,
	userID uuid.UUID,
) (*usersmodel.UserTestMapping, error) {
	var mapping usersmodel.UserTestMapping
	if err := r.db.NewSelect().
		Model(&mapping).
		Where("id = ?", id).
		Where("user_id = ?", userID).
		Limit(1).
		Scan(ctx); err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *Repository) UpsertUserQuestionMapping(ctx context.Context, m *usersmodel.UserQuestionMapping) error {
	var existing usersmodel.UserQuestionMapping
	if err := r.db.NewSelect().
		Model(&existing).
		Where("user_test_mapping_id = ?", m.UserTestMappingID).
		Where("test_section_mapping_id = ?", m.TestSectionMappingID).
		Limit(1).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err := r.db.NewInsert().
				Model(m).
				Column("user_test_mapping_id", "test_section_mapping_id", "question", "user_answer", "changed_windows_count").
				Returning("id").
				Exec(ctx, &m.ID)
			return err
		}
		return err
	}

	m.ID = existing.ID
	_, err := r.db.NewUpdate().
		Model(m).
		Column("question", "user_answer", "changed_windows_count", "marks_obtained", "aif_feedback", "has_graded").
		WherePK().
		Exec(ctx)
	return err
}
