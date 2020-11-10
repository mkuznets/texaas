package opts

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// AWS contains credentials and parameters for AWS APIs (particularly, S3)
type AWS struct {
	Region    string `long:"region" env:"REGION" description:"region id" required:"true"`
	KeyID     string `long:"access-key-id" env:"ACCESS_KEY_ID" description:"access id" required:"true"`
	SecretKey string `long:"secret-access-key" env:"SECRET_ACCESS_KEY" description:"access secret" required:"true"`
	Bucket    string `long:"s3-bucket" env:"S3_BUCKET" description:"S3 bucket name" required:"true"`
}

// Session returns AWS SDK session based on the given credentials
func (a *AWS) Session() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(a.Region),
		Credentials: credentials.NewStaticCredentials(a.KeyID, a.SecretKey, ""),
	}))
}
