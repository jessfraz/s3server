FROM debian:jessie
MAINTAINER Jessica Frazelle <jess@docker.com>

RUN apt-get update && apt-get install -y \
	ca-certificates \
	--no-install-recommends \
	&& rm -rf /var/lib/apt/lists/*

ADD https://jesss.s3.amazonaws.com/binaries/s3server /usr/local/bin/s3server

COPY static /src/static
COPY templates /src/templates

RUN chmod +x /usr/local/bin/s3server

ENTRYPOINT [ "/usr/local/bin/s3server" ]
