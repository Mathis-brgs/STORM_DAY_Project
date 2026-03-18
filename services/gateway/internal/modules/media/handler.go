package media

import (
    "encoding/base64"
    "encoding/json"
    "io"
    "net/http"
    "time"

    "gateway/internal/common"
	"gateway/internal/modules/auth"

    "bytes"
    "log"
)

const maxUploadSize = 50 << 20 // 50 MB

type Handler struct {
    nc common.NatsConn
}

func NewHandler(nc common.NatsConn) *Handler {
    return &Handler{nc: nc}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Extraire le token du header Authorization
	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	if token == "" {
		log.Println("❌ Erreur : Token manquant dans la requête")
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	// 2. Valider le token via NATS (appelle le service user/auth sur Azure)
	valResult, err := auth.ValidateToken(h.nc, token)
	if err != nil {
		log.Printf("❌ Erreur NATS (Auth Service): %v", err)
		http.Error(w, "Authentication service unavailable", http.StatusServiceUnavailable)
		return
	}
	if !valResult.IsValid {
		log.Printf("❌ Erreur : Token invalide pour l'utilisateur %s", valResult.User.Username)
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	log.Printf("✅ Auth réussie pour l'utilisateur : %s", valResult.User.Username)

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Erreur FormFile: %v", err)
		http.Error(w, "form field 'file' required", http.StatusBadRequest)
		return
	}
	err = file.Close()
	if err != nil {
		log.Printf("error closing file: %v", err)
		http.Error(w, "internal error closing file", http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Printf("read file error: %v", err)
		http.Error(w, "internal error reading file", http.StatusInternalServerError)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Encode to base64 for NATS path
	dataBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	req := struct {
		Filename    string `json:"filename"`
		ContentType string `json:"contentType"`
		Size        int64  `json:"size"`
		DataBase64  string `json:"dataBase64"`
	}{
		Filename:    header.Filename,
		ContentType: contentType,
		Size:        int64(buf.Len()),
		DataBase64:  dataBase64,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("marshal error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Send request to media-service via NATS and wait reply
	log.Printf("Envoi de la requête d'upload vers NATS (sujet: media.upload.requested)...")
	reply, err := h.nc.Request("media.upload.requested", payload, 10*time.Second)
	if err != nil {
		log.Printf("nats request error: %v", err)
		http.Error(w, "media service unavailable: "+err.Error(), http.StatusBadGateway)
		return
	}

	// Parse reply (media service responds with JSON { mediaId, key, url } or { error: ... })
	var resp map[string]any
	if err := json.Unmarshal(reply.Data, &resp); err != nil {
		log.Printf("invalid response from media: %v", err)
		http.Error(w, "invalid response from media service", http.StatusBadGateway)
		return
	}

	if errVal, ok := resp["error"]; ok {
		// forward error message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": errVal})
		return
	}

	log.Printf("Upload réussi ! Réponse reçue : %v", resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}