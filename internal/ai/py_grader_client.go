package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type PySectionAttempt struct {
	TestSectionMappingID string   `json:"test_section_mapping_id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	MaxMarks             int      `json:"max_marks"`
	Questions            []string `json:"questions"`
	Answers              []string `json:"answers"`
	TestNotes            []string `json:"test_notes"`
}

type PyGradeRequest struct {
	UserTestMappingID string             `json:"user_test_mapping_id"`
	TestID            string             `json:"test_id"`
	Sections          []PySectionAttempt `json:"sections"`
}

type PySectionResult struct {
	TestSectionMappingID string  `json:"test_section_mapping_id"`
	MarksObtained        int     `json:"marks_obtained"`
	AIFeedback           string  `json:"ai_feedback"`
	Transcript           *string `json:"transcript"`
}

type PyGradeResponse struct {
	Sections []PySectionResult `json:"sections"`
}

type PyGraderClient struct {
	BaseURL    string
	Token      string
	TimeoutSec int
	HTTP       *http.Client
}

func NewPyGraderClient(baseURL string, token string, timeoutSec int) *PyGraderClient {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeoutSec <= 0 {
		timeout = 180 * time.Second
	}
	return &PyGraderClient{
		BaseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		Token:      strings.TrimSpace(token),
		TimeoutSec: timeoutSec,
		HTTP: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *PyGraderClient) Grade(ctx context.Context, payload PyGradeRequest) (*PyGradeResponse, string, error) {
	if c == nil || c.HTTP == nil {
		return nil, "", errors.New("grader client not initialized")
	}
	if c.BaseURL == "" {
		return nil, "", errors.New("grader url is empty")
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/grade", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("X-Grader-Token", c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	rawBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	raw := string(rawBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, raw, errors.New("grader request failed")
	}

	var out PyGradeResponse
	if err := json.Unmarshal(rawBytes, &out); err != nil {
		return nil, raw, err
	}
	return &out, raw, nil
}
