BINARY_DIR := ./bin/

HAPPA_NAME := happahappa
TOKEN_NAME := tokengen
COMMAND_NAME := command

HAPPA_X86_64 := $(BINARY_DIR)$(HAPPA_NAME)
HAPPA_ARM_32 := $(BINARY_DIR)$(HAPPA_NAME)-arm32

TOKEN_X86_64 := $(BINARY_DIR)$(TOKEN_NAME)
TOKEN_ARM_32 := $(BINARY_DIR)$(TOKEN_NAME)-arm32

COMMAND_X86_64 := $(BINARY_DIR)$(COMMAND_NAME)
COMMAND_ARM_32 := $(BINARY_DIR)$(COMMAND_NAME)-arm32

.PHONY: default test build build-arm clean dirs

default: test build build-arm

dirs:
	mkdir -p $(BINARY_DIR)

test:
	go test ./... -v


build: build-happa build-tokengen build-command

build-happa: dirs
	go build -o $(HAPPA_X86_64) ./cmd/happahappa/happahappa

build-tokengen: dirs
	go build -o $(TOKEN_X86_64) ./cmd/happahappa/token

build-command: dirs
	go build -o $(COMMAND_X86_64) ./cmd/happahappa/command

build-arm: build-happa-arm build-tokengen-arm build-command-arm

build-happa-arm: dirs
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(HAPPA_ARM_32) ./cmd/happahappa/happahappa

build-tokengen-arm: dirs
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(TOKEN_ARM_32) ./cmd/happahappa/token

build-command-arm: dirs
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(COMMAND_ARM_32) ./cmd/happahappa/command

clean:
	rm -rf $(BINARY_DIR)
