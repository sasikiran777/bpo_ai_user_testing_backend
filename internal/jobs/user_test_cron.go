package jobs

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"ai_testing/internal/ai"
	"ai_testing/internal/ai/stt"
	"ai_testing/internal/config"
	testsmodel "ai_testing/internal/modules/tests/model"
	usersmodel "ai_testing/internal/modules/users/model"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	userTestStatusInitialized = "initialized"
	userTestStatusSubmitted   = "submitted"
	userTestStatusInGrading   = "in_grading"
	userTestStatusDropped     = "dropped"
	userTestStatusGraded      = "graded"
)

func StartUserTestMappingCron(ctx context.Context, logger *slog.Logger, db *bun.DB, cfg *config.Config) {
	if cfg == nil {
		logger.Error("cron_user_test_mapping_config_nil")
		return
	}
	if !cfg.UserTestCronEnabled {
		logger.Info("cron_user_test_mapping_disabled")
		return
	}

	llm, err := ai.NewChutesLLM(cfg)
	if err != nil {
		logger.Error("ai_llm_init_error", "error", err.Error())
		return
	}
	grader, err := ai.NewGrader(llm)
	if err != nil {
		logger.Error("ai_grader_init_error", "error", err.Error())
		return
	}

	transcriber, err := stt.NewWhisperCLI(cfg.WhisperBin, cfg.WhisperModelPath, cfg.WhisperLanguage)
	if err != nil {
		logger.Error("whisper_init_error", "error", err.Error())
		return
	}

	intervalSeconds := cfg.UserTestCronSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 60
	}

	logger.Info("cron_user_test_mapping_start", "interval_seconds", intervalSeconds)

	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	go func() {
		defer ticker.Stop()
		defer logger.Info("cron_user_test_mapping_stop")

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCronTick(ctx, logger, db, grader, transcriber, cfg.AILogRaw)
			}
		}
	}()
}

