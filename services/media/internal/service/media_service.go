package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/Mathis-brgs/storm-project/services/media/internal/storage"
)

type MediaService struct {
	storage *storage.S3Client
}

type UploadRequest struct {
	Filename    string
	ContentType string
	Size        int64
	DataBase64  string
}

type UploadResponse struct {
	MediaID string `json:"mediaId"`
	Key     string `json:"key"`
}

func NewMediaService(storageClient *storage.S3Client) *MediaService {
	return &MediaService{storage: storageClient}
}

func (s *MediaService) Upload(ctx context.Context, req UploadRequest) (UploadResponse, error) {
	if req.Filename == "" {
		return UploadResponse{}, fmt.Errorf("filename is required")
	}

	key := fmt.Sprintf("media/%d_%s", time.Now().UnixNano(), req.Filename)

	// Placeholder: upload an empty object if no binary is provided.
	body := bytes.NewReader(nil)
	if err := s.storage.UploadFile(ctx, key, body, req.ContentType); err != nil {
		return UploadResponse{}, err
	}

	return UploadResponse{MediaID: key, Key: key}, nil
}

func (s *MediaService) Delete(ctx context.Context, mediaID string) error {
	if mediaID == "" {
		return fmt.Errorf("mediaId is required")
	}
	return s.storage.DeleteFile(ctx, mediaID)
}
