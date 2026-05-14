param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
}

function Require-File {
    param([string]$RelativePath)

    $path = Join-Path $Root $RelativePath
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        Add-Failure "Missing required file: $RelativePath"
    }
}

function Read-Text {
    param([string]$RelativePath)

    $path = Join-Path $Root $RelativePath
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        return ""
    }
    return Get-Content -LiteralPath $path -Raw -Encoding UTF8
}

function Get-GoFiles {
    param([string[]]$RelativeRoots)

    $files = @()
    foreach ($relativeRoot in $RelativeRoots) {
        $path = Join-Path $Root $relativeRoot
        if (Test-Path -LiteralPath $path -PathType Container) {
            $files += Get-ChildItem -LiteralPath $path -Recurse -File -Filter "*.go"
        }
    }
    return $files
}

function Invoke-Subcheck {
    param([string]$RelativePath)

    $path = Join-Path $Root $RelativePath
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        Add-Failure "Missing governance subcheck: $RelativePath"
        return
    }

    & $path -Root $Root
    if ($LASTEXITCODE -ne 0) {
        Add-Failure "Governance subcheck failed: $RelativePath"
    }
}

$requiredFiles = @(
    "AGENTS.md",
    "CLAUDE.md",
    "docs/governance/README.md",
    "docs/governance/rules.md",
    "docs/governance/ai-execution.md",
    "docs/governance/change-management.md",
    "docs/governance/automation-matrix.md",
    "docs/governance/metadata-schema.md",
    "docs/governance/exceptions.yaml",
    "docs/governance/exceptions.template.yaml",
    "docs/conventions/layering.md",
    "docs/conventions/pkg.md",
    "docs/conventions/testing.md",
    "docs/conventions/ci.md",
    "docs/conventions/migrations.md",
    "docs/conventions/async-jobs.md",
    "docs/conventions/security-logging.md",
    "docs/conventions/dependency-injection.md",
    "docs/architecture/governance.md",
    "docs/review/checklist.md",
    "docs/review/governance-change-checklist.md",
    "docs/migrations/gray-release-template.md",
    "docs/adr/README.md",
    "docs/adr/0000-template.md",
    "docs/adr/20260515-governance-ssot-structure.md",
    "docs/adr/20260515-default-backend-direction.md",
    "scripts/check-layering.ps1",
    "scripts/check-test-conventions.ps1",
    "scripts/check-go-package-state.ps1",
    "scripts/check-governance-metadata.ps1",
    "scripts/check-governance-taxonomy.ps1",
    "scripts/check-governance-sync.ps1",
    "scripts/check-rule-links.ps1",
    "scripts/check-exception-expiry.ps1",
    "scripts/check-local-paths.ps1",
    ".github/workflows/governance.yml",
    ".golangci.yml",
    ".gitleaks.toml",
    ".pre-commit-config.yaml",
    "Makefile"
)

foreach ($file in $requiredFiles) {
    Require-File $file
}

$adrTemplate = Read-Text "docs/adr/0000-template.md"
foreach ($section in @("ADR-001", "ADR-002", "ADR-003")) {
    if ($adrTemplate -notmatch [regex]::Escape($section)) {
        Add-Failure "ADR template is missing marker: $section"
    }
}
if ($adrTemplate -notmatch "ADR 0000") {
    Add-Failure "ADR template is missing its title placeholder."
}

$checklist = Read-Text "docs/review/checklist.md"
foreach ($term in @("GOV-P0-001", "GOV-P0-002", "GOV-P0-003", "GOV-P0-004", "GOV-P1-001", "GOV-P1-002", "GOV-P1-003", "GOV-P1-004", "GOV-P1-005", "GOV-P1-006")) {
    if ($checklist -notmatch [regex]::Escape($term)) {
        Add-Failure "Review checklist is missing rule: $term"
    }
}

$migrationTemplate = Read-Text "docs/migrations/gray-release-template.md"
foreach ($step in @("Step 1", "Step 2", "Step 3", "Step 4", "MIG-P1-001", "MIG-P1-002")) {
    if ($migrationTemplate -notmatch [regex]::Escape($step)) {
        Add-Failure "Migration template is missing a required marker: $step"
    }
}

foreach ($subcheck in @(
    "scripts/check-layering.ps1",
    "scripts/check-test-conventions.ps1",
    "scripts/check-governance-metadata.ps1",
    "scripts/check-governance-taxonomy.ps1",
    "scripts/check-governance-sync.ps1",
    "scripts/check-rule-links.ps1",
    "scripts/check-exception-expiry.ps1",
    "scripts/check-local-paths.ps1"
)) {
    Invoke-Subcheck $subcheck
}

$allGoFiles = Get-GoFiles @("cmd", "internal", "pkg")
foreach ($file in $allGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8
    if ($content -match "crypto/md5" -or $content -match "crypto/sha1") {
        Add-Failure "P0 risky hash import detected: $($file.FullName)"
    }
    if ($content -match "zap\.Any\s*\(") {
        Add-Failure "P0 sensitive logging risk: found zap.Any in $($file.FullName)"
    }
    if ($content -match '%\+v') {
        Add-Failure "P0 sensitive logging risk: found %+v formatting in $($file.FullName)"
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Governance check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Governance check passed." -ForegroundColor Green
exit 0
