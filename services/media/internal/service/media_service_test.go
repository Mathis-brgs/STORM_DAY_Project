package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

// mockStorage implémente StorageClient pour les tests (sans MinIO réel).
type mockStorage struct {
	uploadErr error
	deleteErr error
	baseURL   string
}

func (m *mockStorage) UploadFile(_ context.Context, _ string, _ io.Reader, _ string) error {
	return m.uploadErr
}

func (m *mockStorage) DeleteFile(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockStorage) GetFileURL(key string) string {
	return m.baseURL + "/" + key
}

// ── ValidateContentType ───────────────────────────────────────────────────────

func TestValidateContentType_Allowed(t *testing.T) {
	allowed := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"video/mp4", "video/webm", "video/avi",
	}
	for _, ct := range allowed {
		if err := ValidateContentType(ct); err != nil {
			t.Errorf("ValidateContentType(%q) = %v, want nil", ct, err)
		}
	}
}

func TestValidateContentType_CaseInsensitive(t *testing.T) {
	cases := []string{"IMAGE/JPEG", "Image/PNG", "VIDEO/MP4"}
	for _, ct := range cases {
		if err := ValidateContentType(ct); err != nil {
			t.Errorf("ValidateContentType(%q) should be allowed, got %v", ct, err)
		}
	}
}

func TestValidateContentType_Rejected(t *testing.T) {
	rejected := []string{
		"application/pdf", "text/html", "application/octet-stream",
		"", "image/tiff", "application/zip",
	}
	for _, ct := range rejected {
		if err := ValidateContentType(ct); err == nil {
			t.Errorf("ValidateContentType(%q) = nil, want error", ct)
		}
	}
}

// ── Upload ────────────────────────────────────────────────────────────────────

func TestUpload_EmptyFilename(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	_, err := svc.Upload(context.Background(), UploadRequest{
		Filename:    "",
		ContentType: "image/jpeg",
	})
	if err == nil {
		t.Fatal("Upload with empty filename should return error")
	}
}

func TestUpload_InvalidContentType(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	_, err := svc.Upload(context.Background(), UploadRequest{
		Filename:    "test.pdf",
		ContentType: "application/pdf",
	})
	if err == nil {
		t.Fatal("Upload with invalid content type should return error")
	}
}

func TestUpload_StorageError(t *testing.T) {
	svc := NewMediaService(&mockStorage{uploadErr: errors.New("s3 unavailable")})
	_, err := svc.Upload(context.Background(), UploadRequest{
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
		DataBase64:  "",
	})
	if err == nil {
		t.Fatal("Upload should propagate storage error")
	}
}

func TestUpload_Success(t *testing.T) {
	svc := NewMediaService(&mockStorage{baseURL: "http://minio:9000/media"})
	resp, err := svc.Upload(context.Background(), UploadRequest{
		Filename:    "photo.png",
		ContentType: "image/png",
	})
	if err != nil {
		t.Fatalf("Upload unexpectedly failed: %v", err)
	}
	if resp.Key == "" {
		t.Error("Upload response should contain a Key")
	}
	if !strings.Contains(resp.URL, "photo.png") {
		t.Errorf("URL %q should reference the filename", resp.URL)
	}
}

func TestUpload_WithBase64Data(t *testing.T) {
	// "aGVsbG8=" = base64("hello")
	svc := NewMediaService(&mockStorage{baseURL: "http://minio:9000/media"})
	_, err := svc.Upload(context.Background(), UploadRequest{
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
		DataBase64:  "aGVsbG8=",
	})
	if err != nil {
		t.Fatalf("Upload with base64 data failed: %v", err)
	}
}

func TestUpload_InvalidBase64(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	_, err := svc.Upload(context.Background(), UploadRequest{
		Filename:    "test.jpg",
		ContentType: "image/jpeg",
		DataBase64:  "not-valid-base64!!!",
	})
	if err == nil {
		t.Fatal("Upload with invalid base64 should return error")
	}
}

// ── UploadFromReader ──────────────────────────────────────────────────────────

func TestUploadFromReader_EmptyFilename(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	_, err := svc.UploadFromReader(context.Background(), "", "image/jpeg", bytes.NewReader([]byte("data")))
	if err == nil {
		t.Fatal("UploadFromReader with empty filename should return error")
	}
}

func TestUploadFromReader_InvalidContentType(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	_, err := svc.UploadFromReader(context.Background(), "file.exe", "application/octet-stream", bytes.NewReader([]byte("data")))
	if err == nil {
		t.Fatal("UploadFromReader with invalid content type should return error")
	}
}

func TestUploadFromReader_Success(t *testing.T) {
	svc := NewMediaService(&mockStorage{baseURL: "http://minio:9000/media"})
	resp, err := svc.UploadFromReader(context.Background(), "img.webp", "image/webp", bytes.NewReader([]byte("fake-img")))
	if err != nil {
		t.Fatalf("UploadFromReader failed: %v", err)
	}
	if resp.MediaID == "" {
		t.Error("Response should contain a MediaID")
	}
}

// ── GetURL ────────────────────────────────────────────────────────────────────

func TestGetURL_EmptyMediaID(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	_, err := svc.GetURL("")
	if err == nil {
		t.Fatal("GetURL with empty mediaID should return error")
	}
}

func TestGetURL_Success(t *testing.T) {
	svc := NewMediaService(&mockStorage{baseURL: "http://minio:9000/media"})
	url, err := svc.GetURL("media/12345_photo.jpg")
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}
	if url == "" {
		t.Error("GetURL should return a non-empty URL")
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_EmptyMediaID(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	err := svc.Delete(context.Background(), "")
	if err == nil {
		t.Fatal("Delete with empty mediaID should return error")
	}
}

func TestDelete_StorageError(t *testing.T) {
	svc := NewMediaService(&mockStorage{deleteErr: errors.New("object not found")})
	err := svc.Delete(context.Background(), "media/12345_photo.jpg")
	if err == nil {
		t.Fatal("Delete should propagate storage error")
	}
}

func TestDelete_Success(t *testing.T) {
	svc := NewMediaService(&mockStorage{})
	err := svc.Delete(context.Background(), "media/12345_photo.jpg")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
