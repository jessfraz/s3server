package main

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type gcsProvider struct {
	bucket  string
	prefix  string
	baseURL string
	client  *storage.Client
	ctx     context.Context
	b       *storage.BucketHandle
}

// List returns the files in an gcs bucket.
func (c *gcsProvider) List(prefix, delimiter, marker string, max int, q *storage.Query) (files []object, err error) {
	resp := c.b.Objects(c.ctx, q)

	// append to files
	for {
		f, err := resp.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, object{
			Name:    f.Name,
			Size:    f.Size,
			BaseURL: c.BaseURL(),
		})
	}

	return files, nil
}

// Prefix returns the prefix in an gcs bucket.
func (c *gcsProvider) Prefix() string {
	return c.prefix
}

// BaseURL returns the baseURL in an gcs bucket.
func (c *gcsProvider) BaseURL() string {
	return c.baseURL
}
