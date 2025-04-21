.PHONY: build clean test run

# Binary name
BINARY_NAME=dai
# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

# Build
build:
	go build -o $(BINARY_NAME) -v

# Clean
clean:
	go clean
	rm -f $(BINARY_NAME)

# Test
test:
	go test ./...

# Run
run:
	go run main.go

# Install
install: build
	mv $(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)

# All
all: clean build 