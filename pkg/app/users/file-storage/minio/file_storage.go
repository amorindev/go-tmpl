package minio

import (
	"time"

	"github.com/amorindev/go-tmpl/pkg/app/users/port"
	"github.com/minio/minio-go/v7"
)

var _ port.UserFileStg = &FileStorage{}

type FileStorage struct {
	MinioClient *minio.Client
	BucketName  string
	ExpTime     time.Duration
}

func NewUserFileStg(client *minio.Client, bucketName string, expTime time.Duration) *FileStorage {
	return &FileStorage{
		MinioClient: client,
		BucketName:  bucketName,
		ExpTime:     expTime,
	}
}
