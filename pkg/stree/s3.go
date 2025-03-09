package stree

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
	"log"
)

func InitS3Client(bucket, region, accessKey, secretKey, endpoint string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey, // Access Key
			secretKey, // Secret Key
			"",        // Token (optional)
		)),
		config.WithRegion(region), // Region
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           endpoint, // Endpoint (if not specified, default is used)
					SigningRegion: region,
				}, nil
			}),
		),
	)
	cfg.Region = viper.GetString("S3_REGION")
	if err != nil {
		log.Fatalf("Error loading AWS configuration: %v", err)
	}

	// Return S3 client and bucket name
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}
