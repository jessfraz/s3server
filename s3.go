package main

import (
	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

type s3Provider struct {
	bucket  string
	prefix  string
	baseURL string
	client  *s3.S3
}

// List returns the files in an s3 bucket.
func (c *s3Provider) List(prefix, delimiter, marker string, max int, q *storage.Query) (files []object, err error) {
	err = c.client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Delimiter: aws.String(delimiter),
		Prefix: aws.String(prefix),
		MaxKeys: aws.Int64(int64(max)),
	}, func(p *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, o := range p.Contents {
			files = append(files, object{
				Name:    aws.StringValue(o.Key),
				Size:    aws.Int64Value(o.Size),
				BaseURL: c.baseURL,
			})
		}

		return true // continue paging
	})

 	if err != nil {
		logrus.Fatalf("Failed to list objects for bucket, %s, %v", c.bucket, err)
	}

	return files, nil
}

// Prefix returns the prefix in an s3 bucket.
func (c *s3Provider) Prefix() string {
	return c.prefix
}

// BaseURL returns the baseURL in an s3 bucket.
func (c *s3Provider) BaseURL() string {
	return c.baseURL
}
