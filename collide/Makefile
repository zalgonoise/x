app:
	GOOS=linux go build -ldflags="-s -w" -o "out/app" ./cmd/collide
	docker compose -f docker-compose.yaml build

app-fe:
	cp api/openapi/collide/v1/collide_api.swagger.json frontend/static/collide_api.swagger.json
	GOOS=linux go build -ldflags="-s -w" -o "out/app" ./cmd/collide-fe
	docker compose -f docker-compose.frontend.yaml build collidefe

app-localfe:
	cp api/openapi/collide/v1/collide_api.swagger.json frontend/static/collide_api.swagger.json
	GOOS=linux go build -ldflags="-s -w" -o "out/app" ./cmd/collide-fe
	docker compose -f docker-compose.frontend.yaml build collidefe-local

up: app
	docker compose -f docker-compose.yaml up -d --force-recreate

up-fe: app-fe app-localfe
	docker compose -f docker-compose.frontend.yaml up -d --force-recreate

up-observabilitiy:
	docker compose -f docker-compose.observability.yaml up -d --force-recreate

up-all: up-observabilitiy up up-fe

down:
	docker compose -f docker-compose.yaml down

down-fe:
	docker compose -f docker-compose.frontend.yaml down

down-observabilitiy:
	docker compose -f docker-compose.observability.yaml down

down-all: down-fe down down-observabilitiy

clean-observability:
	docker compose -f docker-compose.observability.yaml down -v

clean-all: down-fe down clean-observability

dummy-load:
	go run ./scripts/dummy load

new: clean-all up-all dummy-load

generate:
	go generate ./...
	rm -rf ./diagrams && \
		mkdir -p ./diagrams && \
		go run ./scripts/diagrams generate && \
		docker run --rm -ti -v ./diagrams:/diagrams -w /diagrams docker.io/zalgonoise/graphviz:ubuntu sh -c \
			'dot -Tpng collide_api.dot > collide_api.png'

.PHONY: dep-update
dep-update:
	go get -u ./...
	go mod tidy

##

GO_DEPENDENCIES = google.golang.org/protobuf/cmd/protoc-gen-go \
				google.golang.org/grpc/cmd/protoc-gen-go-grpc \
				github.com/envoyproxy/protoc-gen-validate \
				github.com/bufbuild/buf/cmd/buf \
				github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking \
				github.com/bufbuild/buf/cmd/protoc-gen-buf-lint \
				github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
					github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2

define make-go-dependency
	bin/$(notdir $(subst /v2,,$1)): go.mod
		GOBIN=$(PWD)/bin go install $1
endef

$(foreach dep, $(GO_DEPENDENCIES), $(eval $(call make_go_dependency, $(dep))))

.PHONY: api/proto/buf.lock
api/proto/buf.lock: bin/buf
	@bin/buf mod update api/proto

protolint: api/prot/buf.lock bin/protoc-gen-buf-lint
	bin/buf lint

protobreaking: api/prot/buf.lock bin/protoc-gen-buf-breaking
	bin/buf breaking --against '.git#branch=master'

generate-protoc:
generate-protoc: bin/protoc-gen-grpc-gateway bin/protoc-gen-openapiv2 api/proto/buf.lock bin/protoc-gen-go bin/protoc-gen-go-grpc bin/protoc-gen-validate
	PATH=$(PWD)/bin:$$PATH buf generate

.PHONY: update-proto
update-proto:
	mkdir -p ./api/proto/validate
	curl -sSL --url 'https://github.com/bufbuild/protoc-gen-validate/raw/main/validate/validate.proto' > './api/proto/validate/validate.proto'
