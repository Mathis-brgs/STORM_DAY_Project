package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Mathis-brgs/storm-project/services/media/internal/service"
)

const maxUploadSize = 50 << 20 // 50 MB

type MediaHandler struct {
	service *service.MediaService
}

func NewMediaHandler(svc *service.MediaService) *MediaHandler {
	return &MediaHandler{service: svc}
}

// RegisterRoutes enregistre les routes HTTP du media service
func (h *MediaHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.healthHandler)
	mux.HandleFunc("/media/upload", h.uploadHandler)
	mux.HandleFunc("/media/", h.getMediaHandler)
}

func (h *MediaHandler) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// POST /media/upload — multipart file upload
func (h *MediaHandler) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "fichier trop volumineux (max 50MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "champ 'file' requis dans le formulaire multipart", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	resp, err := h.service.UploadFromReader(r.Context(), header.Filename, contentType, file)
	if err != nil {
		if strings.Contains(err.Error(), "non autorisé") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("upload error: %v", err)
		http.Error(w, "erreur interne lors de l'upload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// GET /media/{key...} — retourne l'URL du fichier ou redirige
func (h *MediaHandler) getMediaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraire la clé depuis l'URL: /media/xxx → xxx
	mediaID := strings.TrimPrefix(r.URL.Path, "/media/")
	if mediaID == "" {
		http.Error(w, "mediaId requis dans l'URL: /media/{id}", http.StatusBadRequest)
		return
	}

	url, err := h.service.GetURL(mediaID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Redirige vers l'URL MinIO/S3 du fichier
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
