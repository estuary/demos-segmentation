.DEFAULT_GOAL := docker_build

registry = ghcr.io/estuary
image_name = demos-segmentation/event-generator
version = $(shell git describe --always --dirty --tags)

.PHONY: docker_build
docker_build:
	docker build -t $(registry)/$(image_name):$(version) -t $(registry)/$(image_name):latest .

.PHONY: docker_push
docker_push: docker_build
	docker push $(registry)/$(image_name):$(version)

.PHONY: docker_push_latest
docker_push_latest: docker_build
	docker push $(registry)/$(image_name):latest
