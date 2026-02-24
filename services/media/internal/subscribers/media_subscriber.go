package subscribers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Mathis-brgs/storm-project/services/media/internal/service"
	"github.com/nats-io/nats.go"
)

const respondErrorLogFormat = "nats respond error: %v"

type UploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	DataBase64  string `json:"dataBase64"`
}

type DeleteRequest struct {
	MediaID string `json:"mediaId"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func StartMediaSubscribers(nc *nats.Conn, mediaService *service.MediaService) error {
	if _, err := nc.Subscribe("media.upload.requested", func(msg *nats.Msg) {
		handleUpload(msg, mediaService)
	}); err != nil {
		return err
	}

	if _, err := nc.Subscribe("media.delete.requested", func(msg *nats.Msg) {
		handleDelete(msg, mediaService)
	}); err != nil {
		return err
	}

	return nil
}

func handleUpload(msg *nats.Msg, mediaService *service.MediaService) {
	var req UploadRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		respondError(msg, "invalid json")
		return
	}

	resp, err := mediaService.Upload(context.Background(), service.UploadRequest{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		Size:        req.Size,
		DataBase64:  req.DataBase64,
	})
	if err != nil {
		respondError(msg, err.Error())
		return
	}

	payload, _ := json.Marshal(resp)
	if err := msg.Respond(payload); err != nil {
		log.Printf(respondErrorLogFormat, err)
	}
}

func handleDelete(msg *nats.Msg, mediaService *service.MediaService) {
	var req DeleteRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		respondError(msg, "invalid json")
		return
	}

	if err := mediaService.Delete(context.Background(), req.MediaID); err != nil {
		respondError(msg, err.Error())
		return
	}

	payload, _ := json.Marshal(map[string]string{"status": "deleted"})
	if err := msg.Respond(payload); err != nil {
		log.Printf(respondErrorLogFormat, err)
	}
}

func respondError(msg *nats.Msg, errMsg string) {
	payload, _ := json.Marshal(ErrorResponse{Error: errMsg})
	if err := msg.Respond(payload); err != nil {
		log.Printf(respondErrorLogFormat, err)
	}
}
