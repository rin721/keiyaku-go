SHELL := /bin/sh
POWERSHELL ?= powershell

.PHONY: governance-check lint test security-scan

governance-check:
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/check-governance.ps1

lint:
	@if [ ! -f go.mod ]; then \
		echo "No go.mod found; skipping golangci-lint for governance-only baseline."; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config=.golangci.yml; \
	else \
		echo "golangci-lint is not installed."; \
		exit 1; \
	fi

test:
	@if [ ! -f go.mod ]; then \
		echo "No go.mod found; skipping go test for governance-only baseline."; \
	else \
		go test ./...; \
	fi

security-scan:
	@if command -v gitleaks >/dev/null 2>&1; then \
		gitleaks detect --config=.gitleaks.toml --source=. --no-banner; \
	else \
		echo "gitleaks is not installed."; \
		exit 1; \
	fi
