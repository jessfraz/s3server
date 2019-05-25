package main

import (
	"cloud.google.com/go/storage"
	"github.com/mitchellh/goamz/s3"
)

type s3Provider struct {
	bucket  string
	prefix  string
	baseURL string
	client  *s3.S3
	b       *s3.Bucket
}

// List returns the files in an s3 bucket.
func (c *s3Provider) List(prefix, delimiter, marker string, max int, q *storage.Query) (files []object, err error) {
	resp, err := c.b.List(prefix, delimiter, marker, max)
	if err != nil {
		return nil, err
	}

	// append to files
	for _, f := range resp.Contents {
		files = append(files, object{
			Name:    f.Key,
			Size:    f.Size,
			BaseURL: c.BaseURL(),
		})
	}

	// recursion for the recursion god
	if resp.IsTruncated && resp.NextMarker != "" {
		f, err := c.List(resp.Prefix, resp.Delimiter, resp.NextMarker, resp.MaxKeys, q)
		if err != nil {
			return nil, err
		}

		// append to files
		files = append(files, f...)
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
