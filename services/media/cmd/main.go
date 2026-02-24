package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mathis-brgs/storm-project/services/media/internal/service"
	"github.com/Mathis-brgs/storm-project/services/media/internal/storage"
	"github.com/Mathis-brgs/storm-project/services/media/internal/subscribers"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// j'ai modifier la fonction main pour ajouter la configuration de nats et minio via les variables d'environnement, parce que je n'arrivais pas a lancer des tests K6
func main() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		if os.Getenv("KUBERNETES_SERVICE_HOST") != "" || os.Getenv("DOCKER_HOST") != "" {
			natsURL = "nats://nats:4222"
		} else {
			natsURL = "nats://localhost:4222"
		}
	}

	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		bucket = "media"
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	s3Client, err := storage.NewS3Client(bucket)
	if err != nil {
		log.Fatal(err)
	}

	mediaService := service.NewMediaService(s3Client)

	if err := subscribers.StartMediaSubscribers(nc, mediaService); err != nil {
		log.Fatal(err)
	}

	log.Printf("Media service listening on NATS: %s", natsURL)

	// Serveur HTTP pour Prometheus /metrics
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		log.Println("Media metrics disponibles sur :8080/metrics")
		if err := http.ListenAndServe(":8080", mux); err != nil {
			log.Printf("Erreur serveur metrics: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Media service shutting down")
}
