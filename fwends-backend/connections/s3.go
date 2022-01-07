package connections

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
)

func OpenS3() (*s3.Client, error) {

	staticResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID && region == viper.GetString("s3_region") {
				return aws.Endpoint{
					PartitionID:       "aws",
					URL:               viper.GetString("s3_endpoint"),
					SigningRegion:     viper.GetString("s3_region"),
					HostnameImmutable: true,
				}, nil
			} else {
				return aws.Endpoint{}, errors.New("unknown endpoint requested")
			}
		},
	)

	cfg := aws.Config{
		Region: viper.GetString("s3_region"),
		Credentials: credentials.NewStaticCredentialsProvider(
			viper.GetString("s3_access_key"),
			viper.GetString("s3_secret_key"),
			"",
		),
		EndpointResolverWithOptions: staticResolver,
	}

	return s3.NewFromConfig(cfg), nil

}
