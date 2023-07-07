.DEFAULT_GOAL:=help

## === Tasks ===

.PHONY: test
## Run tests
test:
	go test -v ./...

.PHONY: lint
## Run linter
lint: files=$(shell find . -type f -name go.mod)
lint: dirs=$(foreach file,$(files),$(dir $(file)) )
lint:
	@for dir in $(dirs); do cd $$dir && pwd && golangci-lint run; done

.PHONY: lint.fix
## Fix lint violations
lint.fix: files=$(shell find . -type f -name go.mod)
lint.fix: dirs=$(foreach file,$(files),$(dir $(file)) )
lint.fix:
	@for dir in $(dirs); do cd $$dir && golangci-lint run --fix; done

## === Utils ===

.PHONY: gomod
## Run go mod tidy
tidy:
	go mod tidy
	cd example && go mod tidy

.PHONY: gomod.outdated
## Show outdated direct dependencies
outdated:
	go list -u -m -json all | go-mod-outdated -update -direct

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
