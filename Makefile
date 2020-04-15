.PHONY: all build clean update fmt test lint vendor

## overridable Makefile variables
# test to run
TESTSET = .
# benchmarks to run
BENCHSET ?= .

# version (defaults to short git hash)
VERSION ?= $(shell git rev-parse --short HEAD)

# use correct sed for platform
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    SED := gsed
else
    SED := sed
endif

PKG_NAME=github.com/danieloliveira079/agones-controller-sample

LDFLAGS := -X "${PKG_NAME}/internal/version.Version=${VERSION}"
LDFLAGS += -X "${PKG_NAME}/internal/version.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "${PKG_NAME}/internal/version.GitCommit=$(shell git rev-parse HEAD)"
LDFLAGS += -X "${PKG_NAME}/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"
LDFLAGS += -X "${PKG_NAME}/internal/version.GoVersion=$(shell go version)"

GO       := GO111MODULE=on GOPRIVATE=github.com/danieloliveira GOSUMDB=off go
GOBUILD  := CGO_ENABLED=0 $(GO) build $(BUILD_FLAG)
GOTEST   := $(GO) test -gcflags='-l' -p 3


FILES    := $(shell find internal cmd -name '*.go' -type f -not -name '*.pb.go' -not -name '*_generated.go' -not -name '*_test.go')
TESTS    := $(shell find internal cmd -name '*.go' -type f -not -name '*.pb.go' -not -name '*_generated.go' -name '*_test.go')

CONTROLLER_BIN := bin/agones-controller

default: clean build

build: $(CONTROLLER_BIN)

$(CONTROLLER_BIN):
	CGO_ENABLED=0 GOOS=linux go build -ldflags '$(LDFLAGS)' -o $(CONTROLLER_BIN) $(PKG_NAME)/cmd

dist: clean
	CGO_ENABLED=0 GOOS=darwin go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(CONTROLLER_BIN)-darwin $(PKG_NAME)/cmd
	# Not tested on the platforms below
	#CGO_ENABLED=0 GOOS=linux go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(CONTROLLER_BIN) $(PKG_NAME)/cmd
	# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(CONTROLLER_BIN).exe $(PKG_NAME)/cmd
	#CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(CONTROLLER_BIN)-armhf $(PKG_NAME)/cmd
	#CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(CONTROLLER_BIN)-arm64 $(PKG_NAME)/cmd

release: clean dist
	zip -r $(CONTROLLER_BIN)-all-platforms.zip bin/agones-controller*

clean:
	rm -f $(CONTROLLER_BIN)*

get:
	$(GO) get ./...
	$(GO) mod verify
	$(GO) mod tidy

update:
	$(GO) get -u -v all
	$(GO) mod verify
	$(GO) mod tidy

fmt:
	gofmt -s -l -w $(FILES) $(TESTS)

lint:
	golangci-lint run

test:
	$(GOTEST) -run=$(TESTSET) ./...
	@echo
	@echo Configured tests ran ok.

test-strict:
	$(GO) test -p 3 -run=$(TESTSET) -gcflags='-l -m' -race ./...
	@echo
	@echo Configured tests ran ok.

bench:
	DEBUG=0 $(GOTEST) -run=nothing -bench=$(BENCHSET) -benchmem ./...
	@echo
	@echo Configured benchmarks ran ok.

vendor:
	$(GO) mod vendor