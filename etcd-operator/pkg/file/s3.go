package file

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3Uploader struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

func NewS3Uploader(Endpoint, AK, SK string) *s3Uploader {
	return &s3Uploader{
		Endpoint:        Endpoint,
		AccessKeyID:     AK,
		SecretAccessKey: SK,
	}
}

// InitClient 初使化 minio client 对象
func (su *s3Uploader) InitClient() (*minio.Client, error) {
	return minio.New(su.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(su.AccessKeyID, su.SecretAccessKey, ""),
		Secure: true,
	})
}

func (su *s3Uploader) Upload(ctx context.Context, filePath string, bucketName string, objectName string) (int64, error) {
	minioClient, err := su.InitClient()
	if err != nil {
		return 0, err
	}

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			fmt.Printf("We already own %s\n", bucketName)
		} else {
			return 0, err
		}
	} else {
		fmt.Printf("Successfully created %s\n", bucketName)
	}

	uploadInfo, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return 0, err
	}
	return uploadInfo.Size, nil
}
