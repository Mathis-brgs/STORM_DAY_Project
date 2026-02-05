package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Mathis-brgs/storm-project/services/media/internal/storage"
)

func main() {
	log.Println("🚀 Démarrage du Media Service...")

	// 1. Initialisation du client S3 (MinIO)
	// On utilise un bucket nommé "media-bucket" (à créer via l'interface MinIO si besoin)
	s3Client, err := storage.NewS3Client("media-bucket")
	if err != nil {
		log.Fatalf("❌ Erreur initialisation S3: %v", err)
	}

	// 2. Exécution de la tâche de test (Upload Hello World)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("📤 Tentative d'upload du fichier test...")

	testContent := "Hello World ! Ceci est un test depuis le SDK AWS Go v2."
	err = s3Client.UploadFile(
		ctx,
		"test/hello-world.txt", // Chemin/Nom du fichier dans le bucket
		strings.NewReader(testContent),
		"text/plain",
	)

	if err != nil {
		log.Fatalf("❌ Échec de l'upload : %v", err)
	}

	log.Println("✅ Succès ! Le fichier 'test/hello-world.txt' est sur MinIO.")

	// Ici, ton service resterait normalement en écoute (HTTP/gRPC)
	// Pour le test, on s'arrête là ou on bloque avec un select{}
	select {}
}
