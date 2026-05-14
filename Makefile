SHELL := /bin/sh
POWERSHELL ?= powershell

.PHONY: governance-check lint test security-scan

governance-check:
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/check-governance.ps1

lint:
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/check-go-package-state.ps1; \
	status=$$?; \
	if [ $$status -eq 2 ]; then \
		echo "未发现 go.mod；跳过 golangci-lint。"; \
	elif [ $$status -eq 3 ]; then \
		echo "未发现可分析 Go package；跳过 golangci-lint。"; \
	elif [ $$status -ne 0 ]; then \
		exit $$status; \
	elif command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config=.golangci.yml; \
	else \
		echo "未安装 golangci-lint。"; \
		exit 1; \
	fi

test:
	@$(POWERSHELL) -NoProfile -ExecutionPolicy Bypass -File scripts/check-go-package-state.ps1; \
	status=$$?; \
	if [ $$status -eq 2 ]; then \
		echo "未发现 go.mod；跳过 go test。"; \
	elif [ $$status -eq 3 ]; then \
		echo "未发现可测试 Go package；跳过 go test。"; \
	elif [ $$status -ne 0 ]; then \
		exit $$status; \
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
