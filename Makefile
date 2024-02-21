NAME 			      := mintscan-union
VERSION               := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT                := $(shell git log -1 --format='%H')
SHORT_COMMIT 		  := $(shell git rev-parse --short=10 HEAD)
DESTDIR         	  ?= $(GOPATH)/bin/${NAME}
BUILD_FLAGS 		  := -ldflags "-w -s \
	-X github.com/cosmostation/cosmostation-coreum/chain-exporter/exporter.Version=${VERSION} \
	-X github.com/cosmostation/cosmostation-coreum/chain-exporter/exporter.Commit=${COMMIT}"

## Show all make target commands.
help:
	@make2help $(MAKEFILE_LIST)

## Print out application information.
version: 
	@echo "NAME: ${NAME}"
	@echo "VERSION: ${VERSION}"
	@echo "COMMIT: ${COMMIT}"

build: build_exporter build_mintscan

## build chain-exporter
build_exporter: go.sum
	@echo "-> Building chain-exporter"
	@go build -mod=readonly $(BUILD_FLAGS) -o build/chain-exporter-${SHORT_COMMIT} ./cmd/chain-exporter

## build mintscan(es-handler)
build_mintscan: go.sum
	@echo "-> Building mintscan"
	@go build -mod=readonly $(BUILD_FLAGS) -o ./build/mintscan-${SHORT_COMMIT} ./cmd/mintscan

install: install_exporter install_mintscan

## build chain-exporter and move to GOBIN
install_exporter: go.sum
	@echo "-> Installing chain-exporter"
	@go build -mod=readonly $(BUILD_FLAGS) -o build/chain-exporter-${SHORT_COMMIT} ./cmd/chain-exporter
	@mv ./build/* ${shell go env GOBIN}/

## build mintscan and move to GOBIN
install_mintscan: go.sum
	@echo "-> Installing mintscan"
	@go build -mod=readonly $(BUILD_FLAGS) -o ./build/mintscan-${SHORT_COMMIT} ./cmd/mintscan
	@mv ./build/* ${shell go env GOBIN}/

## Clean build directory
clean:
	@echo "-> Cleaning ./build/*"
	@rm -rf build/*

## Clean binary at GOBIN
clean-install:
	@echo "-> Cleaning your GOBIN"
	@rm ${shell go env GOBIN}/chain-exporter*
	@rm ${shell go env GOBIN}/mintscan*