.PHONY : default generate test-unit test-integration coverage coverage-report lint build run clean

SHELL=/bin/bash -e -o pipefail
PWD = $(shell pwd)

# constants
DOCKER_REPO = authz
DOCKER_TAG = latest

all: git-hooks generate tidy ## Initializes all tools

out:
	@mkdir -p out

generate: generate-protoc bin/mockery
	PATH=$(PWD)/bin:$$PATH go generate -tags integration ./...

test-unit: lint
	mkdir -p reports/coverage
	go test ./... -race -coverprofile=reports/coverage/coverage.out

test-integration:
	mkdir -p reports/coverage
	go test ./... -race -tags=integration -coverprofile=reports/coverage/coverage.out

download: ## Downloads the dependencies
	@go mod download

tidy: ## Cleans up go.mod and go.sum
	@go mod tidy

fmt: ## Formats all code with go fmt
	@go fmt ./...

test-build: ## Tests whether the code compiles
	@go build -o /dev/null ./...

build: out/bin ## Builds all binaries

GO_BUILD = mkdir -pv "$(@)" && go build -ldflags="-w -s" -o "$(@)" ./...
.PHONY: out/bin
out/bin:
	$(GO_BUILD)

test: ## Runs all tests
	@go test -tags integration $(ARGS) ./...

coverage: out/report.json ## Displays coverage per func on cli
	go tool cover -func=out/cover.out

html-coverage: out/report.json ## Displays the coverage results in the browser
	go tool cover -html=out/cover.out

test-reports: out/report.json

.PHONY: out/report.json
out/report.json: out
	@go test -tags integration -count 1 ./... -coverprofile=out/cover.out --json | tee "$(@)"

clean: ## Cleans up everything
	@rm -rf bin out protodeps

docker: ## Builds docker image
	docker buildx build -t $(DOCKER_REPO):$(DOCKER_TAG) .
# Go dependencies versioned through tools.go
GO_DEPENDENCIES = google.golang.org/protobuf/cmd/protoc-gen-go \
				google.golang.org/grpc/cmd/protoc-gen-go-grpc \
				github.com/envoyproxy/protoc-gen-validate \
				github.com/bufbuild/buf/cmd/buf \
                github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking \
                github.com/bufbuild/buf/cmd/protoc-gen-buf-lint \
                github.com/vektra/mockery/v2
# additional dependencies for grpc-gateway
GO_DEPENDENCIES += github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
				github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

define make-go-dependency
  # target template for go tools, can be referenced e.g. via /bin/<tool>
  bin/$(notdir $(subst /v2,,$1)): go.mod
	GOBIN=$(PWD)/bin go install $1
endef

# this creates a target for each go dependency to be referenced in other targets
$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make-go-dependency, $(dep))))

.PHONY: api/proto/buf.lock
api/proto/buf.lock: bin/buf
	@bin/buf mod update api/proto

protolint: api/proto/buf.lock bin/protoc-gen-buf-lint ## Lints your protobuf files
	bin/buf lint

protobreaking: api/proto/buf.lock bin/protoc-gen-buf-breaking ## Compares your current protobuf with the version on master to find breaking changes
	bin/buf breaking --against '.git#branch=main'

generate-protoc: ## Generates code from protobuf files
generate-protoc: bin/protoc-gen-grpc-gateway bin/protoc-gen-openapiv2 api/proto/buf.lock bin/protoc-gen-go bin/protoc-gen-go-grpc bin/protoc-gen-validate
	PATH=$(PWD)/bin:$$PATH buf generate

ci: lint-reports test-reports ## Executes lint and test and generates reports

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''
