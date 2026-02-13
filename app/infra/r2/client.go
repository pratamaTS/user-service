package r2

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	S3            *s3.Client
	Bucket        string
	PublicBaseURL string
}

func New() (*Client, error) {
	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET")
	publicBase := os.Getenv("R2_PUBLIC_BASE_URL")

	log.Print("R2_ENDPOINT:", endpoint)
	log.Print("R2_BUCKET:", bucket)
	log.Print("R2_PUBLIC_BASE_URL:", publicBase)
	log.Print("R2_ACCESS_KEY_ID:", accessKey)
	log.Print("R2_SECRET_ACCESS_KEY:", secretKey)

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" || publicBase == "" {
		return nil, fmt.Errorf("missing R2 env config")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // penting untuk S3-compatible (R2)
	})

	return &Client{
		S3:            s3Client,
		Bucket:        bucket,
		PublicBaseURL: publicBase,
	}, nil
}
