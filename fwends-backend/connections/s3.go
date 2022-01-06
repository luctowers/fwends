package connections

import (
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func OpenS3() (*s3.Client, error) {

	staticResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID && region == os.Getenv("S3_REGION") {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               os.Getenv("S3_ENDPOINT"),
					SigningRegion:     os.Getenv("S3_REGION"),
					HostnameImmutable: true,
				}, nil
			} else {
				return aws.Endpoint{}, errors.New("unknown endpoint requested")
			}
		},
	)

	cfg := aws.Config{
		Region: os.Getenv("S3_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY"),
			os.Getenv("S3_SECRET_KEY"),
			"",
		),
		EndpointResolverWithOptions: staticResolver,
	}

	return s3.NewFromConfig(cfg), nil

}
