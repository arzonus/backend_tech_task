export CGO_ENABLED=0
export GO111MODULE=on

build-api:
	go build \
        -a \
        -installsuffix cgo \
        -tags netgo \
        -o ./bin/api \
            cmd/api/main.go

build-external:
	go build \
        -a \
        -installsuffix cgo \
        -tags netgo \
        -o ./bin/external \
            cmd/external/main.go

test:
	export CGO_ENABLED=0
	go test -v $$(go list ./... | grep -v /tests/ ) -coverprofile=coverage.out && \
	go tool cover -func=coverage.out

lint:
	golangci-lint run -v

generate-grpc:
	protoc \
      -I proto \
      -I ${GOPATH}/src \
      -I ${GOPATH}/src/github.com/envoyproxy/protoc-gen-validate \
      --go_out="plugins=grpc:pkg/proto" \
      --validate_out="lang=go:pkg/proto" \
      proto/*.proto

run-dev-env:
	docker-compose \
		-f deployments/docker-compose.dev.yaml \
		up -d --build

stop-dev-env:
	docker-compose \
		-f deployments/docker-compose.dev.yaml \
		rm -f

rerun-dev-env: stop-dev-env run-dev-env

run-local-env:
	docker-compose \
		-f deployments/docker-compose.dev.yaml \
		-f deployments/docker-compose.local.yaml \
		up -d --build --remove-orphans

stop-local-env:
	docker-compose \
		-f deployments/docker-compose.dev.yaml \
		-f deployments/docker-compose.local.yaml \
		rm -f

rerun-local-env: stop-local-env run-local-env