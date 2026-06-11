package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"ai_testing/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ResolvedAudio struct {
	Reference   string
	S3Bucket    string
	S3Key       string
	FileName    string
	ContentType string
}

type AudioStore struct {
	s3Region          string
	s3Bucket          string
	s3AccessKeyID     string
	s3SecretAccessKey string
	s3Prefix          string

	s3Once sync.Once
	s3Cli  *s3.Client
	s3Err  error
}

func NewAudioStore(cfg config.Config) *AudioStore {
	return &AudioStore{
		s3Region:          strings.TrimSpace(cfg.AWSS3Region),
		s3Bucket:          strings.TrimSpace(cfg.AWSS3Bucket),
		s3AccessKeyID:     strings.TrimSpace(cfg.AWSS3AccessKeyID),
		s3SecretAccessKey: strings.TrimSpace(cfg.AWSS3SecretAccessKey),
		s3Prefix:          strings.Trim(strings.TrimSpace(cfg.AWSS3Prefix), "/"),
	}
}

func (s *AudioStore) Save(ctx context.Context, fileHeader *multipart.FileHeader, userID uuid.UUID, userTestMappingID uuid.UUID, sectionID uuid.UUID) (string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".webm"
	}

	fileName := uuid.NewString() + ext
	return s.saveToS3(ctx, fileHeader, userID, userTestMappingID, sectionID, fileName)
}

func (s *AudioStore) Delete(ctx context.Context, ref string) error {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	resolved, err := s.Resolve(ref, uuid.Nil, uuid.Nil, uuid.Nil)
	if err != nil {
		return err
	}
	client, err := s.s3Client(ctx)
	if err != nil {
		return err
	}
	_, err = client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(resolved.S3Bucket),
		Key:    aws.String(resolved.S3Key),
	})
	return err
}

func (s *AudioStore) Resolve(ref string, userID uuid.UUID, userTestMappingID uuid.UUID, sectionID uuid.UUID) (*ResolvedAudio, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, errors.New("audio not found")
	}
	if !strings.HasPrefix(ref, "s3://") {
		return nil, errors.New("audio reference must use s3")
	}
	return s.resolveS3(ref, userID, userTestMappingID, sectionID)
}

func (s *AudioStore) Open(ctx context.Context, resolved *ResolvedAudio) (io.ReadCloser, string, error) {
	if resolved == nil {
		return nil, "", errors.New("audio not found")
	}

	client, err := s.s3Client(ctx)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(resolved.S3Bucket),
		Key:    aws.String(resolved.S3Key),
	})
	if err != nil {
		return nil, "", err
	}

	contentType := strings.TrimSpace(resolved.ContentType)
	if contentType == "" && resp.ContentType != nil {
		contentType = strings.TrimSpace(*resp.ContentType)
	}
	if contentType == "" {
		contentType = detectContentType(resolved.FileName)
	}
	return resp.Body, contentType, nil
}

func (s *AudioStore) saveToS3(ctx context.Context, fileHeader *multipart.FileHeader, userID uuid.UUID, userTestMappingID uuid.UUID, sectionID uuid.UUID, fileName string) (string, error) {
	if s.s3Bucket == "" {
		return "", errors.New("AWS_S3_BUCKET is required")
	}
	client, err := s.s3Client(ctx)
	if err != nil {
		return "", err
	}

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	key := s.objectKey(userID, userTestMappingID, sectionID, fileName)
	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = detectContentType(fileName)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.s3Bucket),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	return "s3://" + s.s3Bucket + "/" + key, nil
}

func (s *AudioStore) resolveS3(ref string, userID uuid.UUID, userTestMappingID uuid.UUID, sectionID uuid.UUID) (*ResolvedAudio, error) {
	if s.s3Bucket == "" {
		return nil, errors.New("AWS_S3_BUCKET is required")
	}
	withoutScheme := strings.TrimPrefix(ref, "s3://")
	parts := strings.SplitN(withoutScheme, "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return nil, errors.New("invalid s3 audio path")
	}

	bucket := strings.TrimSpace(parts[0])
	key := strings.TrimSpace(parts[1])
	if bucket != s.s3Bucket {
		return nil, errors.New("audio bucket mismatch")
	}

	expectedPrefix := s.objectKeyPrefix(userID, userTestMappingID, sectionID)
	if userID != uuid.Nil && !strings.HasPrefix(key, expectedPrefix+"/") {
		return nil, errors.New("audio does not belong to the user")
	}

	return &ResolvedAudio{
		Reference:   ref,
		S3Bucket:    bucket,
		S3Key:       key,
		FileName:    path.Base(key),
		ContentType: detectContentType(key),
	}, nil
}

func (s *AudioStore) objectKey(userID uuid.UUID, userTestMappingID uuid.UUID, sectionID uuid.UUID, fileName string) string {
	return s.objectKeyPrefix(userID, userTestMappingID, sectionID) + "/" + fileName
}

func (s *AudioStore) objectKeyPrefix(userID uuid.UUID, userTestMappingID uuid.UUID, sectionID uuid.UUID) string {
	parts := []string{}
	if s.s3Prefix != "" {
		parts = append(parts, s.s3Prefix)
	}
	parts = append(parts, "storage", "audio")
	if userID != uuid.Nil {
		parts = append(parts, userID.String(), userTestMappingID.String(), sectionID.String())
	}
	return path.Clean(path.Join(parts...))
}

func (s *AudioStore) s3Client(ctx context.Context) (*s3.Client, error) {
	s.s3Once.Do(func() {
		if s.s3Region == "" {
			s.s3Err = errors.New("AWS_S3_REGION is required")
			return
		}

		opts := []func(*awsconfig.LoadOptions) error{
			awsconfig.WithRegion(s.s3Region),
		}
		if s.s3AccessKeyID != "" && s.s3SecretAccessKey != "" {
			opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				s.s3AccessKeyID,
				s.s3SecretAccessKey,
				"",
			)))
		}

		cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
		if err != nil {
			s.s3Err = err
			return
		}
		s.s3Cli = s3.NewFromConfig(cfg)
	})
	if s.s3Err != nil {
		return nil, s.s3Err
	}
	return s.s3Cli, nil
}

func detectContentType(name string) string {
	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(name)))
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}

func (r *ResolvedAudio) String() string {
	if r == nil {
		return ""
	}
	return fmt.Sprintf("s3://%s/%s", r.S3Bucket, r.S3Key)
}
