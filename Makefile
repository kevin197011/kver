# Copyright (c) 2025 kk
#
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

APP_NAME = kver
GO      ?= go

.PHONY: all build run test lint fmt clean cross-build

all: build

build:
	$(GO) build -o $(APP_NAME) main.go

run:
	$(GO) run main.go

test:
	$(GO) test ./...

lint:
	$(GO) install golang.org/x/lint/golint@latest
	golint ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -f $(APP_NAME)

cross-build:
	GOOS=darwin  GOARCH=amd64  $(GO) build -o $(APP_NAME)-darwin-amd64 main.go
	GOOS=darwin  GOARCH=arm64  $(GO) build -o $(APP_NAME)-darwin-arm64 main.go
	GOOS=linux   GOARCH=amd64  $(GO) build -o $(APP_NAME)-linux-amd64 main.go
	GOOS=linux   GOARCH=arm64  $(GO) build -o $(APP_NAME)-linux-arm64 main.go