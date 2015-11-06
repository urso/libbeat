#/bin/bash

### VARIABLE SETUP ###

GLIDE=GO15VENDOREXPERIMENT=1 $(GOPATH)/bin/glide
GO=GO15VENDOREXPERIMENT=1 go
GOGET=go get
GOFMT=gofmt
GOTESTCOVER=GO15VENDOREXPERIMENT=1 $(GOPATH)/bin/gotestcover

MODULES=$$($(GLIDE) novendor)

# Hidden directory to install dependencies for jenkins
export PATH := ./bin:$(PATH)
GOFILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')
SHELL=/bin/bash
ES_HOST?="elasticsearch-200"
BUILD_DIR=build
COVERAGE_DIR=${BUILD_DIR}/coverage
PROCESSES?= 4
TIMEOUT?= 90


### BUILDING ###

# Builds libbeat. No binary created as it is a library
.PHONY: build
build: vendortool deps
	$(GO) build $(MODULES)

# Cross-compile libbeat for the OS and architectures listed in
# crosscompile.bash. The binaries are placed in the ./bin dir.
.PHONY: crosscompile
crosscompile: vendortool $(GOFILES)
	mkdir -p ${BUILD_DIR}/bin
	source scripts/crosscompile.bash; OUT='${BUILD_DIR}/bin' go-build-all

.PHONY: vendortool
vendortool:
	# first make sure we have glide
	$(GOGET) github.com/Masterminds/glide


# Fetch dependencies
deps: vendor

vendor: glide.yaml
	# install dependencies
	$(GLIDE) install

# Checks project and source code if everything is according to standard
.PHONY: check
check:
	# This should be modified so it throws an error on the build system in case the output is not empty
	$(GOFMT) -d $(GOFILES)
	$(GO) vet $(MODULES)

# Cleans up directory and source code with gofmt
.PHONY: clean
clean:
	$(GOFMT) $(GOFILES)
	-rm -r build

# Shortcut for continuous integration
# This should always run before merging.
.PHONY: ci
ci:
	make
	make check
	make testsuite

### Testing ###
# All tests are always run with coverage reporting enabled


# Prepration for tests
.PHONY: prepare-tests
prepare-tests:
	mkdir -p ${COVERAGE_DIR}
	# coverage tools
	$(GOGET) golang.org/x/tools/cmd/cover
	# gotestcover is needed to fetch coverage for multiple packages
	$(GOGET) github.com/pierrre/gotestcover

# Runs the unit tests
.PHONY: unit-tests
unit-tests: prepare-tests
	#go test -short ./...
	$(GOTESTCOVER) -coverprofile=${COVERAGE_DIR}/unit.cov -short -covermode=count $(MODULES)

# Run integration tests. Unit tests are run as part of the integration tests
.PHONY: integration-tests
integration-tests: prepare-tests
	$(GOTESTCOVER) -coverprofile=${COVERAGE_DIR}/integration.cov -covermode=count $(MODULES)

# Runs the integration inside a virtual environment. This can be run on any docker-machine (local, remote)
.PHONY: integration-tests-environment
integration-tests-environment:
	make prepare-tests
	make build-image
	NAME=$$(docker-compose run -d libbeat make integration-tests | awk 'END{print}') || exit 1; \
	echo "docker libbeat test container: '$$NAME'"; \
	docker attach $$NAME; CODE=$$?;\
	mkdir -p ${COVERAGE_DIR}; \
	docker cp $$NAME:/go/src/github.com/elastic/libbeat/${COVERAGE_DIR}/integration.cov $(shell pwd)/${COVERAGE_DIR}/; \
	docker rm $$NAME > /dev/null; \
	exit $$CODE

# Runs the system tests
.PHONY: system-tests
system-tests: libbeat.test prepare-tests system-tests-setup
	. build/system-tests/env/bin/activate; nosetests -w tests/system --processes=${PROCESSES} --process-timeout=$(TIMEOUT)
	# Writes count mode on top of file
	echo 'mode: count' > ${COVERAGE_DIR}/system.cov
	# Collects all system coverage files and skips top line with mode
	tail -q -n +2 ./build/system-tests/run/**/*.cov >> ${COVERAGE_DIR}/system.cov

# Runs the system tests
.PHONY: system-tests
system-tests-setup: tests/system/requirements.txt
	test -d env || virtualenv build/system-tests/env > /dev/null
	. build/system-tests/env/bin/activate && pip install -Ur tests/system/requirements.txt > /dev/null
	touch build/system-tests/env/bin/activate


# Run benchmark tests
.PHONY: benchmark-tests
benchmark-tests:
	# No benchmark tests exist so far
	#$(GO) test -short -bench=. $(MODULES)

# Runs all tests and generates the coverage reports
.PHONY: testsuite
testsuite:
	make integration-tests-environment
	make system-tests
	make benchmark-tests
	make coverage-report


# Generates a coverage report from the existing coverage files
# It assumes that some covrage reports already exists, otherwise it will fail
.PHONY: coverage-report
coverage-report:
	# Writes count mode on top of file
	echo 'mode: count' > ./${COVERAGE_DIR}/full.cov
	# Collects all coverage files and skips top line with mode
	tail -q -n +2 ./${COVERAGE_DIR}/*.cov >> ./${COVERAGE_DIR}/full.cov
	$(GO) tool cover -html=./${COVERAGE_DIR}/full.cov -o ${COVERAGE_DIR}/full.html



### CONTAINER ENVIRONMENT ####

# Builds the environment to test libbeat
.PHONY: build-image
build-image: write-environment
	docker-compose build

# Runs the environment so the redis and elasticsearch can also be used for local development
# To use it for running the test, set ES_HOST and REDIS_HOST environment variable to the ip of your docker-machine.
.PHONY: start-environment
start-environment: stop-environment
	docker-compose up -d redis elasticsearch-173 elasticsearch-200 logstash

.PHONY: stop-environment
stop-environment:
	-docker-compose stop
	-docker-compose rm -f
	-docker ps -a  | grep libbeat | grep Exited | awk '{print $$1}' | xargs docker rm

.PHONY: write-environment
write-environment:
	mkdir -p build
	echo "ES_HOST=${ES_HOST}" > build/test.env
	echo "ES_PORT=9200" >> build/test.env
