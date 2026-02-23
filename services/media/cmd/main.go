package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mathis-brgs/storm-project/services/media/internal/handlers"
	"github.com/Mathis-brgs/storm-project/services/media/internal/service"
	"github.com/Mathis-brgs/storm-project/services/media/internal/storage"
	"github.com/Mathis-brgs/storm-project/services/media/internal/subscribers"
	"github.com/nats-io/nats.go"
)

func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		if os.Getenv("KUBERNETES_SERVICE_HOST") != "" || os.Getenv("DOCKER_HOST") != "" {
			natsURL = "nats://nats:4222"
		} else {
			natsURL = "nats://localhost:4222"
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "media"
	}

	// Connexion NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Client S3/MinIO
	s3Client, err := storage.NewS3Client(bucket)
	if err != nil {
		log.Fatal(err)
	}

	mediaService := service.NewMediaService(s3Client)

	// Démarrer les subscribers NATS
	if err := subscribers.StartMediaSubscribers(nc, mediaService); err != nil {
		log.Fatal(err)
	}
	log.Printf("NATS subscribers actifs sur: %s", natsURL)

	// Démarrer le serveur HTTP
	mux := http.NewServeMux()
	handler := handlers.NewMediaHandler(mediaService)
	handler.RegisterRoutes(mux)

	go func() {
		log.Printf("HTTP server démarré sur :%s", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Media service shutting down")
}
