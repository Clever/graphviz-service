SHELL := /bin/bash
PKG := github.com/Clever/aws-cost-notifier/cmd
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := graphviz-service
.PHONY: test build vendor clean all run $(PKGS)

GOVERSION := $(shell go version | grep 1.5)
ifeq "$(GOVERSION)" ""
  $(error must be running Go version 1.5)
endif
export GO15VENDOREXPERIMENT=1

all: test build

build:
	# Disable CGO and link completely statically so executable runs in contains based on alpine.
	# i.e. containers that don't use glibc.  Also hardcode OS and architecture to make double sure.
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./$(EXECUTABLE) -a -installsuffix cgo

test: $(PKGS)

GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go get github.com/golang/lint/golint

GODEP := $(GOPATH)/bin/godep
$(GODEP):
	go get -u github.com/tools/godep

$(PKGS): $(GOLINT)
	@echo ""
	@echo "FORMATTING $@..."
	gofmt -w=true $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "LINTING $@..."
	$(GOLINT) $(GOPATH)/src/$@/*.go
	@echo ""
	@echo "TESTING $@..."
	go test -v $@

clean:
	rm bin/*

vendor: $(GODEP)
	$(GODEP) save $(PKGS)
	find vendor/ -path '*/vendor' -type d | xargs -IX rm -r X # remove any nested vendor directories

run: build
	sudo docker build -t graphviz-service .
	sudo docker run -p :8081:80 -e "PORT=80" graphviz-service
