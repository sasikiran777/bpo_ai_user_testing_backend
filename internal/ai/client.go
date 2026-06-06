package ai

import (
	"ai_testing/internal/config"
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func NewChutesLLM(cfg *config.Config) (llms.Model, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if cfg.ChutesAIURL == "" {
		return nil, errors.New("CHUTES_AI_URL is required")
	}
	if cfg.ChutesAIAPIKey == "" {
		return nil, errors.New("CHUTES_AI_API_KEY is required")
	}
	if cfg.DeepseekModel == "" {
		return nil, errors.New("DEEPSEEK_MODEL is required")
	}

	return openai.New(
		openai.WithBaseURL(cfg.ChutesAIURL),
		openai.WithToken(cfg.ChutesAIAPIKey),
		openai.WithModel(cfg.DeepseekModel),
	)
}
