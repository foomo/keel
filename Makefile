.DEFAULT_GOAL:=help
-include .makerc

# --- Config -----------------------------------------------------------------

GOMODS=$(shell find . -type f -name go.mod)
# Newline hack for error output
define br


endef

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .mise .husky go.work
	@:

# Ensure go.work file
go.work:
	@go work init
	@go work use -r .
	@go work sync

.PHONY: .mise
# Install dependencies
.mise: msg := $(br)$(br)Please ensure you have 'mise' installed and activated!$(br)$(br)$$ brew update$(br)$$ brew install mise$(br)$(br)See the documentation: https://mise.jdx.dev/getting-started.html$(br)$(br)
.mise:
ifeq (, $(shell command -v mise))
	$(error ${msg})
endif
	@mise install

.PHONY: .husky
# Configure git hooks for husky
.husky:
	@git config core.hooksPath .husky

### Tasks

.PHONY: check
## Run lint & tests
check: tidy lint test

.PHONY: tidy
## Run go mod tidy
tidy:
	@echo "〉go mod tidy"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && go mod tidy) &&) true

.PHONY: lint
## Run linter
lint:
	@echo "〉golangci-lint run"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && golangci-lint run) &&) true

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@echo "〉golangci-lint run fix"
	@$(foreach mod,$(GOMODS), (cd $(dir $(mod)) && echo "📂 $(dir $(mod))" && golangci-lint run --fix) &&) true

.PHONY: test
## Run tests
test:
	@echo "〉go test"
	@GO_TEST_TAGS=-skip go test -coverprofile=coverage.out -tags=safe -race work

.PHONY: outdated
## Show outdated direct dependencies
outdated:
	@echo "〉go mod outdated"
	@go list -u -m -json all | go-mod-outdated -update -direct

.PHONY: release
## Create release TAG=1.0.0
release: MODS=$(shell find . -type f -name 'go.mod' -mindepth 2 -not -path './examples/*')
release:
ifndef TAG
	$(error $(br)$(br)TAG variable is required.$(br)Usage: make release TAG=1.0.0$(br)$(br))
endif
	@echo "〉️Create release"
	@echo "🔖 v$(TAG)" && git tag v$(TAG)
	@$(foreach mod,$(MODS), (echo "🔖 $(patsubst %/,%,$(patsubst ./%,%,$(basename $(dir $(mod)))))/v$(TAG)" && git tag $(patsubst %/,%,$(patsubst ./%,%,$(basename $(dir $(mod)))))/v$(TAG)") &&) true
	@echo
	@read -p "Do you want to push the tags to the remote? [y/N] " yn; \
	@case $$yn in [Yy]*) git push origin --tags ;; *) echo "Skipping git push." ;; esac

### Utils

.PHONY: docs
## Open go docs
docs:
	@go doc -http

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

