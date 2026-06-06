BINARY_DIR := ./bin/
BINARY_NAME := happahappa
BINARY_X86_64 := $(BINARY_DIR)$(BINARY_NAME)
BINARY_ARM_32 := $(BINARY_DIR)$(BINARY_NAME)-arm32

.PHONY: default test build run clean dirs

default: test build build-arm

dirs:
	mkdir -p $(BINARY_DIR)

test:
	go test ./... -v 4

build: dirs
	go build -o $(BINARY_X86_64) .

build-arm: dirs
	GOOS=linux GOARCH=arm go build -o $(BINARY_ARM_32) .

run: build
	$(BINARY)

clean:
	rm -rf $(BINARY_DIR)
