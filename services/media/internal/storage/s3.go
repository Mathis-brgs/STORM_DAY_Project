package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MinIOClient gère la connexion à MinIO (stockage objet local compatible S3).
// En production Azure, le media service utilise Azure Blob Storage à la place.
type MinIOClient struct {
	client     *s3.Client
	bucketName string
	endpoint   string
}

// Endpoint retourne l'URL de base MinIO (ex: http://localhost:9000)
func (s *MinIOClient) Endpoint() string {
	return s.endpoint
}

// BucketName retourne le nom du bucket
func (s *MinIOClient) BucketName() string {
	return s.bucketName
}

// GetFileURL retourne l'URL publique d'un objet
func (s *MinIOClient) GetFileURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucketName, key)
}

// NewMinIOClient initialise la connexion à MinIO via le SDK AWS v2 (compatible S3).
func NewMinIOClient(bucketName string) (*MinIOClient, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")

	if endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT manquant")
	}
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY ou MINIO_SECRET_KEY manquant")
	}

	// Si l'endpoint contient déjà un schéma (http:// ou https://), on l'utilise tel quel
	resolvedEndpoint := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		resolvedEndpoint = "http://" + endpoint
	}
	// Supprimer le slash final si présent (le SDK AWS l'ajoute lui-même)
	resolvedEndpoint = strings.TrimRight(resolvedEndpoint, "/")

	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(resolvedEndpoint),
		Region:       "us-east-1", // Requis par le SDK mais ignoré par MinIO/Azure
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: true,
	})

	return &MinIOClient{
		client:     client,
		bucketName: bucketName,
		endpoint:   resolvedEndpoint,
	}, nil
}

// UploadFile envoie un fichier (io.Reader) vers le bucket
func (s *MinIOClient) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("erreur upload MinIO: %w", err)
	}
	return nil
}

// DeleteFile supprime un objet du bucket
func (s *MinIOClient) DeleteFile(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("erreur delete MinIO: %w", err)
	}
	return nil
}