package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct {
	client     *s3.Client
	bucketName string
}

// NewS3Client initialise la connexion MinIO avec le SDK AWS v2
func NewS3Client(bucketName string) (*S3Client, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")

	if endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT manquant")
	}
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY ou MINIO_SECRET_KEY manquant")
	}

	client := s3.New(s3.Options{
		// L'endpoint doit avoir le préfixe http:// (MinIO local n'a pas de TLS)
		BaseEndpoint: aws.String("http://" + endpoint),
		Region:       "us-east-1", // Requis par le SDK mais ignoré par MinIO
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		// Obligatoire pour MinIO car il ne supporte pas le virtual-hosted style
		UsePathStyle: true,
	})

	return &S3Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadFile envoie un fichier (io.Reader) vers le bucket
func (s *S3Client) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("erreur upload S3: %w", err)
	}
	return nil
}