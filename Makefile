.PHONY: lint test vendor clean copy_src

default: lint test

lint:
	golangci-lint run -v

test:
	go test -v -cover ./...

build:
	go build -v -o traefik-aws-plugin

vendor:
	go mod vendor

clean:
	rm -rf ./vendor

copy_src:
	mkdir -p go/src/github.com/bluecatengineering/traefik-aws-plugin
	cp -r ecs local log s3 signer .traefik.yml go.mod Makefile aws.go aws_test.go go/src/github.com/bluecatengineering/traefik-aws-plugin/
