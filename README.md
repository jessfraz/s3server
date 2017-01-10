s3server
========

[![Travis CI](https://travis-ci.org/jessfraz/s3server.svg?branch=master)](https://travis-ci.org/jessfraz/s3server)

Static server for s3 or gcs files.

```console
$ s3server -h
Usage of ./s3server:
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
    r.j3ss.co/s3server -bucket s3://hugthief/gifs

# On Google Cloud Storage
$ docker run --restart always -d \
    --name gifs \
    -p 8080:8080 \
    -v ~/configs/path/config.json:/creds.json:ro \
    -e GOOGLE_APPLICATION_CREDENTIALS=/creds.json \
    r.j3ss.co/s3server -provider gcs -bucket gcs://misc.j3ss.co/gifs
```

![screenshot](screenshot.png)
