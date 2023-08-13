DOCKER_USERNAME ?= ikirisfallen
APPLICATION_NAME ?= testserver-grpc
GIT_HASH ?= $(shell git log --format="%h" -n 1)
_BUILD_ARGS_TAG ?= $(GIT_HASH)
_BUILD_ARGS_RELEASE_TAG ?= latest
_BUILD_ARGS_DOCKERFILE ?= Dockerfile

gen protoc _protoc:
	shopt -s globstar; protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		./**/proto/*.proto

clean _clean:
	shopt -s globstar; rm -f ./**/*.pb.go

build _build:
	go build ./...

all _all:
	_protoc _build

_builder:
	docker build --tag $(DOCKER_USERNAME)/$(APPLICATION_NAME):$(_BUILD_ARGS_TAG) -f $(_BUILD_ARGS_DOCKERFILE) .
 
_pusher:
	docker push $(DOCKER_USERNAME)/$(APPLICATION_NAME):$(_BUILD_ARGS_TAG)

_releaser:
	docker pull $(DOCKER_USERNAME)/$(APPLICATION_NAME):$(_BUILD_ARGS_TAG)
	docker tag  $(DOCKER_USERNAME)/$(APPLICATION_NAME):$(_BUILD_ARGS_TAG) $(DOCKER_USERNAME)/$(APPLICATION_NAME):latest
	docker push $(DOCKER_USERNAME)/$(APPLICATION_NAME):$(_BUILD_ARGS_RELEASE_TAG)
 
docker_build:
	$(MAKE) _builder
 
docker_push:
	$(MAKE) _pusher

docker_release:
	$(MAKE) _releaser
