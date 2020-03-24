package main

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type cloud interface {
	List(prefix, delimiter, marker string, max int, q *storage.Query) ([]object, error)
	Prefix() string
	BaseURL() string
}

func newProvider(provider, bucket, s3Endpoint, s3Region, s3AccessKey, s3SecretKey string) (cloud, error) {
	bucket, prefix := cleanBucketName(bucket)

	if provider == "s3" {
		// auth with aws
		conf := newAwsConfig(s3Endpoint, s3Region, s3AccessKey, s3SecretKey)

		// create the client
		p := s3Provider{bucket: bucket, prefix: prefix}
		p.client = s3.New(awsSession.New(conf))
		p.baseURL = fmt.Sprintf("%s.%s", p.bucket, *conf.Endpoint)
		return &p, nil
	}

	p := gcsProvider{bucket: bucket, prefix: prefix}
	p.ctx = context.Background()
	client, err := storage.NewClient(p.ctx)
	if err != nil {
		return nil, err
	}
	p.client = client
	p.b = client.Bucket(p.bucket)
	p.baseURL = p.bucket
	if !strings.Contains(p.bucket, "j3ss.co") {
		p.baseURL += ".storage.googleapis.com"
	}
	return &p, nil
}

// cleanBucketName returns the bucket and prefix
// for a given s3bucket.
func cleanBucketName(bucket string) (string, string) {
	bucket = strings.TrimPrefix(bucket, "s3://")
	bucket = strings.TrimPrefix(bucket, "gcs://")
	parts := strings.SplitN(bucket, "/", 2)
	if len(parts) == 1 {
		return bucket, "/"
	}

	return parts[0], parts[1]
}

func newAwsConfig(endpoint, region, accessKey, secretKey string) *aws.Config {
	conf := aws.NewConfig()
	conf.WithEndpoint(fmt.Sprintf("s3.%s.%s", region, endpoint))
	conf.WithRegion(region)
	conf.WithCredentials(awsCredentials.NewChainCredentials([]awsCredentials.Provider{
		&awsCredentials.StaticProvider{
			Value: awsCredentials.Value{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			},
		},
		&awsCredentials.EnvProvider{},
		&awsCredentials.SharedCredentialsProvider{},
	}))
	return conf
}
