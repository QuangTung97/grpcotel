.PHONY:  generate install-tools

ROOT := ${PWD}

generate:
	mkdir -p rpc/backend
	cd proto && \
		protoc -I. \
			--go_out=paths=source_relative:${ROOT}/rpc/backend \
			--go-grpc_out=paths=source_relative:${ROOT}/rpc/backend \
			--grpc-gateway_out=logtostderr=true,paths=source_relative:${ROOT}/rpc/backend \
			backend.proto

install-tools:
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc