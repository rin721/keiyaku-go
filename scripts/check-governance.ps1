param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function T {
    param([string]$Base64)
    return [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($Base64))
}

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
}

function Require-File {
    param([string]$RelativePath)
    $path = Join-Path $Root $RelativePath
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        Add-Failure "$((T '57y65bCR5b+F6ZyA5paH5Lu277ya'))$RelativePath"
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
        Add-Failure "Missing subcheck script: $RelativePath"
        return
    }

    & $path -Root $Root
    if ($LASTEXITCODE -ne 0) {
        Add-Failure "Subcheck failed: $RelativePath"
    }
}

$requiredFiles = @(
    "AGENTS.md",
    "docs/governance/README.md",
    "docs/governance/ai-execution.md",
    "docs/governance/automation-matrix.md",
    "docs/governance/exceptions.yaml",
    "docs/conventions/layering.md",
    "docs/architecture/governance.md",
    "docs/review/checklist.md",
    "docs/migrations/gray-release-template.md",
    "docs/adr/README.md",
    "docs/adr/0000-template.md",
    "scripts/check-layering.ps1",
    "scripts/check-test-conventions.ps1",
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
        Add-Failure "$((T 'QURSIOaooeadv+W/hemhu+WMheWQq+ajgOafpeeCue+8mg=='))$section"
    }
}
if ($adrTemplate -notmatch "ADR 0000") {
    Add-Failure (T 'QURSIOaooeadv+W/hemhu+WMheWQqyBBRFIg5qCH6aKY')
}

$checklist = Read-Text "docs/review/checklist.md"
foreach ($term in @("REV-P0-001", "REV-P0-002", "REV-P0-003", "REV-P1-001", "REV-P1-002")) {
    if ($checklist -notmatch [regex]::Escape($term)) {
        Add-Failure "$((T '5Luj56CB6K+E5a6h5riF5Y2V5b+F6aG75YyF5ZCr5qOA5p+l54K577ya'))$term"
    }
}

$migrationTemplate = Read-Text "docs/migrations/gray-release-template.md"
foreach ($step in @("Step 1", "Step 2", "Step 3", "Step 4", "MIG-P1-001", "MIG-P1-002")) {
    if ($migrationTemplate -notmatch [regex]::Escape($step)) {
        Add-Failure "$((T 'TWlncmF0aW9uIOaooeadv+W/hemhu+WMheWQq+ajgOafpeeCueaIluatpemqpO+8mg=='))$step"
    }
}

$governance = Read-Text "docs/architecture/governance.md"
foreach ($rule in @("GOV-P0-001", "GOV-P0-002", "GOV-P0-003", "GOV-P0-004", "GOV-P1-001", "GOV-P1-002")) {
    if ($governance -notmatch [regex]::Escape($rule)) {
        Add-Failure "$((T '5rK755CG5paH5qGj5b+F6aG75YyF5ZCr5qOA5p+l54K577ya'))$rule"
    }
}

Invoke-Subcheck "scripts/check-layering.ps1"
Invoke-Subcheck "scripts/check-test-conventions.ps1"

$allGoFiles = Get-GoFiles @("cmd", "internal", "pkg")
foreach ($file in $allGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    if ($content -match "crypto/md5" -or $content -match "crypto/sha1") {
        Add-Failure "$((T 'UDAg6auY5Y2x5ZOI5biMIGltcG9ydO+8mg=='))$($file.FullName)"
    }
    if ($content -match "zap\.Any\s*\(") {
        Add-Failure "$((T 'UDAg5pWP5oSf5pel5b+X6aOO6Zmp77ya5Y+R546wIHphcC5BbnnvvJo='))$($file.FullName)"
    }
    if ($content -match '%\+v') {
        Add-Failure "$((T 'UDAg5pWP5oSf5pel5b+X6aOO6Zmp77ya5Y+R546wICUrdiDmoLzlvI/ljJbvvJo='))$($file.FullName)"
    }
}

if ($failures.Count -gt 0) {
    Write-Host (T '5rK755CG5qOA5p+l5aSx6LSl77ya') -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host (T '5rK755CG5qOA5p+l6YCa6L+H44CC') -ForegroundColor Green
