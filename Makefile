.DEFAULT_GOAL:=help
-include .makerc

# --- Config ------------------------------------------------------------------

GOMODS=$(shell find . -type f -name go.mod)
# Newline hack for error output
define br


endef

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .mise .lefthook go.work
	@:

.PHONY: .mise
# Install dependencies
.mise:
ifeq (, $(shell command -v mise))
	$(error $(br)$(br)Please ensure you have 'mise' installed and activated!$(br)$(br)  $$ brew update$(br)  $$ brew install mise$(br)$(br)See the documentation: https://mise.jdx.dev/getting-started.html)
endif
	@mise install

.PHONY: .lefthook
# Configure git hooks for lefthook
.lefthook:
	@lefthook install --reset-hooks-path

# Ensure go.work file
go.work:
	@echo "〉initializing go work"
	@go work init && go work use -r . && go work sync

### Tasks

.PHONY: check
## Run lint & tests
check: tidy generate lint test audit

.PHONY: tidy
## Run go mod tidy
tidy: go.work
	@echo "〉go mod tidy"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && go mod tidy) &&) true
	@go work use -r . && go work sync

.PHONY: lint
## Run linter
lint:
	@echo "〉golangci-lint run"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && golangci-lint run) &&) true

.PHONY: lint.fix
## Run golangci-lint & fix
lint.fix:
	@echo "〉golangci-lint run fix"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && golangci-lint run --fix) &&) true

.PHONY: generate
## Run go generate
generate: go.work
	@echo "〉go generate"
	@go generate work

.PHONY: test
## Run tests
test: go.work
	@echo "〉go test"
	@GO_TEST_TAGS=-skip go test -tags=safe -shuffle=on -coverprofile=coverage.out work

.PHONY: test.race
## Run tests with -race
test.race: go.work
	@echo "〉go test with -race"
	@GO_TEST_TAGS=-skip go test -tags=safe -shuffle=on -coverprofile=coverage.out -race work

.PHONY: test.update
## Run tests with -update
test.update: go.work
	@echo "〉go test with -update"
	@GO_TEST_TAGS=-skip go test -tags=safe --shuffle=on coverprofile=coverage.out -update work

.PHONY: test.bench
## Run tests with -bench
test.bench: go.work
	@echo "〉go bench"
	@GO_TEST_TAGS=-skip go test -tags=safe -bench=. -benchmem work

### Dependencies

.PHONY: audit
## Run security audit
audit:
	@echo "〉security audit"
	#@trivy fs . --format table --severity HIGH,CRITICAL
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go govulncheck ./...

.PHONY: outdated
## Show outdated direct dependencies
outdated:
	@echo "〉mise"
	@mise outdated -l --local
	@echo "〉go mod outdated"
	@find . -name 'go.mod' -exec dirname {} \; | xargs -I {} sh -c 'cd {} && go list -u -m -json all' \; | go-mod-outdated -update -direct

.PHONY: upgrade
## Show outdated direct dependencies
upgrade: go.work
	@echo "〉go mod upgrade"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && go get -u ./...) &&) true
	@$(Make) tidy

### Release

.PHONY: tag.submodules
## Create tags for submodules TAG=1.0.0
tag.submodules:
	@echo "$(TAG)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "❌ TAG must be vX.Y.Z format"; exit 1; }
	@git rev-parse "$(TAG)" >/dev/null 2>&1 || { echo "❌ Tag $(TAG) does not exist"; exit 1; }
	@echo "🔖 Creating submodule tags..."
	@find . -type f -name 'go.mod' -mindepth 2 -not -path './vendor/*' -exec sh -c 'dir=$$(dirname {} | sed "s|^\./||"); tag="$$dir/$(TAG)"; git rev-parse "$$tag" >/dev/null 2>&1 || { echo "🔖 $$tag"; git tag "$$tag"; }' \;
	@echo "🔖 Pushing tags..."
	@git push origin --tags

### Documentation

.PHONY: docs
## Open docs
docs:
	@echo "〉starting docs"
	@cd docs && bun install && bun run dev

.PHONY: docs.build
## Open docs
docs.build:
	@echo "〉building docs"
	@cd docs && bun install && bun run build

.PHONY: godocs
## Open go docs
godocs:
	@echo "〉starting go docs"
	@go doc -http

### Utils

.PHONY: help
## Show help text
help:
	@echo "Keel\n"
	@echo "Usage:\n  make [task]"
	@awk '{ \
		if($$0 ~ /^### /){ \
			if(help) printf "%-23s %s\n\n", cmd, help; help=""; \
			printf "\n%s:\n", substr($$0,5); \
		} else if($$0 ~ /^[a-zA-Z0-9._-]+:/){ \
			cmd = substr($$0, 1, index($$0, ":")-1); \
			if(help) printf "  %-23s %s\n", cmd, help; help=""; \
		} else if($$0 ~ /^##/){ \
			help = help ? help "\n                        " substr($$0,3) : substr($$0,3); \
		} else if(help){ \
			print "\n                        " help "\n"; help=""; \
		} \
	}' $(MAKEFILE_LIST)

