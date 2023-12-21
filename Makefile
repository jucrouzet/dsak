GO                 := $(shell command -v go 2> /dev/null)
GOLANGCILINT       := $(shell command -v golangci-lint 2> /dev/null)

VERSION            := $(shell grep "\[[0-9]\+\.[0-9]\+\.[0-9]\+\]" CHANGELOG.md | head -n 1 | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+')
BUILD              := $(shell git rev-parse HEAD)

DSAK               := build/dsak
LDFLAGS            := -X github.com/jucrouzet/dsak/internal/pkg/version.V=$(VERSION) \
                      -X github.com/jucrouzet/dsak/internal/pkg/version.Build=$(BUILD)

all: clean build-worker build-api

install-hooks: go-check
	@echo ">> Compiling commit-msg hook..."
	@go build -o .git/hooks/commit-msg .git-templates/commit-msg/main.go
	@echo ">> Installing pre-commit message hook..."
	@cp .git-templates/hooks/pre-commit .git/hooks/pre-commit

clean-bin:
	@echo ">> Cleaning old binaries"
	@rm -f $(DSAK)
	@rm -f $(DSAK)*

clean: clean-bin

go-check:
	@[ "${GO}" ] || ( echo ">> Go is not installed"; exit 1 )

linter-check: go-check
	@[ "${GOLANGCILINT}" ] || ( echo ">> Installing golangci-lint" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3 )

go-format: go-check
	@echo ">> formatting code"
	@go fmt ./...

go-vet: go-check
	@echo ">> vetting code"
	@go vet ./...

go-lint: go-format go-vet linter-check 
	@echo ">> linting code"
	@golangci-lint run ./...

build: go-check
	@echo ">> building binaries ..."
	@./crosscompile.sh


lint: go-lint

test: go-check
	@echo ">> Running tests"
	@go test -v ./...