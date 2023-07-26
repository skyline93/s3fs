# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=s3fs
BINARY_UNIX=$(BINARY_NAME)_unix

all: build
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	$(GOGET) github.com/aws/aws-sdk-go/aws
	$(GOGET) github.com/aws/aws-sdk-go/aws/credentials
	$(GOGET) github.com/aws/aws-sdk-go/aws/session
	$(GOGET) github.com/aws/aws-sdk-go/service/s3
	$(GOGET) github.com/sevlyar/go-daemon
	$(GOGET) github.com/spf13/cobra
	$(GOGET) bazil.org/fuse
