.DEFAULT_GOAL=help

.PHONY: all clean fmt lint vet build ls_bins ls_srcs imports

BINS := cigarmender
SRCS := $(wildcard  ./cmd/cigarmender/*.go)
BIN_DIR := ./bin
BASH_BINS := $(BIN_DIR)/$(BINS)
WIN_BINS := $(BIN_DIR)/$(BINS).exe
GOPATH := $(shell go env GOPATH)

all: fmt imports vet lint test build ## Lint, test and build binaries

build: vet lint ## Build binaries with go build
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

lint: fmt ## Static analysis with golangci-lint
	golangci-lint run ./...

imports: ## Organise imports with goimports
	goimports -l -w .

clean: ls_bins ## Remove binaries
	@echo "Cleaning..."
	rm -rvf ${BASH_BINS}
	rm -rvf ${WIN_BINS}
	@echo

test: fmt imports vet lint ## Run all tests
	gotest -v ./...

gh_token:
	@echo "Exporting github token"
	export GITHUB_TOKEN="$(< ~/.secrets/goreleaser.ghtoken)"

release: gh_token ## Release the latest tagged version with goreleaser
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

help: ## Display this help message
	@echo "CIGARMender - homopolymer-aware deletion centering of aligned reads"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
