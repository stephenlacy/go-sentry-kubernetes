VERSION=$(shell git describe --always --long)
GOBUILD=go build .
DOCKER_USER=stevelacy
NAME=go-sentry-kubernetes
IMAGE=$(DOCKER_USER)/$(NAME):$(VERSION)

all: docker

build:
	$(GOBUILD)

build_linux:
	GOOS=linux GOARCH=386 CGO_ENABLED=0 $(GOBUILD)

docker:
	docker build -t $(IMAGE) .

push:
	docker push $(IMAGE)

clean:
	rm -f $(NAME)
