PROJECT_ROOT:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist

.PHONY: all
all: test

.PHONY: clean
clean::
	-rm -rf ${DIST_DIR}/

.PHONY: test
test:
	go test -v -race -short -tags no_e2e $(shell go list ./... | grep -v /vendor/) 

.PHONY: lint
lint:
	golangci-lint run --config golangci.yml

.PHONY: run-example
run-example:
	go run $(shell go list ./example/...) --config example/config

.PHONY: e2e
e2e:
	@go test -v ${PROJECT_ROOT}/test/e2e/e2e_test.go

.PHONY: coverage
coverage:
	go test -covermode=count -coverprofile=profile.cov $(shell go list ./... | grep -v /vendor/)
	go tool cover -func=profile.cov

.PHONY: check-license
check-license:
	./scripts/check_license.sh
