package stt

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type WhisperCLI struct {
	Bin       string
	ModelPath string
	Language  string
}

func NewWhisperCLI(bin string, modelPath string, language string) (*WhisperCLI, error) {
	bin = strings.TrimSpace(bin)
	modelPath = strings.TrimSpace(modelPath)
	language = strings.TrimSpace(language)

	if bin == "" {
		return nil, errors.New("whisper bin is empty")
	}
	if modelPath == "" {
		return nil, errors.New("whisper model path is empty")
	}
	if language == "" {
		language = "en"
	}

	return &WhisperCLI{
		Bin:       bin,
		ModelPath: modelPath,
		Language:  language,
	}, nil
}

func (w *WhisperCLI) Transcribe(ctx context.Context, inputAudioPath string) (string, error) {
	inputAudioPath = strings.TrimSpace(inputAudioPath)
	if inputAudioPath == "" {
		return "", errors.New("input audio path is empty")
	}

	inPath, err := absPathFromCWD(inputAudioPath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(inPath); err != nil {
		return "", err
	}

	modelPath, err := absPathFromCWD(w.ModelPath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(modelPath); err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "whisper-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	wavPath := filepath.Join(tmpDir, "audio.wav")
	if err := toWav(ctx, inPath, wavPath); err != nil {
		return "", err
	}

	outBase := filepath.Join(tmpDir, "out")
	outTxt := outBase + ".txt"

	var stderr bytes.Buffer
	cmd := exec.CommandContext(
		ctx,
		w.Bin,
		"-m", modelPath,
		"-l", w.Language,
		"-nt",
		"-otxt",
		"-of", outBase,
		"-f", wavPath,
	)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return "", err
		}
		return "", errors.New(msg)
	}

	b, err := os.ReadFile(outTxt)
	if err != nil {
		return "", err
	}

	text := normalizeTranscript(string(b))
	if text == "" {
		return "", errors.New("empty transcript")
	}
	return text, nil
}

func toWav(ctx context.Context, inputPath string, outputWavPath string) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", inputPath,
		"-ar", "16000",
		"-ac", "1",
		"-c:a", "pcm_s16le",
		outputWavPath,
	)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return err
		}
		return errors.New(msg)
	}
	return nil
}

func normalizeTranscript(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, " ")
}

func absPathFromCWD(p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, p), nil
}
