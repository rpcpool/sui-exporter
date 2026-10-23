# Sui gRPC protos are vendored from MystenLabs/sui-apis at SUI_APIS_REF.
SUI_APIS_REF := aa9ba2c13b2a9348059b38110e64af39849a0377
GO_PKG := github.com/rpcpool/sui-exporter/internal/suirpc

PROTOS := $(sort $(wildcard proto/sui/rpc/v2/*.proto))
M_OPTS := $(foreach p,$(PROTOS),--go_opt=M$(p:proto/%=%)=$(GO_PKG) --go-grpc_opt=M$(p:proto/%=%)=$(GO_PKG)) \
	--go_opt=Mgoogle/rpc/status.proto=google.golang.org/genproto/googleapis/rpc/status \
	--go-grpc_opt=Mgoogle/rpc/status.proto=google.golang.org/genproto/googleapis/rpc/status

.PHONY: build test proto proto-fetch

build:
	go build -o sui-exporter ./cmd/sui-exporter

test:
	go test ./...

# Regenerates internal/suirpc. Needs protoc, protoc-gen-go and protoc-gen-go-grpc on PATH.
proto:
	rm -f internal/suirpc/*.pb.go
	protoc -I proto \
		--go_out=internal/suirpc --go_opt=paths=source_relative \
		--go-grpc_out=internal/suirpc --go-grpc_opt=paths=source_relative \
		$(M_OPTS) $(PROTOS:proto/%=%)
	mv internal/suirpc/sui/rpc/v2/*.go internal/suirpc/
	rm -rf internal/suirpc/sui

# Replaces the vendored protos with the ones at SUI_APIS_REF.
proto-fetch:
	rm -rf proto/sui proto/google
	mkdir -p proto/google/rpc
	curl -sSL https://codeload.github.com/MystenLabs/sui-apis/tar.gz/$(SUI_APIS_REF) \
		| tar xz --strip-components=1 -C proto --wildcards '*/proto/sui/rpc/v2/*.proto' '*/proto/google/rpc/status.proto'
	mv proto/proto/sui proto/sui && mv proto/proto/google/rpc/status.proto proto/google/rpc/ && rm -rf proto/proto
