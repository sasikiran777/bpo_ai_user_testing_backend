package service

import (
	"context"
	"errors"

	"ai_testing/internal/modules/tests/dto"
	"ai_testing/internal/modules/tests/model"
	"ai_testing/internal/modules/tests/repository"
	usersmodel "ai_testing/internal/modules/users/model"

	"github.com/google/uuid"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]model.Test, error) {
	tests, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	testIDs := make([]uuid.UUID, 0, len(tests))
	for _, t := range tests {
		testIDs = append(testIDs, t.ID)
	}

	sections, err := s.repo.ListSectionsByTestIDs(ctx, testIDs)
	if err != nil {
		return nil, err
	}

	byTestID := make(map[uuid.UUID][]model.TestSectionMapping, len(tests))
	for _, s := range sections {
		byTestID[s.TestID] = append(byTestID[s.TestID], s)
	}

	for i := range tests {
		tests[i].Sections = byTestID[tests[i].ID]
	}

	return tests, nil
}

func (s *Service) ListForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]dto.TestWithUserStatusResponse, error) {
	tests, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	mappings, err := s.repo.ListUserTestMappingsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	byTestID := make(map[uuid.UUID]usersmodel.UserTestMapping, len(mappings))
	for _, m := range mappings {
		byTestID[m.TestID] = m
	}

	resp := make([]dto.TestWithUserStatusResponse, 0, len(tests))
	for _, t := range tests {
		if m, ok := byTestID[t.ID]; ok {
			mm := m
			resp = append(resp, dto.ToTestWithUserStatusResponse(t, &mm))
			continue
		}
		resp = append(resp, dto.ToTestWithUserStatusResponse(t, nil))
	}

	return resp, nil
}

func (s *Service) GetSectionsByTestID(
	ctx context.Context,
	testID uuid.UUID,
) (*model.Test, error) {
	test, err := s.repo.GetByID(ctx, testID)
	if err != nil {
		return nil, err
	}

	sections, err := s.repo.ListSectionsByTestID(ctx, testID)
	if err != nil {
		return nil, err
	}

	test.Sections = sections
	return test, nil
}

func (s *Service) CreateUserTestMapping(
	ctx context.Context,
	userID uuid.UUID,
	testID uuid.UUID,
	microPhonePermission bool,
) (*usersmodel.UserTestMapping, error) {
	if _, err := s.repo.GetByID(ctx, testID); err != nil {
		return nil, err
	}

	mapping := usersmodel.UserTestMapping{
		UserID:               userID,
		TestID:               testID,
		MicroPhonePermission: microPhonePermission,
	}

	if err := s.repo.CreateUserTestMapping(ctx, &mapping); err != nil {
		return nil, err
	}

	return &mapping, nil
}

func (s *Service) GetUserTestStatus(
	ctx context.Context,
	userID uuid.UUID,
	testID uuid.UUID,
) (*usersmodel.UserTestMapping, error) {
	if _, err := s.repo.GetByID(ctx, testID); err != nil {
		return nil, err
	}
	return s.repo.GetUserTestMappingByUserIDAndTestID(ctx, userID, testID)
}

func (s *Service) SaveAnswers(
	ctx context.Context,
	userID uuid.UUID,
	payload dto.SaveAnswersRequest,
) (*usersmodel.UserQuestionMapping, error) {
	mapping, err := s.repo.GetUserTestMappingByIDAndUserID(ctx, payload.UserTestMappingID, userID)
	if err != nil {
		return nil, err
	}

	section, err := s.repo.GetSectionByID(ctx, payload.SectionID)
	if err != nil {
		return nil, err
	}

	if section.TestID != mapping.TestID {
		return nil, errors.New("section does not belong to the test")
	}

	m := usersmodel.UserQuestionMapping{
		UserTestMappingID:    payload.UserTestMappingID,
		TestSectionMappingID: payload.SectionID,
		Question:             payload.Questions,
		UserAnswer:           payload.Answers,
		MarksObtained:        0,
		AIFeedback:           "",
		ChangedWindowsCount:  payload.ChangedWindowsCount,
		HasGraded:            false,
	}

	if err := s.repo.UpsertUserQuestionMapping(ctx, &m); err != nil {
		return nil, err
	}

	return &m, nil
}

func (s *Service) SaveAudioAnswer(
	ctx context.Context,
	userID uuid.UUID,
	userTestMappingID uuid.UUID,
	sectionID uuid.UUID,
	question string,
	changedWindowsCount int,
	audioPath string,
) (*usersmodel.UserQuestionMapping, error) {
	mapping, err := s.repo.GetUserTestMappingByIDAndUserID(ctx, userTestMappingID, userID)
	if err != nil {
		return nil, err
	}

	section, err := s.repo.GetSectionByID(ctx, sectionID)
	if err != nil {
		return nil, err
	}

	if section.TestID != mapping.TestID {
		return nil, errors.New("section does not belong to the test")
	}

	m := usersmodel.UserQuestionMapping{
		UserTestMappingID:    userTestMappingID,
		TestSectionMappingID: sectionID,
		Question:             []string{question},
		UserAnswer:           []string{audioPath},
		MarksObtained:        0,
		AIFeedback:           "",
		ChangedWindowsCount:  changedWindowsCount,
		HasGraded:            false,
	}

	if err := s.repo.UpsertUserQuestionMapping(ctx, &m); err != nil {
		return nil, err
	}

	return &m, nil
}
