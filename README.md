s3server
========

[![Travis CI](https://travis-ci.org/jessfraz/s3server.svg?branch=master)](https://travis-ci.org/jessfraz/s3server)

Static server for s3 or gcs files.

## Installation

#### Binaries

- **darwin** [386](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-darwin-386) / [amd64](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-darwin-amd64)
- **freebsd** [386](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-freebsd-386) / [amd64](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-freebsd-amd64)
- **linux** [386](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-linux-386) / [amd64](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-linux-amd64) / [arm](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-linux-arm) / [arm64](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-linux-arm64)
- **solaris** [amd64](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-solaris-amd64)
- **windows** [386](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-windows-386) / [amd64](https://github.com/jessfraz/s3server/releases/download/v0.1.0/s3server-windows-amd64)

#### Via Go

```bash
$ go get github.com/jessfraz/s3server
```

## Usage

```console
$ s3server -h
     _        _   _
 ___| |_ __ _| |_(_) ___ ___  ___ _ ____   _____ _ __
/ __| __/ _` | __| |/ __/ __|/ _ \ '__\ \ / / _ \ '__|
\__ \ || (_| | |_| | (__\__ \  __/ |   \ V /  __/ |
|___/\__\__,_|\__|_|\___|___/\___|_|    \_/ \___|_|

 Server to index & view files in a s3 or Google Cloud Storage bucket.
 Version: v0.1.0
 Build: e5a60e2

  -bucket string
        bucket path from which to serve files
  -cert string
        path to ssl certificate
  -interval string
        interval to generate new index.html's at (default "5m")
  -key string
        path to ssl key
  -p string
        port for server to run on (default "8080")
  -provider string
        cloud provider (ex. s3, gcs) (default "s3")
  -s3key string
        s3 access key
  -s3region string
        aws region for the bucket (default "us-west-2")
  -s3secret string
        s3 access secret
  -v    print version and exit (shorthand)
  -version
        print version and exit
```

**run with the docker image**

```console
# On AWS S3
$ docker run -d \
    --restart always \
    -e AWS_ACCESS_KEY_ID \
    -e AWS_SECRET_ACCESS_KEY \
    -p 8080:8080 \
    --name s3server \
    --tmpfs /tmp \
    r.j3ss.co/s3server -bucket s3://hugthief/gifs

# On Google Cloud Storage
$ docker run --restart always -d \
    --name gifs \
    -p 8080:8080 \
    -v ~/configs/path/config.json:/creds.json:ro \
    -e GOOGLE_APPLICATION_CREDENTIALS=/creds.json \
    --tmpfs /tmp \
    r.j3ss.co/s3server -provider gcs -bucket gcs://misc.j3ss.co/gifs
```

![screenshot](screenshot.png)
