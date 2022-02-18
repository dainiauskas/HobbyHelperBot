VERSION := $(shell git describe --abbrev=0 --tags)

BUILD := $(shell git rev-parse --short HEAD)
BUILD_NAME := $(shell basename "$(PWD)")

GOBASE 	:= $(shell pwd)
GOBIN 	:= $(GOBASE)/bin

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-s -X main.Version=$(VERSION) -X main.Build=$(BUILD) -X main.Name=$(BUILD_NAME)"

build:
	GOOS=linux CGO_ENABLED=0 go build -tags netgo -a $(LDFLAGS) -o $(GOBIN)/${BUILD_NAME}

push-to-server: build
	rsync -r -a -v -e ssh $(GOBIN)/${BUILD_NAME} root@fpv-coffee.eu:/opt/hobbyHelper/${BUILD_NAME}
	ssh root@fpv-coffee.eu systemctl restart HobbyHelperBot