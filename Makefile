
VERSION := $(shell git describe --tags || git rev-parse --short HEAD)
IMG ?= xmapst/otter-alert

.PHONY: build docker
build:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/otter-alert cmd/main.go

docker:
	docker build --rm --network host -t $(IMG):$(VERSION) .
	docker push $(IMG):$(VERSION)
	docker rmi $(IMG):$(VERSION)