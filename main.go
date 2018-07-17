package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
	"time"

	"cloud.google.com/go/storage"
	units "github.com/docker/go-units"
	"github.com/genuinetools/pkg/cli"
	"github.com/jessfraz/s3server/version"
	"github.com/sirupsen/logrus"
)

var (
	provider string
	bucket   string
	interval time.Duration

	s3AccessKey string
	s3SecretKey string
	s3Region    string

	port     string
	certFile string
	keyFile  string

	updating bool

	debug bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "s3server"
	p.Description = "Server to index & view files in a s3 or Google Cloud Storage bucket"

	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("global", flag.ExitOnError)
	p.FlagSet.StringVar(&provider, "provider", "s3", "cloud provider (ex. s3, gcs)")
	p.FlagSet.StringVar(&bucket, "bucket", "", "bucket path from which to serve files")
	p.FlagSet.DurationVar(&interval, "interval", 5*time.Minute, "interval to generate new index.html's at")

	p.FlagSet.StringVar(&s3AccessKey, "s3key", "", "s3 access key")
	p.FlagSet.StringVar(&s3SecretKey, "s3secret", "", "s3 access secret")
	p.FlagSet.StringVar(&s3Region, "s3region", "us-west-2", "aws region for the bucket")

	p.FlagSet.StringVar(&port, "p", "8080", "port for server to run on")

	p.FlagSet.StringVar(&certFile, "cert", "", "path to ssl certificate")
	p.FlagSet.StringVar(&keyFile, "key", "", "path to ssl key")

	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if provider != "s3" && provider != "gcs" {
			return fmt.Errorf("%s is not a valid provider, try `s3` or `gcs`", provider)
		}

		return nil
	}

	// Set the main program action.
	p.Action = func(ctx context.Context, args []string) error {
		ticker := time.NewTicker(interval)

		// On ^C, or SIGTERM handle exit.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		go func() {
			for sig := range c {
				ticker.Stop()
				logrus.Infof("Received %s, exiting.", sig.String())
				os.Exit(0)
			}
		}()

		// create a new provider
		p, err := newProvider(provider, bucket, s3Region, s3AccessKey, s3SecretKey)
		if err != nil {
			logrus.Fatalf("Creating new provider failed: %v", err)
		}

		// get the path to the static directory
		wd, err := os.Getwd()
		if err != nil {
			logrus.Fatalf("Getting working directory failed: %v", err)
		}
		staticDir := filepath.Join(wd, "static")

		// create the initial index
		if err := createStaticIndex(p, staticDir); err != nil {
			logrus.Fatalf("Creating initial static index failed: %v", err)
		}

		go func() {
			// create more indexes every X minutes based off interval
			for range ticker.C {
				if !updating {
					if err := createStaticIndex(p, staticDir); err != nil {
						logrus.Warnf("creating static index failed: %v", err)
						updating = false
					}
				}
			}
		}()

		// create mux server
		mux := http.NewServeMux()

		// static files handler
		staticHandler := http.FileServer(http.Dir(staticDir))
		mux.Handle("/", staticHandler)

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
		return nil
	}

	// Run our program.
	p.Run()
}

type object struct {
	Name    string
	BaseURL string
	Size    int64
}

type data struct {
	SiteURL     string
	LastUpdated string
	Files       []object
}

func createStaticIndex(p cloud, staticDir string) error {
	updating = true

	// get the files
	max := 2000
	q := &storage.Query{
		Prefix: p.Prefix(),
	}

	logrus.Infof("fetching files from %s", p.BaseURL())
	files, err := p.List(p.Prefix(), p.Prefix(), "", max, q)
	if err != nil {
		return fmt.Errorf("Listing all files in bucket failed: %v", err)
	}

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

	// create temporoary file to save template to
	logrus.Info("creating temporary file for template")
	f, err := ioutil.TempFile("", "s3server")
	if err != nil {
		return fmt.Errorf("creating temp file failed: %v", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	// parse & execute the template
	logrus.Info("parsing and executing the template")
	templateDir := filepath.Join(staticDir, "../templates")
	lp := filepath.Join(templateDir, "layout.html")

	d := data{
		Files:       files,
		LastUpdated: time.Now().Local().Format(time.RFC1123),
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(lp))
	if err := tmpl.ExecuteTemplate(f, "layout", d); err != nil {
		return fmt.Errorf("execute template failed: %v", err)
	}
	f.Close()

	index := filepath.Join(staticDir, "index.html")
	logrus.Infof("renaming the temporary file %s to %s", f.Name(), index)
	if _, err := moveFile(index, f.Name()); err != nil {
		return fmt.Errorf("renaming result from %s to %s failed: %v", f.Name(), index, err)
	}
	updating = false
	return nil
}

func moveFile(dst, src string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()

	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()

	i, err := io.Copy(df, sf)
	if err != nil {
		return i, err
	}

	// Cleanup
	err = os.Remove(src)
	return i, err
}
