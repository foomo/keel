.DEFAULT_GOAL:=help

## === Tasks ===

.PHONY: act
## Run github push action
act.push:
	@act

.PHONY: check
## Run tests and linters
check: test lint

.PHONY: test
## Run tests
test:
	gotestsum --format short-verbose ./...

lint%:files=$(shell find . -type f -name go.mod)
lint%:dirs=$(foreach file,$(files),$(dir $(file)) )

.PHONY: lint
## Run linter
lint:
	@for dir in $(dirs); do cd $$dir && golangci-lint run; done

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@for dir in $(dirs); do cd $$dir && golangci-lint run --fix; done

.PHONY: lint.super
## Run super linter
lint.super:
	docker run --rm -it \
		-e 'RUN_LOCAL=true' \
		-e 'DEFAULT_BRANCH=main' \
		-e 'IGNORE_GITIGNORED_FILES=true' \
		-e 'VALIDATE_JSCPD=false' \
		-e 'VALIDATE_GO=false' \
		-v $(PWD):/tmp/lint \
		github/super-linter

## === Utils ===

## Show help text
help:
	@awk '{ \
			if ($$0 ~ /^.PHONY: [a-zA-Z\-\_0-9]+$$/) { \
				helpCommand = substr($$0, index($$0, ":") + 2); \
				if (helpMessage) { \
					printf "\033[36m%-23s\033[0m %s\n", \
						helpCommand, helpMessage; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^[a-zA-Z\-\_0-9.]+:/) { \
				helpCommand = substr($$0, 0, index($$0, ":")); \
				if (helpMessage) { \
					printf "\033[36m%-23s\033[0m %s\n", \
						helpCommand, helpMessage"\n"; \
					helpMessage = ""; \
				} \
			} else if ($$0 ~ /^##/) { \
				if (helpMessage) { \
					helpMessage = helpMessage"\n                        "substr($$0, 3); \
				} else { \
					helpMessage = substr($$0, 3); \
				} \
			} else { \
				if (helpMessage) { \
					print "\n                        "helpMessage"\n" \
				} \
				helpMessage = ""; \
			} \
		}' \
		$(MAKEFILE_LIST)
