package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/Mathis-brgs/storm-project/services/media/internal/service"
)

// mockMediaService implémente MediaServiceIface pour les tests sans MinIO.
type mockMediaService struct {
	uploadResp service.UploadResponse
	uploadErr  error
	getURLResp string
	getURLErr  error
}

func (m *mockMediaService) UploadFromReader(_ context.Context, _, _ string, _ io.Reader) (service.UploadResponse, error) {
	return m.uploadResp, m.uploadErr
}

func (m *mockMediaService) GetURL(_ string) (string, error) {
	return m.getURLResp, m.getURLErr
}

// ── /health ───────────────────────────────────────────────────────────────────

func TestHealthHandler(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.healthHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("health: got %d, want 200", w.Code)
	}
}

// ── POST /media/upload ────────────────────────────────────────────────────────

func TestUploadHandler_WrongMethod(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{})
	req := httptest.NewRequest(http.MethodGet, "/media/upload", nil)
	w := httptest.NewRecorder()
	h.uploadHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("wrong method: got %d, want 405", w.Code)
	}
}

func TestUploadHandler_MissingFile(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/media/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	h.uploadHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing file: got %d, want 400", w.Code)
	}
}

func TestUploadHandler_ServiceError(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{uploadErr: errors.New("type non autorisé")})

	body, contentType := buildMultipartBody(t, "test.jpg", "image/jpeg", []byte("fake"))
	req := httptest.NewRequest(http.MethodPost, "/media/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.uploadHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("service error (non autorisé): got %d, want 400", w.Code)
	}
}

func TestUploadHandler_InternalError(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{uploadErr: errors.New("storage down")})

	body, contentType := buildMultipartBody(t, "test.jpg", "image/jpeg", []byte("fake"))
	req := httptest.NewRequest(http.MethodPost, "/media/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.uploadHandler(w, req)
	// "non autorisé" n'est pas dans l'erreur → 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("internal error: got %d, want 500", w.Code)
	}
}

func TestUploadHandler_Success(t *testing.T) {
	expected := service.UploadResponse{MediaID: "media/123_test.jpg", Key: "media/123_test.jpg", URL: "http://minio/media/123_test.jpg"}
	h := NewMediaHandler(&mockMediaService{uploadResp: expected})

	body, contentType := buildMultipartBody(t, "test.jpg", "image/jpeg", []byte("fake-image"))
	req := httptest.NewRequest(http.MethodPost, "/media/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	h.uploadHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("success upload: got %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

// ── GET /media/{key} ──────────────────────────────────────────────────────────

func TestGetMediaHandler_WrongMethod(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{})
	req := httptest.NewRequest(http.MethodPost, "/media/abc", nil)
	w := httptest.NewRecorder()
	h.getMediaHandler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("wrong method: got %d, want 405", w.Code)
	}
}

func TestGetMediaHandler_MissingID(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{})
	req := httptest.NewRequest(http.MethodGet, "/media/", nil)
	w := httptest.NewRecorder()
	h.getMediaHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing id: got %d, want 400", w.Code)
	}
}

func TestGetMediaHandler_ServiceError(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{getURLErr: errors.New("mediaId is required")})
	req := httptest.NewRequest(http.MethodGet, "/media/abc", nil)
	w := httptest.NewRecorder()
	h.getMediaHandler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("service error: got %d, want 400", w.Code)
	}
}

func TestGetMediaHandler_Success(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{getURLResp: "http://minio/media/abc"})
	req := httptest.NewRequest(http.MethodGet, "/media/abc", nil)
	w := httptest.NewRecorder()
	h.getMediaHandler(w, req)
	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("redirect: got %d, want 307", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "http://minio/media/abc" {
		t.Errorf("Location: got %q, want http://minio/media/abc", loc)
	}
}

// ── RegisterRoutes ────────────────────────────────────────────────────────────

func TestRegisterRoutes(t *testing.T) {
	h := NewMediaHandler(&mockMediaService{})
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	routes := []string{"/health", "/media/upload", "/media/"}
	for _, path := range routes {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Errorf("route %q not registered", path)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildMultipartBody crée un body multipart avec un champ "file" ayant le Content-Type voulu.
func buildMultipartBody(t *testing.T, filename, contentType string, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", contentType)
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("createPart: %v", err)
	}
	part.Write(data)
	writer.Close()

	return body, writer.FormDataContentType()
}
