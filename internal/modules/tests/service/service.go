package service

import (
	"context"
	"errors"
	"strings"

	"ai_testing/internal/modules/tests/dto"
	"ai_testing/internal/modules/tests/model"
	"ai_testing/internal/modules/tests/repository"
	usersmodel "ai_testing/internal/modules/users/model"
	"ai_testing/internal/storage"

	"github.com/google/uuid"
)

type Service struct {
	repo       *repository.Repository
	audioStore *storage.AudioStore
}

func New(repo *repository.Repository, audioStore *storage.AudioStore) *Service {
	return &Service{repo: repo, audioStore: audioStore}
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
	for _, section := range sections {
		byTestID[section.TestID] = append(byTestID[section.TestID], section)
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
	for _, mapping := range mappings {
		byTestID[mapping.TestID] = mapping
	}

	resp := make([]dto.TestWithUserStatusResponse, 0, len(tests))
	for _, t := range tests {
		if mapping, ok := byTestID[t.ID]; ok {
			m := mapping
			resp = append(resp, dto.ToTestWithUserStatusResponse(t, &m))
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

	userTestMapping := usersmodel.UserTestMapping{
		UserID:               userID,
		TestID:               testID,
		MicroPhonePermission: microPhonePermission,
	}

	if err := s.repo.CreateUserTestMapping(ctx, &userTestMapping); err != nil {
		return nil, err
	}

	return &userTestMapping, nil
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

	if mapping.Status == "initialized" {
		if err := s.repo.MarkUserTestInProgress(ctx, mapping.ID); err != nil {
			return nil, err
		}
	}

	section, err := s.repo.GetSectionByID(ctx, payload.SectionID)
	if err != nil {
		return nil, err
	}

	if section.TestID != mapping.TestID {
		return nil, errors.New("section does not belong to the test")
	}

	testNotes := payload.TestNotes
	if testNotes == nil {
		testNotes = []string{}
	}

	userQuestionMapping := usersmodel.UserQuestionMapping{
		UserTestMappingID:    payload.UserTestMappingID,
		TestSectionMappingID: payload.SectionID,
		Question:             payload.Questions,
		UserAnswer:           payload.Answers,
		TestNotes:            testNotes,
		MarksObtained:        0,
		AIFeedback:           "",
		ChangedWindowsCount:  payload.ChangedWindowsCount,
		HasGraded:            false,
	}

	if err := s.repo.UpsertUserQuestionMapping(ctx, &userQuestionMapping); err != nil {
		return nil, err
	}

	return &userQuestionMapping, nil
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

	userQuestionMapping := usersmodel.UserQuestionMapping{
		UserTestMappingID:    userTestMappingID,
		TestSectionMappingID: sectionID,
		Question:             []string{question},
		UserAnswer:           []string{audioPath},
		TestNotes:            []string{},
		MarksObtained:        0,
		AIFeedback:           "",
		ChangedWindowsCount:  changedWindowsCount,
		HasGraded:            false,
	}

	if err := s.repo.UpsertUserQuestionMapping(ctx, &userQuestionMapping); err != nil {
		return nil, err
	}

	if mapping.Status != "submitted" && mapping.Status != "graded" {
		if err := s.repo.MarkUserTestSubmitted(ctx, mapping.ID); err != nil {
			return nil, err
		}
	}

	return &userQuestionMapping, nil
}

func (s *Service) DropUserTest(
	ctx context.Context,
	userID uuid.UUID,
	userTestMappingID uuid.UUID,
) error {
	mapping, err := s.repo.GetUserTestMappingByIDAndUserID(ctx, userTestMappingID, userID)
	if err != nil {
		return err
	}
	return s.repo.MarkUserTestDropped(ctx, mapping.ID)
}

func (s *Service) GetRandomSpeakingTopic(
	ctx context.Context,
	sectionID uuid.UUID,
) (*model.SpeakingTopic, error) {
	return s.repo.GetRandomSpeakingTopic(ctx, sectionID)
}

func (s *Service) GetRandomWritingTopic(
	ctx context.Context,
	sectionID uuid.UUID,
) (*model.WritingTopic, error) {
	return s.repo.GetRandomWritingTopic(ctx, sectionID)
}

func (s *Service) GetRandomReadingComprehension(
	ctx context.Context,
	sectionID uuid.UUID,
) (*model.ReadingComprehension, error) {
	return s.repo.GetRandomReadingComprehension(ctx, sectionID)
}

func (s *Service) GetUserTestResults(
	ctx context.Context,
	userID uuid.UUID,
	userTestMappingID uuid.UUID,
) (*dto.UserTestResultsResponse, error) {
	mapping, err := s.repo.GetUserTestMappingByIDAndUserID(ctx, userTestMappingID, userID)
	if err != nil {
		return nil, err
	}

	test, err := s.repo.GetByID(ctx, mapping.TestID)
	if err != nil {
		return nil, err
	}

	sections, err := s.repo.ListSectionsByTestID(ctx, mapping.TestID)
	if err != nil {
		return nil, err
	}

	answers, err := s.repo.ListUserQuestionMappingsByUserTestMappingID(ctx, mapping.ID)
	if err != nil {
		return nil, err
	}

	answersBySection := make(map[uuid.UUID]usersmodel.UserQuestionMapping, len(answers))
	for _, answerRow := range answers {
		answersBySection[answerRow.TestSectionMappingID] = answerRow
	}

	resp := dto.ToUserTestResultsResponse(*test, *mapping, sections, answersBySection)
	return &resp, nil
}

func (s *Service) GetUserAudio(
	ctx context.Context,
	userID uuid.UUID,
	userTestMappingID uuid.UUID,
	sectionID uuid.UUID,
) (*storage.ResolvedAudio, error) {
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

	row, err := s.repo.GetUserQuestionMappingByUserTestMappingIDAndSectionID(ctx, mapping.ID, sectionID)
	if err != nil {
		return nil, err
	}

	if len(row.UserAnswer) == 0 {
		return nil, errors.New("audio not found")
	}

	audioPath := strings.TrimSpace(row.UserAnswer[0])
	return s.audioStore.Resolve(audioPath, userID, userTestMappingID, sectionID)
}
