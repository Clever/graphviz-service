include golang.mk
.DEFAULT_GOAL := test

.PHONY: all test build vendor clean run $(PKGS)
SHELL := /bin/bash
PKG := github.com/Clever/aws-cost-notifier/cmd
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := graphviz-service
$(eval $(call golang-version-check,1.24))

all: test build

build:
	# Disable CGO and link completely statically so executable runs in contains based on alpine.
	# i.e. containers that don't use glibc.  Also hardcode OS and architecture to make double sure.
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./$(EXECUTABLE) -a -installsuffix cgo

test: $(PKGS)

$(PKGS): golang-test-all-deps
				$(call golang-fmt,$@)
				$(call goalng-lint,$@)
				$(call golang-vet,$@)

clean:
	rm bin/*

run: build
	docker build -t graphviz-service .
	docker run -p :5001:80 -e "PORT=80" graphviz-service



install_deps:
	go mod vendor