func runCronTick(
	ctx context.Context,
	logger *slog.Logger,
	db *bun.DB,
	grader *ai.Grader,
	transcriber *stt.WhisperCLI,
	aiLogRaw bool,
) {
	start := time.Now()
	logger.Info("cron_user_test_mapping_tick_start")
	initializedScanned, droppedCount := dropExpiredInitialized(ctx, logger, db)
	submittedScanned, gradedMappings, gradedRows := gradeSubmitted(ctx, logger, db, grader, transcriber, aiLogRaw)
	logger.Info(
		"cron_user_test_mapping_tick_finish",
		"initialized_scanned", initializedScanned,
		"dropped_count", droppedCount,
		"submitted_scanned", submittedScanned,
		"graded_mappings", gradedMappings,
		"graded_rows", gradedRows,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

func dropExpiredInitialized(ctx context.Context, logger *slog.Logger, db *bun.DB) (int, int) {
	rows, err := listInitializedMappings(ctx, db, 200)
	if err != nil {
		logger.Error("cron_initialized_scan_error", "error", err.Error())
		return 0, 0
	}

	now := time.Now()
	droppedCount := 0
	for _, r := range rows {
		sections, err := listActiveSectionsByTestID(ctx, db, r.TestID)
		if err != nil {
			logger.Error("cron_sections_scan_error", "test_id", r.TestID.String(), "error", err.Error())
			continue
		}

		totalMaxTime := totalMaxTimeMinutes(sections)
		if totalMaxTime <= 0 {
			continue
		}
		if now.Sub(r.CreatedAt) <= time.Duration(totalMaxTime)*time.Minute {
			continue
		}

		if err := updateUserTestMappingStatus(ctx, db, r.ID, userTestStatusDropped, &now); err != nil {
			logger.Error("cron_drop_update_error", "user_test_mapping_id", r.ID.String(), "error", err.Error())
			continue
		}
		droppedCount++
	}
	return len(rows), droppedCount
}

func gradeSubmitted(
	ctx context.Context,
	logger *slog.Logger,
	db *bun.DB,
	grader *ai.Grader,
	transcriber *stt.WhisperCLI,
	aiLogRaw bool,
) (int, int, int) {
	rows, err := listMappingsForGrading(ctx, db, 25)
	if err != nil {
		logger.Error("cron_submitted_scan_error", "error", err.Error())
		return 0, 0, 0
	}

	gradedMappings := 0
	gradedRows := 0
	for _, r := range rows {
		n, didGrade := gradeOneMapping(ctx, logger, db, grader, transcriber, r, aiLogRaw)
		if didGrade {
			gradedMappings++
		}
		gradedRows += n
	}
	return len(rows), gradedMappings, gradedRows
}

func gradeOneMapping(
	ctx context.Context,
	logger *slog.Logger,
	db *bun.DB,
	grader *ai.Grader,
	transcriber *stt.WhisperCLI,
	mapping usersmodel.UserTestMapping,
	aiLogRaw bool,
) (int, bool) {
	sections, err := listActiveSectionsByTestID(ctx, db, mapping.TestID)
	if err != nil {
		logger.Error("cron_sections_scan_error", "test_id", mapping.TestID.String(), "error", err.Error())
		return 0, false
	}

	answers, err := listUserQuestionMappingsByUserTestMappingID(ctx, db, mapping.ID)
	if err != nil {
		logger.Error("cron_answers_scan_error", "user_test_mapping_id", mapping.ID.String(), "error", err.Error())
		return 0, false
	}

	if len(answers) == 0 {
		return 0, false
	}

	answersBySection, pendingCount := buildAnswersIndex(answers)

	maxMarksBySection := buildMaxMarksIndex(sections)

	attempt := ai.Attempt{
		UserTestMappingID: mapping.ID,
		TestID:            mapping.TestID,
		Sections:          make([]ai.SectionAttempt, 0, len(sections)),
	}
	speakingNeedsTranscript := false
	speakingNeedsTranscript, pendingCount = appendAttemptSections(
		ctx,
		logger,
		db,
		transcriber,
		mapping.ID,
		sections,
		answersBySection,
		&attempt,
		pendingCount,
	)

	if pendingCount == 0 && !speakingNeedsTranscript {
		markMappingAsGraded(ctx, logger, db, mapping)
		return 0, false
	}

	aiStart := time.Now()
	markMappingAsInGrading(ctx, logger, db, mapping)
	logger.Info("cron_ai_call_start", "user_test_mapping_id", mapping.ID.String())
	actx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	grades, prompt, raw, err := grader.GradeWithRaw(actx, attempt)
	cancel()
	if err != nil {
		logAIDebug(ctx, logger, mapping.ID, prompt, raw, aiLogRaw)
		logger.Error("cron_ai_grade_error", "user_test_mapping_id", mapping.ID.String(), "error", err.Error())
		return 0, false
	}
	logger.Info(
		"cron_ai_call_finish",
		"user_test_mapping_id", mapping.ID.String(),
		"duration_ms", time.Since(aiStart).Milliseconds(),
	)
	logAIDebug(ctx, logger, mapping.ID, prompt, raw, aiLogRaw)

	updatedRows := applyGrades(
		ctx,
		logger,
		db,
		mapping.ID,
		attempt.Sections,
		grades,
		answersBySection,
		maxMarksBySection,
	)

	cnt, err := db.NewSelect().
		Model((*usersmodel.UserQuestionMapping)(nil)).
		Where("user_test_mapping_id = ?", mapping.ID).
		Where("has_graded = false OR COALESCE(ai_feedback,'') = ''").
		Count(ctx)
	if err != nil {
		logger.Error("cron_ungraded_count_error", "user_test_mapping_id", mapping.ID.String(), "error", err.Error())
		return updatedRows, true
	}
	if cnt != 0 {
		return updatedRows, true
	}

	markMappingAsGraded(ctx, logger, db, mapping)
	return updatedRows, true
}

func listInitializedMappings(ctx context.Context, db *bun.DB, limit int) ([]usersmodel.UserTestMapping, error) {
	var rows []usersmodel.UserTestMapping
	err := db.NewSelect().
		Model(&rows).
		Where("status = ?", userTestStatusInitialized).
		Order("created_at ASC").
		Limit(limit).
		Scan(ctx)
	return rows, err
}

func listMappingsForGrading(ctx context.Context, db *bun.DB, limit int) ([]usersmodel.UserTestMapping, error) {
	var rows []usersmodel.UserTestMapping
	err := db.NewSelect().
		Model(&rows).
		Where("grading_completed = false").
		Where("status IN (?)", bun.In([]string{userTestStatusSubmitted, userTestStatusInGrading})).
		Order("completed_at ASC").
		Limit(limit).
		Scan(ctx)
	return rows, err
}

func listUserQuestionMappingsByUserTestMappingID(
	ctx context.Context,
	db *bun.DB,
	userTestMappingID uuid.UUID,
) ([]usersmodel.UserQuestionMapping, error) {
	var rows []usersmodel.UserQuestionMapping
	if err := db.NewSelect().
		Model(&rows).
		Where("user_test_mapping_id = ?", userTestMappingID).
		Scan(ctx); err != nil {
		return nil, err
	}
	return rows, nil
}

func totalMaxTimeMinutes(sections []testsmodel.TestSectionMapping) int {
	totalMaxTime := 0
	for _, s := range sections {
		totalMaxTime += s.MaxTime
	}
	return totalMaxTime
}

func updateUserTestMappingStatus(
	ctx context.Context,
	db *bun.DB,
	id uuid.UUID,
	status string,
	completedAt *time.Time,
) error {
	m := usersmodel.UserTestMapping{Status: status}
	if completedAt != nil {
		m.CompletedAt = *completedAt
	}
	m.ID = id

	cols := []string{"status"}
	if completedAt != nil {
		cols = append(cols, "completed_at")
	}
	_, err := db.NewUpdate().
		Model(&m).
		Column(cols...).
		WherePK().
		Exec(ctx)
	return err
}

func buildAnswersIndex(
	answers []usersmodel.UserQuestionMapping,
) (map[uuid.UUID]usersmodel.UserQuestionMapping, int) {
	answersBySection := make(map[uuid.UUID]usersmodel.UserQuestionMapping, len(answers))
	pendingCount := 0
	for _, answer := range answers {
		answersBySection[answer.TestSectionMappingID] = answer
		if !answer.HasGraded || (answer.HasGraded && answer.AIFeedback == "") {
			pendingCount++
		}
	}
	return answersBySection, pendingCount
}

func buildMaxMarksIndex(sections []testsmodel.TestSectionMapping) map[uuid.UUID]int {
	maxMarksBySection := make(map[uuid.UUID]int, len(sections))
	for _, section := range sections {
		maxMarksBySection[section.ID] = section.MaxMarks
	}
	return maxMarksBySection
}

func appendAttemptSections(
	ctx context.Context,
	logger *slog.Logger,
	db *bun.DB,
	transcriber *stt.WhisperCLI,
	userTestMappingID uuid.UUID,
	sections []testsmodel.TestSectionMapping,
	answersBySection map[uuid.UUID]usersmodel.UserQuestionMapping,
	attempt *ai.Attempt,
	pendingCount int,
) (bool, int) {
	speakingNeedsTranscript := false
	for _, s := range sections {
		a, ok := answersBySection[s.ID]
		if !ok {
			continue
		}

		if isSpeakingSection(s.Name, s.Description) {
			audioPath := ""
			if len(a.UserAnswer) > 0 {
				audioPath = strings.TrimSpace(a.UserAnswer[0])
			}
			if audioPath == "" {
				continue
			}
			if transcriber == nil {
				logger.Error("cron_transcriber_nil", "user_test_mapping_id", userTestMappingID.String())
				continue
			}

			transcript := ""
			if len(a.TestNotes) > 0 {
				transcript = strings.TrimSpace(a.TestNotes[0])
			}
			if transcript == "" {
				speakingNeedsTranscript = true
				t, err := transcribeAndPersist(ctx, logger, db, transcriber, userTestMappingID, s.ID, a.ID, audioPath)
				if err != nil {
					continue
				}
				transcript = t
				a.TestNotes = []string{transcript}
				a.HasGraded = false
				pendingCount++
			}

			attempt.Sections = append(attempt.Sections, ai.SectionAttempt{
				TestSectionMappingID: s.ID,
				SectionName:          s.Name,
				SectionDescription:   s.Description,
				MaxMarks:             s.MaxMarks,
				Questions:            a.Question,
				Answers:              []string{transcript},
				TestNotes:            []string{transcript},
			})
			continue
		}

		attempt.Sections = append(attempt.Sections, ai.SectionAttempt{
			TestSectionMappingID: s.ID,
			SectionName:          s.Name,
			SectionDescription:   s.Description,
			MaxMarks:             s.MaxMarks,
			Questions:            a.Question,
			Answers:              a.UserAnswer,
			TestNotes:            a.TestNotes,
		})
	}
	return speakingNeedsTranscript, pendingCount
}

func transcribeAndPersist(
	ctx context.Context,
	logger *slog.Logger,
	db *bun.DB,
	transcriber *stt.WhisperCLI,
	userTestMappingID uuid.UUID,
	testSectionMappingID uuid.UUID,
	userQuestionMappingID uuid.UUID,
	audioPath string,
) (string, error) {
	transcribeStart := time.Now()
	logger.Info(
		"cron_transcribe_start",
		"user_test_mapping_id", userTestMappingID.String(),
		"test_section_mapping_id", testSectionMappingID.String(),
	)
	tctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	t, err := transcriber.Transcribe(tctx, audioPath)
	cancel()
	if err != nil {
		logger.Error(
			"cron_transcribe_error",
			"user_test_mapping_id", userTestMappingID.String(),
			"test_section_mapping_id", testSectionMappingID.String(),
			"error", err.Error(),
		)
		return "", err
	}
	logger.Info(
		"cron_transcribe_finish",
		"user_test_mapping_id", userTestMappingID.String(),
		"test_section_mapping_id", testSectionMappingID.String(),
		"duration_ms", time.Since(transcribeStart).Milliseconds(),
	)

	u := usersmodel.UserQuestionMapping{
		TestNotes:     []string{t},
		HasGraded:     false,
		MarksObtained: 0,
		AIFeedback:    "",
	}
	u.ID = userQuestionMappingID
	if _, err := db.NewUpdate().
		Model(&u).
		Column("test_notes", "has_graded", "marks_obtained", "ai_feedback").
		WherePK().
		Exec(ctx); err != nil {
		logger.Error("cron_transcript_save_error", "user_question_mapping_id", userQuestionMappingID.String(), "error", err.Error())
		return "", err
	}

	return t, nil
}

func markMappingAsInGrading(ctx context.Context, logger *slog.Logger, db *bun.DB, mapping usersmodel.UserTestMapping) {
	if mapping.Status == userTestStatusInGrading {
		return
	}
	m := usersmodel.UserTestMapping{Status: userTestStatusInGrading}
	m.ID = mapping.ID
	if _, err := db.NewUpdate().
		Model(&m).
		Column("status").
		WherePK().
		Exec(ctx); err != nil {
		logger.Error("cron_mark_in_grading_error", "user_test_mapping_id", mapping.ID.String(), "error", err.Error())
	}
}

func markMappingAsGraded(ctx context.Context, logger *slog.Logger, db *bun.DB, mapping usersmodel.UserTestMapping) {
	if mapping.GradingCompleted && mapping.Status == userTestStatusGraded {
		return
	}
	m := usersmodel.UserTestMapping{GradingCompleted: true, Status: userTestStatusGraded}
	m.ID = mapping.ID
	if _, err := db.NewUpdate().
		Model(&m).
		Column("grading_completed", "status").
		WherePK().
		Exec(ctx); err != nil {
		logger.Error("cron_mark_grading_completed_error", "user_test_mapping_id", mapping.ID.String(), "error", err.Error())
	}
}

func applyGrades(
	ctx context.Context,
	logger *slog.Logger,
	db *bun.DB,
	userTestMappingID uuid.UUID,
	sectionOrder []ai.SectionAttempt,
	grades []ai.SectionGrade,
	answersBySection map[uuid.UUID]usersmodel.UserQuestionMapping,
	maxMarksBySection map[uuid.UUID]int,
) int {
	expectedOrder := make([]uuid.UUID, 0, len(sectionOrder))
	for _, s := range sectionOrder {
		expectedOrder = append(expectedOrder, s.TestSectionMappingID)
	}

	updatedRows := 0
	for i, g := range grades {
		sectionID, err := uuid.Parse(strings.TrimSpace(g.TestSectionMappingID))
		if err != nil {
			if i < len(expectedOrder) {
				sectionID = expectedOrder[i]
				logger.Warn(
					"cron_ai_invalid_section_id_fallback",
					"user_test_mapping_id", userTestMappingID.String(),
					"test_section_mapping_id", g.TestSectionMappingID,
					"fallback_test_section_mapping_id", sectionID.String(),
					"error", err.Error(),
				)
			} else {
				logger.Error(
					"cron_ai_invalid_section_id",
					"user_test_mapping_id", userTestMappingID.String(),
					"test_section_mapping_id", g.TestSectionMappingID,
					"error", err.Error(),
				)
				continue
			}
		}

		existing, ok := answersBySection[sectionID]
		if !ok {
			continue
		}
		maxMarks, ok := maxMarksBySection[sectionID]
		if !ok {
			continue
		}

		marks := g.MarksObtained
		if marks < 0 {
			marks = 0
		}
		if marks > maxMarks {
			marks = maxMarks
		}

		u := usersmodel.UserQuestionMapping{
			MarksObtained: marks,
			AIFeedback:    g.AIFeedback,
			HasGraded:     true,
		}
		u.ID = existing.ID
		if _, err := db.NewUpdate().
			Model(&u).
			Column("marks_obtained", "ai_feedback", "has_graded").
			WherePK().
			Exec(ctx); err != nil {
			logger.Error("cron_grade_update_error", "user_question_mapping_id", existing.ID.String(), "error", err.Error())
			continue
		}
		updatedRows++
	}

	return updatedRows
}

func listActiveSectionsByTestID(ctx context.Context, db *bun.DB, testID uuid.UUID) ([]testsmodel.TestSectionMapping, error) {
	var sections []testsmodel.TestSectionMapping
	if err := db.NewSelect().
		Model(&sections).
		Where("test_id = ?", testID).
		Where("is_active = true").
		Order("created_at ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	return sections, nil
}

func isSpeakingSection(name string, description string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "speak" || n == "speaking" {
		return true
	}
	d := strings.ToLower(strings.TrimSpace(description))
	return strings.Contains(d, "speak")
}

func logAIDebug(ctx context.Context, logger *slog.Logger, userTestMappingID uuid.UUID, prompt string, raw string, aiLogRaw bool) {
	if logger.Enabled(ctx, slog.LevelDebug) {
		logger.Debug("cron_ai_prompt", "user_test_mapping_id", userTestMappingID.String(), "prompt", truncate(prompt, 8000))
		logger.Debug("cron_ai_raw", "user_test_mapping_id", userTestMappingID.String(), "raw", truncate(raw, 8000))
		return
	}
	if !aiLogRaw {
		return
	}
	logger.Info("cron_ai_prompt", "user_test_mapping_id", userTestMappingID.String(), "prompt", truncate(prompt, 2000))
	logger.Info("cron_ai_raw", "user_test_mapping_id", userTestMappingID.String(), "raw", truncate(raw, 2000))
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}
