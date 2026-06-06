package ai

import "github.com/google/uuid"

type SectionAttempt struct {
	TestSectionMappingID uuid.UUID
	SectionName          string
	SectionDescription   string
	MaxMarks             int
	Questions            []string
	Answers              []string
	TestNotes            []string
}

type Attempt struct {
	UserTestMappingID uuid.UUID
	TestID            uuid.UUID
	Sections          []SectionAttempt
}

type SectionGrade struct {
	TestSectionMappingID string `json:"test_section_mapping_id"`
	MarksObtained        int    `json:"marks_obtained"`
	AIFeedback           string `json:"ai_feedback"`
}
