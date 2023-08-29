PKG ?= ./...
TESTARGS ?=

checks:
	go vet ${PKG}
	command -v staticcheck && staticcheck ${PKG}
	command -v golangci-lint && golangci-lint run ${PKG}

test:
	go test ${TESTARGS} ${PKG}

build:
	go build -v ${PKG}

install:
	go install -v ${PKG}
