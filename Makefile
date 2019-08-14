# Image URL to use all building/pushing image targets
IMG ?= aflc/extended-resource-toleration-webhook
TAG ?= 0.1
CERT_DAYS ?= 365
KS_NAMESPACE ?= default
KS_ARG ?= ""

build:
	go build -gcflags 'all=-N -l' -o bin/ert-webhook .

build-docker:
	docker build -t $(IMG):$(TAG) .
	@echo Built $(IMG):$(TAG)

apply-k8s:
	CERT_DAYS=$(CERT_DAYS) KS_NAMESPACE=$(KS_NAMESPACE) ./apply-k8s.sh $(KS_ARG)