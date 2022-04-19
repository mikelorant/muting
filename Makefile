SHELL = /bin/bash

REPO = mikelorant
NAME = muting
NAMESPACE = default
TAG = $(shell git describe --abbrev=0 --tags | sed 's/^v//g')
KUBECTL = kubectl
HELMFILE = helmfile

.PHONY: test
test:
	go test -v ./...

.PHONY: image
image: test
	docker build \
		-t ${REPO}/${NAME}:${TAG} \
		-t ${REPO}/${NAME}:latest \
		.

.PHONY: release
release: image
	docker push ${REPO}/${NAME}:${TAG}
	docker push ${REPO}/${NAME}:latest

.PHONY: apply
apply:
	${KUBECTL} create namespace ${NAMESPACE} --dry-run=client -o yaml | ${KUBECTL} apply -f -
	${HELMFILE} --namespace ${NAMESPACE} apply

.PHONY: example
example:
	${KUBECTL} label namespace ${NAMESPACE} muting=enabled --overwrite=true
	${KUBECTL} apply --namespace ${NAMESPACE} --filename example
