# s3server

[![Travis CI](https://img.shields.io/travis/jessfraz/s3server.svg?style=for-the-badge)](https://travis-ci.org/jessfraz/s3server)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/s3server)

Static server for s3 or gcs files.

 * [Installation](README.md#installation)
      * [Binaries](README.md#binaries)
      * [Via Go](README.md#via-go)
 * [Usage](README.md#usage)

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/jessfraz/s3server/releases).

#### Via Go

```console
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
 Version: v0.2.2
 Build: 0ea0e32

  -bucket string
        bucket path from which to serve files
  -cert string
        path to ssl certificate
  -interval duration
        interval to generate new index.html's at (default 5m0s)
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
