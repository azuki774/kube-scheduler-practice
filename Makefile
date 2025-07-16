SHELL=/bin/bash
IMAGE_NAME:=kube-scheduler-practice:latest
.PHONY: bin
bin:
	go build -a -tags "netgo" -installsuffix netgo  -ldflags="-s -w -extldflags \"-static\" \
	-X main.version=$(git describe --tag --abbrev=0) \
	-X main.revision=$(git rev-list -1 HEAD) \
	-X main.build=$(git describe --tags)" \
	-o bin/ ./...

.PHONY: bin-docker
bin-docker:
	go build -a -tags "netgo" -installsuffix netgo  -ldflags="-s -w -extldflags \"-static\" \
	-X main.version=$(git describe --tag --abbrev=0) \
	-X main.revision=$(git rev-list -1 HEAD) \
	-X main.build=$(git describe --tags)" \
	-o /app/ ./...

.PHONY: build
build:
	docker build -t $(IMAGE_NAME) .

.PHONY: kind-load
kind-load:
	kind load docker-image $(IMAGE_NAME)

.PHONY: test
test:
	go test -v ./...
