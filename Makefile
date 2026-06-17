# Project Navigator — build & install tasks.
# Run `make` (or `make help`) to list targets.

.DEFAULT_GOAL := help

BINARY   ?= pn
SHELL_RC ?= $(HOME)/.zshrc

.PHONY: help build install uninstall shell-init test clean

help: ## List available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

build: ## Build the pn binary into ./$(BINARY)
	go build -o $(BINARY) ./cmd/pn

install: ## Install pn into your Go bin dir (go env GOBIN / GOPATH/bin)
	go install ./cmd/pn
	@echo "Installed $(BINARY). Ensure your Go bin dir is on PATH:"
	@echo "  export PATH=\"$$(go env GOPATH)/bin:$$PATH\""
	@echo "Next: run 'make shell-init' to enable the cd wrapper."

uninstall: ## Remove the go-installed pn binary
	rm -f "$$(go env GOPATH)/bin/$(BINARY)"

shell-init: ## Append the cd wrapper function to $(SHELL_RC)
	@if grep -q 'command pn "$$@"' $(SHELL_RC) 2>/dev/null; then \
		echo "pn() wrapper already present in $(SHELL_RC) — nothing to do."; \
	else \
		printf '\n# project-navigator: let `pn` change the shell directory\npn() {\n  result=$$(command pn "$$@")\n  if echo "$$result" | grep -q "^cd "; then\n    eval "$$result"\n  else\n    echo "$$result"\n  fi\n}\n' >> $(SHELL_RC) ; \
		echo "Added pn() wrapper to $(SHELL_RC). Reload your shell: exec $$SHELL"; \
	fi

test: ## Vet and compile-check
	go vet ./...
	go build -o /dev/null ./cmd/pn

clean: ## Remove the built binary
	rm -f $(BINARY)
