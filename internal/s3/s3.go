package s3

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3 struct {
	session *session.Session
	bucket  string
}

func New(sess *session.Session, bucket string) *S3 {
	return &S3{sess, bucket}
}

func (s *S3) Upload(ctx context.Context, key string, body io.Reader) error {
	uploader := s3manager.NewUploader(s.session)

	_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

func (s *S3) URL(ctx context.Context, key string) (string, error) {
	svc := s3.New(s.session)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	req.SetContext(ctx)
	return req.Presign(10 * time.Minute)
}
