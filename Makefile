PWD:=$(shell pwd)

test:
	@go run main.go -o stdout https://news.ycombinator.com/

fmt:
	@go fmt main.go

build:
	@go build .

lint:
	@golangci-lint run --out-format tab
