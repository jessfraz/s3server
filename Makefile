server:
	@go build ./...
	@docker build --rm --force-rm -t jess/s3server .
	@docker run --rm -it -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -p 8080:8080 jess/s3server -s3bucket s3://hugthief/gifs
