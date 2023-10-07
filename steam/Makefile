# Go dependencies versioned through tools.go
GO_DEPENDENCIES = google.golang.org/protobuf/cmd/protoc-gen-go \
				google.golang.org/grpc/cmd/protoc-gen-go-grpc \
				github.com/envoyproxy/protoc-gen-validate \
				github.com/bufbuild/buf/cmd/buf \
                github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking \
                github.com/bufbuild/buf/cmd/protoc-gen-buf-lint \
                github.com/vektra/mockery/v2

define make-go-dependency
  # target template for go tools, can be referenced e.g. via /bin/<tool>
  bin/$(notdir $(subst /v2,,$1)): go.mod
	GOBIN=$(PWD)/bin go install $1@latest
endef


# this creates a target for each go dependency to be referenced in other targets
$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make-go-dependency, $(dep))))

.PHONY: proto/buf.lock
proto/buf.lock: bin/buf
	@bin/buf mod update proto

protolint: proto/buf.lock bin/protoc-gen-buf-lint ## Lints your protobuf files
	bin/buf lint

protobreaking: proto/buf.lock bin/protoc-gen-buf-breaking ## Compares your current protobuf with the version on master to find breaking changes
	bin/buf breaking --against '.git#branch=main'

generate-protoc: ## Generates code from protobuf files
generate-protoc:  proto/buf.lock bin/protoc-gen-go bin/protoc-gen-go-grpc bin/protoc-gen-validate
	PATH=$(PWD)/bin:$$PATH buf generate