SHELL := /bin/sh
POWERSHELL ?= powershell

.PHONY: governance-check lint test security-scan

governance-check:
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/check-governance.ps1

lint:
	@if [ ! -f go.mod ]; then \
		echo "未发现 go.mod；当前为治理基线阶段，跳过 golangci-lint。"; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config=.golangci.yml; \
	else \
		echo "未安装 golangci-lint。"; \
		exit 1; \
	fi

test:
	@if [ ! -f go.mod ]; then \
		echo "未发现 go.mod；当前为治理基线阶段，跳过 go test。"; \
	else \
		go test ./...; \
	fi

security-scan:
	@if command -v gitleaks >/dev/null 2>&1; then \
		gitleaks detect --config=.gitleaks.toml --source=. --no-banner; \
	else \
		echo "未安装 gitleaks。"; \
		exit 1; \
	fi
