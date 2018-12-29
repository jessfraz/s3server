# s3server

[![Travis CI](https://img.shields.io/travis/jessfraz/s3server.svg?style=for-the-badge)](https://travis-ci.org/jessfraz/s3server)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/s3server)

Static server for s3 or gcs files.

**Table of Contents**

<!-- toc -->

<!-- tocstop -->

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
s3server -  Server to index & view files in a s3 or Google Cloud Storage bucket.

Usage: s3server <command>

Flags:

  --bucket    bucket path from which to serve files (default: <none>)
  --cert      path to ssl certificate (default: <none>)
  -d          enable debug logging (default: false)
  --interval  interval to generate new index.html's at (default: 5m0s)
  --key       path to ssl key (default: <none>)
  -p          port for server to run on (default: 8080)
  --provider  cloud provider (ex. s3, gcs) (default: s3)
  --s3key     s3 access key (default: <none>)
  --s3region  aws region for the bucket (default: us-west-2)
  --s3secret  s3 access secret (default: <none>)

Commands:

  version  Show the version information.
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
