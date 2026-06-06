package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

type Grader struct {
	llm llms.Model
}

func NewGrader(llm llms.Model) (*Grader, error) {
	if llm == nil {
		return nil, errors.New("llm is nil")
	}
	return &Grader{llm: llm}, nil
}

func (g *Grader) Grade(ctx context.Context, attempt Attempt) ([]SectionGrade, error) {
	grades, _, _, err := g.GradeWithRaw(ctx, attempt)
	return grades, err
}

func (g *Grader) GradeWithRaw(
	ctx context.Context,
	attempt Attempt,
) ([]SectionGrade, string, string, error) {
	prompt := buildPrompt(attempt)
	out, err := llms.GenerateFromSinglePrompt(ctx, g.llm, prompt)
	if err != nil {
		return nil, prompt, "", err
	}

	jsonText, ok := extractJSONArray(strings.TrimSpace(out))
	if !ok {
		return nil, prompt, out, errors.New("model output is not a JSON array")
	}

	var grades []SectionGrade
	if err := json.Unmarshal([]byte(jsonText), &grades); err != nil {
		return nil, prompt, out, err
	}
	return grades, prompt, out, nil
}

func buildPrompt(attempt Attempt) string {
	var b strings.Builder
	b.WriteString("You are an examiner. Grade each section independently.\n")
	b.WriteString("Return ONLY valid JSON array. No markdown.\n")
	b.WriteString("marks_obtained must be between 0 and max_marks.\n")
	b.WriteString("IMPORTANT: Do NOT change test_section_mapping_id values and do NOT change the order of the JSON array.\n")
	b.WriteString("Return the same JSON array structure provided in OUTPUT_TEMPLATE with only marks_obtained and ai_feedback filled.\n\n")

	b.WriteString(fmt.Sprintf("user_test_mapping_id: %s\n", attempt.UserTestMappingID.String()))
	b.WriteString(fmt.Sprintf("test_id: %s\n\n", attempt.TestID.String()))

	for i, s := range attempt.Sections {
		b.WriteString(fmt.Sprintf("section_%d:\n", i+1))
		b.WriteString(fmt.Sprintf("test_section_mapping_id: %s\n", s.TestSectionMappingID.String()))
		b.WriteString(fmt.Sprintf("name: %s\n", s.SectionName))
		b.WriteString(fmt.Sprintf("description: %s\n", s.SectionDescription))
		b.WriteString(fmt.Sprintf("max_marks: %d\n", s.MaxMarks))
		if len(s.TestNotes) > 0 {
			b.WriteString("test_notes:\n")
			for _, n := range s.TestNotes {
				b.WriteString(fmt.Sprintf("- %s\n", n))
			}
		}
		b.WriteString("answers:\n")
		n := len(s.Questions)
		if len(s.Answers) < n {
			n = len(s.Answers)
		}
		for j := 0; j < n; j++ {
			b.WriteString(fmt.Sprintf("Q: %s\n", s.Questions[j]))
			b.WriteString(fmt.Sprintf("A: %s\n", s.Answers[j]))
		}
		b.WriteString("\n")
	}

	b.WriteString("OUTPUT_TEMPLATE:\n")
	b.WriteString("[\n")
	for i, s := range attempt.Sections {
		b.WriteString(fmt.Sprintf("  {\"test_section_mapping_id\":\"%s\",\"section_name\":\"%s\",\"max_marks\":%d,\"marks_obtained\":0,\"ai_feedback\":\"\"}", s.TestSectionMappingID.String(), s.SectionName, s.MaxMarks))
		if i < len(attempt.Sections)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("]\n\n")
	b.WriteString("Return the JSON now.\n")
	return b.String()
}

func extractJSONArray(s string) (string, bool) {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start == -1 || end == -1 || end <= start {
		return "", false
	}
	return s[start : end+1], true
}
