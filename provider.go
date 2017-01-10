package main

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type cloud interface {
	List(prefix, delimiter, marker string, max int, q *storage.Query) ([]object, error)
	Prefix() string
	BaseURL() string
}

func newProvider(provider, bucket, s3Region, s3AccessKey, s3SecretKey string) (cloud, error) {
	if provider == "s3" {
		// auth with aws
		auth, err := aws.GetAuth(s3AccessKey, s3SecretKey)
		if err != nil {
			return nil, err
		}

		// create the client
		region, err := getRegion(s3Region)
		if err != nil {
			return nil, err
		}

		p := s3Provider{bucket: bucket}
		p.client = s3.New(auth, region)
		bucket, p.prefix = cleanBucketName(p.bucket)
		p.b = p.client.Bucket(bucket)
		p.baseURL = p.bucket + ".s3.amazonaws.com"
		return &p, nil
	}

	p := gcsProvider{bucket: bucket}
	p.ctx = context.Background()
	client, err := storage.NewClient(p.ctx)
	if err != nil {
		return nil, err
	}
	p.client = client
	p.bucket, p.prefix = cleanBucketName(p.bucket)
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

// getRegion returns the aws region that is matches a given string.
func getRegion(name string) (aws.Region, error) {
	var regions = map[string]aws.Region{
		aws.APNortheast.Name:  aws.APNortheast,
		aws.APSoutheast.Name:  aws.APSoutheast,
		aws.APSoutheast2.Name: aws.APSoutheast2,
		aws.EUCentral.Name:    aws.EUCentral,
		aws.EUWest.Name:       aws.EUWest,
		aws.USEast.Name:       aws.USEast,
		aws.USWest.Name:       aws.USWest,
		aws.USWest2.Name:      aws.USWest2,
		aws.USGovWest.Name:    aws.USGovWest,
		aws.SAEast.Name:       aws.SAEast,
	}
	region, ok := regions[name]
	if !ok {
		return aws.Region{}, fmt.Errorf("No region matches %s", name)
	}
	return region, nil
}
