package main

import (
	"context"

	"cloud.google.com/go/storage"
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
	resp, err := c.b.List(c.ctx, q)
	if err != nil {
		return nil, err
	}

	// append to files
	for _, f := range resp.Results {
		files = append(files, object{
			Name:    f.Name,
			Size:    f.Size,
			BaseURL: c.BaseURL(),
		})
	}

	// recursion for the recursion god
	if resp.Next != nil {
		f, err := c.List(prefix, delimiter, marker, max, resp.Next)
		if err != nil {
			return nil, err
		}

		// append to files
		files = append(files, f...)
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
