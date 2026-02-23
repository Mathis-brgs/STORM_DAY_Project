package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Mathis-brgs/storm-project/services/media/internal/storage"
)

// Types MIME autorisés pour l'upload
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

// ValidateContentType vérifie que le type MIME est autorisé (image ou vidéo)
func ValidateContentType(contentType string) error {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if !allowedMimeTypes[ct] {
		return fmt.Errorf("type de fichier non autorisé: %s (autorisés: image/jpeg, image/png, image/gif, image/webp, video/mp4, video/webm, video/avi)", ct)
	}
	return nil
}

// Upload via base64 (NATS)
func (s *MediaService) Upload(ctx context.Context, req UploadRequest) (UploadResponse, error) {
	if req.Filename == "" {
		return UploadResponse{}, fmt.Errorf("filename is required")
	}

	if req.ContentType != "" {
		if err := ValidateContentType(req.ContentType); err != nil {
			return UploadResponse{}, err
		}
	}

	key := fmt.Sprintf("media/%d_%s", time.Now().UnixNano(), req.Filename)

	var body io.Reader
	if req.DataBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.DataBase64)
		if err != nil {
			return UploadResponse{}, fmt.Errorf("erreur décodage base64: %w", err)
		}
		body = bytes.NewReader(decoded)
	} else {
		body = bytes.NewReader(nil)
	}

	if err := s.storage.UploadFile(ctx, key, body, req.ContentType); err != nil {
		return UploadResponse{}, err
	}

	url := s.storage.GetFileURL(key)
	return UploadResponse{MediaID: key, Key: key, URL: url}, nil
}

// UploadFromReader upload un fichier depuis un io.Reader (HTTP multipart)
func (s *MediaService) UploadFromReader(ctx context.Context, filename, contentType string, reader io.Reader) (UploadResponse, error) {
	if filename == "" {
		return UploadResponse{}, fmt.Errorf("filename is required")
	}

	if err := ValidateContentType(contentType); err != nil {
		return UploadResponse{}, err
	}

	key := fmt.Sprintf("media/%d_%s", time.Now().UnixNano(), filename)

	if err := s.storage.UploadFile(ctx, key, reader, contentType); err != nil {
		return UploadResponse{}, err
	}

	url := s.storage.GetFileURL(key)
	return UploadResponse{MediaID: key, Key: key, URL: url}, nil
}

// GetURL retourne l'URL publique d'un media par sa clé
func (s *MediaService) GetURL(mediaID string) (string, error) {
	if mediaID == "" {
		return "", fmt.Errorf("mediaId is required")
	}
	return s.storage.GetFileURL(mediaID), nil
}

func (s *MediaService) Delete(ctx context.Context, mediaID string) error {
	if mediaID == "" {
		return fmt.Errorf("mediaId is required")
	}
	return s.storage.DeleteFile(ctx, mediaID)
}
