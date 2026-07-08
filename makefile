.DEFAULT_GOAL=build

.PHONY: all clean fmt lint vet build ls_bins ls_srcs imports

BINS := cigarmender
SRCS := $(wildcard  ./cmd/cigarmender/*.go)
BIN_DIR := ./bin
BASH_BINS := $(BIN_DIR)/$(BINS)
WIN_BINS := $(BIN_DIR)/$(BINS).exe
GOPATH := $(shell go env GOPATH)

all: fmt imports vet lint test build 

build: vet lint
	@echo "Go building (${SRCS})"
	mkdir -p $(BIN_DIR)
	go build -o $(BASH_BINS) $(SRCS)
	@echo

win-build: vet lint
	@echo "Go building (${SRCS})"
	mkdir -p $(BIN_DIR)
	go build -o $(WIN_BINS) $(SRCS)
	@echo

vet: fmt
	go vet ./...

fmt:
	go fmt ./...

lint: fmt
	golangci-lint run ./...

imports:
	goimports -l -w .

clean: ls_bins
	@echo "Cleaning..."
	rm -rvf ${BASH_BINS}
	rm -rvf ${WIN_BINS}
	@echo

test: fmt imports vet lint
	gotest -v ./...

gh_token:
	@echo "Exporting github token"
	export GITHUB_TOKEN="$(< ~/.secrets/goreleaser.ghtoken)"

release: gh_token
	@echo "Releasing..."
	goreleaser release --clean
	@echo

win-tools: tools-golint tools-goimports win-tools-golangci-lint
	
win-tools-golangci-lint:
	@echo "Installing golangci-lint"
	choco install golangci-lint
	@echo

tools-golint:
	@echo "Installing golangci-lint"
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(GOPATH)/bin v2.12.2
	@echo

tools-goimports:
	@echo "Installing goimports"
	go install golang.org/x/tools/cmd/goimports@latest
	@echo

ls_bins: 
	@echo "${BASH_BINS}"
	@echo "${WIN_BINS}"

ls_srcs: 
	@echo "${SRCS}"
