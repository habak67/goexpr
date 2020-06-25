# Get GOPATH from the current go environment
GOPATH := $(shell go env | grep GOPATH= | sed 's/^GOPATH=\"\(.*\)\"$$/\1/')
HABAK_PATH := github.com/habak67
PROJ_PATH := $(HABAK_PATH)/goexpr

build:
	go build $(PROJ_PATH)

test:
	go test $(PROJ_PATH)
