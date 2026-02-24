package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/Mathis-brgs/storm-project/services/media/internal/storage"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
	"video/mp4":  true,
	"video/webm": true,
	"video/avi":  true,
}

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
	URL     string `json:"url"`
}

func NewMediaService(storageClient *storage.S3Client) *MediaService {
	return &MediaService{storage: storageClient}
}

func validateContentType(contentType string) error {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if !allowedMimeTypes[ct] {
		return fmt.Errorf("type de fichier non autorisé: %s (autorisés: image/jpeg, image/png, image/gif, image/webp, video/mp4, video/webm, video/avi)", ct)
	}
	return nil
}

func (s *MediaService) Upload(ctx context.Context, req UploadRequest) (UploadResponse, error) {
	if req.Filename == "" {
		return UploadResponse{}, fmt.Errorf("filename is required")
	}

	if err := validateContentType(req.ContentType); err != nil {
		return UploadResponse{}, err
	}

	key := fmt.Sprintf("media/%d_%s", time.Now().UnixNano(), req.Filename)

	var body []byte
	if req.DataBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.DataBase64)
		if err != nil {
			return UploadResponse{}, fmt.Errorf("erreur décodage base64: %w", err)
		}
		body = decoded
	}

	if err := s.storage.UploadFile(ctx, key, bytes.NewReader(body), req.ContentType); err != nil {
		return UploadResponse{}, err
	}

	url := s.storage.GetFileURL(key)
	return UploadResponse{MediaID: key, Key: key, URL: url}, nil
}

func (s *MediaService) Delete(ctx context.Context, mediaID string) error {
	if mediaID == "" {
		return fmt.Errorf("mediaId is required")
	}
	return s.storage.DeleteFile(ctx, mediaID)
}
