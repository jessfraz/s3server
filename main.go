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

	"github.com/Sirupsen/logrus"
	units "github.com/docker/go-units"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

var (
	s3Bucket    string
	s3AccessKey string
	s3SecretKey string
	s3Region    string
	port        string
	certFile    string
	keyFile     string
)

// cleanBucketName returns the bucket and prefix
// for a given s3bucket.
func cleanBucketName(bucket string) (string, string) {
	bucket = strings.TrimPrefix(bucket, "s3://")
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

// listFiles lists the files in a specific s3 bucket.
func listFiles(prefix, delimiter, marker string, maxKeys int, b *s3.Bucket) (files []s3.Key, err error) {
	resp, err := b.List(prefix, delimiter, marker, maxKeys)
	if err != nil {
		return nil, err
	}

	// append to files
	files = append(files, resp.Contents...)

	// recursion for the recursion god
	if resp.IsTruncated && resp.NextMarker != "" {
		f, err := listFiles(resp.Prefix, resp.Delimiter, resp.NextMarker, resp.MaxKeys, b)
		if err != nil {
			return nil, err
		}

		// append to files
		files = append(files, f...)
	}

	return files, nil
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

// Handler is the object which contains data to pass to the http handler functions.
type Handler struct {
	Files []s3.Key
}

func (h *Handler) serveTemplate(w http.ResponseWriter, r *http.Request) {
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

func init() {
	flag.StringVar(&s3Bucket, "s3bucket", "", "bucket path from which to serve files")
	flag.StringVar(&s3AccessKey, "s3key", "", "s3 access key")
	flag.StringVar(&s3SecretKey, "s3secret", "", "s3 access secret")
	flag.StringVar(&s3Region, "s3region", "us-west-2", "aws region for the bucket")
	flag.StringVar(&port, "p", "8080", "port for server to run on")

	flag.StringVar(&certFile, "cert", "", "path to ssl certificate")
	flag.StringVar(&keyFile, "key", "", "path to ssl key")
	flag.Parse()
}

func main() {
	// auth with aws
	auth, err := aws.GetAuth(s3AccessKey, s3SecretKey)
	if err != nil {
		logrus.Fatalf("Could not auth to AWS: %v", err)
	}

	// create the client
	region, err := getRegion(s3Region)
	if err != nil {
		logrus.Fatal(err)
	}
	client := s3.New(auth, region)

	// get the files in the bucket
	bucket, prefix := cleanBucketName(s3Bucket)
	// get the bucket
	b := client.Bucket(bucket)
	files, err := listFiles(prefix, prefix, "", 2000, b)
	if err != nil {
		logrus.Fatalf("Listing all files in bucket failed: %v", err)
	}

	// create mux server
	mux := http.NewServeMux()

	// static files handler
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("/src/static")))
	mux.Handle("/static/", staticHandler)

	// template handler
	h := Handler{
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
