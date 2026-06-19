.DEFAULT_GOAL=build

.PHONY: all clean fmt lint vet build ls_bins ls_srcs imports

SRCS := $(wildcard *.go)
BASH_BINS := $(SRCS:%.go=%)
WIN_BINS := $(SRCS:%.go=%.exe)

all: fmt imports vet lint build 

build: vet lint
	@echo "Go building (${SRCS})"
	go build ${SRCS}
	@echo

vet: fmt
	go vet ./...

fmt:
	go fmt ./...

lint: fmt
	golint ./...

imports:
	goimports -l -w .

clean: ls_bins
	@echo "Cleaning..."
	rm -rvf ${BASH_BINS}
	rm -rvf ${WIN_BINS}
	@echo

win-tools: tools-golint tools-goimports win-tools-golangci-lint
	
win-tools-golangci-lint:
	@echo "Installing golangci-lint"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1
	@echo

tools-golint:
	@echo "Installing golint"
	go install golang.org/x/lint/golint
	@echo

tools-goimports:
	@echo "Installing goimports"
	go install golang.org/x/tools/cmd/goimports
	@echo

ls_bins: 
	@echo "${BASH_BINS}"
	@echo "${WIN_BINS}"

ls_srcs: 
	@echo "${SRCS}"
