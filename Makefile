GO ?= go
GOLANGCI_LINT_VERSION ?= v2.12.2
ACTIONLINT_VERSION ?= v1.7.12
GOVULNCHECK_VERSION ?= latest
GOVULNCHECK_TOOLCHAIN ?= go1.26.4+auto

PKGS := ./...
GO_DIRS := cmd internal

.PHONY: verify fmt fmt-check tidy tidy-check vet lint test test-race cover build vuln workflow-lint clean

verify: tidy-check fmt-check vet lint test test-race build vuln workflow-lint

fmt:
	gofmt -w $(GO_DIRS)

fmt-check:
	@files="$$(gofmt -l $(GO_DIRS))"; \
	if [ -n "$$files" ]; then \
		echo "Go files need formatting:"; \
		echo "$$files"; \
		echo "Run: make fmt"; \
		exit 1; \
	fi

tidy:
	$(GO) mod tidy

tidy-check:
	$(GO) mod tidy -diff

vet:
	$(GO) vet $(PKGS)

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run $(PKGS)

test:
	$(GO) test $(PKGS)

test-race:
	$(GO) test -race $(PKGS)

cover:
	$(GO) test -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -func=coverage.out

build:
	$(GO) build -trimpath -o bin/nerdfont-install ./cmd/nerdfont-install

vuln:
	GOTOOLCHAIN=$(GOVULNCHECK_TOOLCHAIN) $(GO) run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) $(PKGS)

workflow-lint:
	$(GO) run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)

clean:
	rm -rf bin dist coverage.out
