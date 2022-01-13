package config

type S3Config struct {
	Endpoint    string `mapstructure:"s3_endpoint" validate:"required"`
	Region      string `mapstructure:"s3_region" validate:"required"`
	AccessKey   string `mapstructure:"s3_access_key" validate:"required"`
	SecretKey   string `mapstructure:"s3_secret_key" validate:"required"`
	MediaBucket string `mapstructure:"s3_media_bucket" validate:"required"`
}
