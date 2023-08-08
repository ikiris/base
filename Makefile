gen protoc _protoc:
	shopt -s globstar; protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		./**/proto/*.proto

clean _clean:
	shopt -s globstar; rm -f ./**/*.pb.go

_build:
	go build ./...

all _all:
	_protoc _build