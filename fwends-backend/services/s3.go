package services

import (
	"errors"
	"fwends-backend/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3(cfg *config.S3Config) (*s3.Client, error) {

	staticResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID && region == cfg.Region {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               cfg.Endpoint,
					SigningRegion:     cfg.Region,
					HostnameImmutable: true,
				}, nil
			} else {
				return aws.Endpoint{}, errors.New("unknown endpoint requested")
			}
		},
	)

	awscfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		),
		EndpointResolverWithOptions: staticResolver,
	}

	return s3.NewFromConfig(awscfg), nil

}
