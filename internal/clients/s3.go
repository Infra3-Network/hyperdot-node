package clients

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// SimpleS3Cliet wrap minio client and it support s3-compatible storage api.
type SimpleS3Cliet struct {
	client *minio.Client
}

// NewSimpleS3Client creates a new simple client for s3.
func NewSimpleS3Client(endpoint, accessKeyID, secretAccessKey string, useSSL bool) *SimpleS3Cliet {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &SimpleS3Cliet{client: client}
}

// Make a bucket by given name. if the bucket exists, it will do nothing and return nil.
func (s *SimpleS3Cliet) MakeBucket(ctx context.Context, bucketName string) error {
	err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := s.client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			return nil
		}

		return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
	}
	log.Printf("Successfully created %s\n", bucketName)
	return nil
}

// Put object to associated bucket.
func (s *SimpleS3Cliet) Put(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64) (*minio.UploadInfo, error) {
	options := minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	}
	info, err := s.client.PutObject(ctx, bucketName, objectName, reader, objectSize, options)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// Get object from associated bucket.
func (s *SimpleS3Cliet) Get(ctx context.Context, bucketName, objectName string) (*minio.Object, error) {
	return s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
}
