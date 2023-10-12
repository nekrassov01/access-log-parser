GOBIN ?= $(shell go env GOPATH)/bin
VERSION := $$(make -s show-version)

HAS_LINT := $(shell command -v $(GOBIN)/golangci-lint 2> /dev/null)
HAS_STATICCHECK := $(shell command -v $(GOBIN)/staticcheck 2> /dev/null)
HAS_GOSEC := $(shell command -v $(GOBIN)/gosec 2> /dev/null)
HAS_VULNCHECK := $(shell command -v $(GOBIN)/govulncheck 2> /dev/null)
HAS_GOBUMP := $(shell command -v $(GOBIN)/gobump 2> /dev/null)

BIN_LINT := github.com/golangci/golangci-lint/cmd/golangci-lint@latest
BIN_STATICCHECK := honnef.co/go/tools/cmd/staticcheck@latest
BIN_GOSEC := github.com/securego/gosec/v2/cmd/gosec@latest
BIN_GOVULNCHECK := golang.org/x/vuln/cmd/govulncheck@latest
BIN_GOBUMP := github.com/x-motemen/gobump/cmd/gobump@latest

export GO111MODULE=on

.PHONY: check
check: test vet golangci-lint gosec govulncheck

.PHONY: deps
deps:
ifndef HAS_LINT
	go install $(BIN_LINT)
endif
#ifndef HAS_STATICCHECK
#	go install $(BIN_STATICCHECK)
#endif
ifndef HAS_GOSEC
	go install $(BIN_GOSEC)
endif
ifndef HAS_VULNCHECK
	go install $(BIN_GOVULNCHECK)
endif
ifndef HAS_GOBUMP
	go install $(BIN_GOBUMP)
endif

.PHONY: golangci-lint
golangci-lint:
	golangci-lint run -v

#.PHONY: staticcheck
#staticcheck: deps
#	$(GOBIN)/staticcheck -checks all

.PHONY: gosec
gosec: deps
	$(GOBIN)/gosec ./...

.PHONY: govulncheck
govulncheck: deps
	$(GOBIN)/govulncheck ./...

.PHONY: show-version
show-version: deps
	$(GOBIN)/gobump show -r .

.PHONY: publish
publish: deps
ifneq ($(shell git status --porcelain),)
	$(error git workspace is dirty)
endif
ifneq ($(shell git rev-parse --abbrev-ref HEAD),main)
	$(error current branch is not main)
endif
	@gobump up -w .
	git commit -am "bump up version to $(VERSION)"
	git tag "v$(VERSION)"
	git push origin main
	git push origin "refs/tags/v$(VERSION)"

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test -race -cover -v ./... -coverprofile=cover.out -covermode=atomic

.PHONY: cover
cover:
	go tool cover -html=cover.out -o cover.html

.PHONY: clean
clean:
	go clean
