package repository

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db                *gorm.DB
	minio             *minio.Client
	minio_endpoint    string
	minio_bucket_name string
}

type RepositorySettings struct {
	PostgresDSN     string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucketName string
}

func New(settings *RepositorySettings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(settings.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	useSSL := false

	minioClient, err := minio.New(settings.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(settings.MinioAccessKey, settings.MinioSecretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:                db,
		minio:             minioClient,
		minio_endpoint:    settings.MinioEndpoint,
		minio_bucket_name: settings.MinioBucketName,
	}, nil
}
