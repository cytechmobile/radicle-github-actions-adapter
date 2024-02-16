# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go get -u ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...

## lint: run lint control checks
.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	CGO_ENABLED=1 go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	CGO_ENABLED=1 go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build: build the cmd/rest application
.PHONY: build
build:
	go build -ldflags "-X 'radicle-github-actions-adapter/pkg/version.Version=development' -X 'radicle-github-actions-adapter/pkg/version.BuildTime=$(shell date)'" -o=/tmp/bin/radicle-github-actions-adapter ./cmd/github-actions-adapter

## run: run the cmd/rest application
.PHONY: run
run: build
	/tmp/bin/radicle-github-actions-adapter
