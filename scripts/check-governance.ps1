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
    return Get-Content -LiteralPath $path -Raw
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

$requiredFiles = @(
    "docs/architecture/governance.md",
    "docs/review/checklist.md",
    "docs/migrations/gray-release-template.md",
    "docs/adr/README.md",
    "docs/adr/0000-template.md",
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
foreach ($section in @("Background and Context", "Decision", "Consequences")) {
    if ($adrTemplate -notmatch [regex]::Escape($section)) {
        Add-Failure "ADR template must include section: $section"
    }
}
if ($adrTemplate -notmatch "Status:") {
    Add-Failure "ADR template must include a Status field"
}

$checklist = Read-Text "docs/review/checklist.md"
foreach ($term in @("Model safety", "Contract safety", "Sensitive logging", "Traceability", "Idempotency")) {
    if ($checklist -notmatch [regex]::Escape($term)) {
        Add-Failure "Review checklist must cover: $term"
    }
}

$migrationTemplate = Read-Text "docs/migrations/gray-release-template.md"
foreach ($step in @("Step 1", "Step 2", "Step 3", "Step 4", "Dual Write", "Backfill")) {
    if ($migrationTemplate -notmatch [regex]::Escape($step)) {
        Add-Failure "Migration template must include: $step"
    }
}

$governance = Read-Text "docs/architecture/governance.md"
foreach ($rule in @("No Model Penetration", "No Reverse Dependency", "No Unsafe Password Hashing", "No Plaintext Sensitive Data Logging", "Traceability", "Pragmatic DDD Layering")) {
    if ($governance -notmatch [regex]::Escape($rule)) {
        Add-Failure "Governance document must cover: $rule"
    }
}

$pkgGoFiles = Get-GoFiles @("pkg")
foreach ($file in $pkgGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    if ($content -match '"[^"]*/internal(/[^"]*)?"' -or $content -match '"internal/[^"]+"') {
        Add-Failure "P0 dependency violation: pkg code imports internal package in $($file.FullName)"
    }
}

$lowerLayerGoFiles = Get-GoFiles @("internal/domain", "internal/repository", "internal/infrastructure")
foreach ($file in $lowerLayerGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    if ($content -match '"[^"]*/internal/(api|handler)(/[^"]*)?"' -or $content -match '"internal/(api|handler)/[^"]+"') {
        Add-Failure "P0 contract dependency violation: lower layer imports transport contract in $($file.FullName)"
    }
}

$allGoFiles = Get-GoFiles @("cmd", "internal", "pkg")
foreach ($file in $allGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    if ($content -match "crypto/md5" -or $content -match "crypto/sha1") {
        Add-Failure "P0 unsafe hash import found in $($file.FullName)"
    }
    if ($content -match "zap\.Any\s*\(") {
        Add-Failure "P0 sensitive logging risk: zap.Any found in $($file.FullName)"
    }
    if ($content -match '%\+v') {
        Add-Failure "P0 sensitive logging risk: %+v formatting found in $($file.FullName)"
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
