package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"cloud.google.com/go/storage"
	"github.com/Sirupsen/logrus"
	units "github.com/docker/go-units"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
	"golang.org/x/net/context"
)

var (
	provider string
	bucket   string

	s3AccessKey string
	s3SecretKey string
	s3Region    string

	port     string
	certFile string
	keyFile  string
)

func init() {
	flag.StringVar(&provider, "provider", "s3", "cloud provider (ex. s3, gcs)")
	flag.StringVar(&bucket, "bucket", "", "bucket path from which to serve files")

	flag.StringVar(&s3AccessKey, "s3key", "", "s3 access key")
	flag.StringVar(&s3SecretKey, "s3secret", "", "s3 access secret")
	flag.StringVar(&s3Region, "s3region", "us-west-2", "aws region for the bucket")

	flag.StringVar(&port, "p", "8080", "port for server to run on")

	flag.StringVar(&certFile, "cert", "", "path to ssl certificate")
	flag.StringVar(&keyFile, "key", "", "path to ssl key")

	flag.Parse()

	if provider != "s3" && provider != "gcs" {
		logrus.Fatalf("%s is not a valid provider, try `s3` or `gcs`.", provider)
	}
}

func main() {
	// create a new provider
	p, err := newProvider(provider, bucket, s3Region, s3AccessKey, s3SecretKey)
	if err != nil {
		logrus.Fatalf("Creating new provider failed: %v", err)
	}

	// get the files
	max := 2000
	q := &storage.Query{
		Prefix:     p.Prefix(),
		MaxResults: max,
	}
	files, err := p.List(p.Prefix(), p.Prefix(), "", max, q)
	if err != nil {
		logrus.Fatalf("Listing all files in bucket failed: %v", err)
	}

	// create mux server
	mux := http.NewServeMux()

	// static files handler
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("/src/static")))
	mux.Handle("/static/", staticHandler)

	// template handler
	h := handler{
		Files: files,
	}
	mux.HandleFunc("/", h.serveTemplate)

	// set up the server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	logrus.Infof("Starting server on port %q", port)
	if certFile != "" && keyFile != "" {
		logrus.Fatal(server.ListenAndServeTLS(certFile, keyFile))
	} else {
		logrus.Fatal(server.ListenAndServe())
	}
}

type object struct {
	Name    string
	BaseURL string
	Size    int64
}

type s3Provider struct {
	bucket  string
	prefix  string
	baseURL string
	client  *s3.S3
	ctx     context.Context
	b       *s3.Bucket
}

type gcsProvider struct {
	bucket  string
	prefix  string
	baseURL string
	client  *storage.Client
	ctx     context.Context
	b       *storage.BucketHandle
}

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

// JSONResponse is a map[string]string
// response from the web server.
type JSONResponse map[string]string

// String returns the string representation of the
// JSONResponse object.
func (j JSONResponse) String() string {
	str, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{
  "error": "%v"
}`, err)
	}

	return string(str)
}

// handler is the object which contains data to pass to the http handler functions.
type handler struct {
	Files []object
}

func (h *handler) serveTemplate(w http.ResponseWriter, r *http.Request) {
	templateDir := path.Join("/src", "templates")
	lp := path.Join(templateDir, "layout.html")

	// set up custom functions
	funcMap := template.FuncMap{
		"ext": func(name string) string {
			return strings.TrimPrefix(filepath.Ext(name), ".")
		},
		"base": func(name string) string {
			parts := strings.Split(name, "/")
			return parts[len(parts)-1]
		},
		"size": func(s int64) string {
			return units.HumanSize(float64(s))
		},
	}

	// parse & execute the template
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(lp))
	if err := tmpl.ExecuteTemplate(w, "layout", h.Files); err != nil {
		writeError(w, fmt.Sprintf("Execute template failed: %v", err))
		return
	}
}

// writeError sends an error back to the requester
// and also logs the error.
func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, JSONResponse{
		"error": msg,
	})
	logrus.Printf("writing error: %s", msg)
	return
}
