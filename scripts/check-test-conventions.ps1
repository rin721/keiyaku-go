param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
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

function Has-BuildTag {
    param(
        [string]$Content,
        [string]$Tag
    )

    return $Content -match "(?m)^//go:build\s+.*\b$([regex]::Escape($Tag))\b"
}

function Get-RelativePath {
    param(
        [string]$BasePath,
        [string]$FullPath
    )

    $base = (Resolve-Path -LiteralPath $BasePath).Path.TrimEnd("\") + "\"
    $full = (Resolve-Path -LiteralPath $FullPath).Path
    return $full.Substring($base.Length).Replace("\", "/")
}

$testFiles = Get-GoFiles @("cmd", "internal", "pkg") | Where-Object { $_.Name -like "*_test.go" }

foreach ($file in $testFiles) {
    $relativePath = Get-RelativePath $Root $file.FullName
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $importsTestcontainers = $content -match '"github\.com/testcontainers/testcontainers-go(/[^"]*)?"'
    $hasIntegrationTag = Has-BuildTag $content "integration"

    if ($importsTestcontainers -and -not $hasIntegrationTag) {
        Add-Failure "GOV-P1-TEST: testcontainers tests must use the integration build tag: $relativePath"
    }

    if (($relativePath -like "internal/domain/*" -or $relativePath -like "internal/application/*") -and $importsTestcontainers) {
        Add-Failure "GOV-P1-TEST: domain/application tests must stay fast and must not import testcontainers: $relativePath"
    }

    if ($hasIntegrationTag -and -not $importsTestcontainers -and ($relativePath -like "internal/repository/*" -or $relativePath -like "internal/infrastructure/*")) {
        Add-Failure "GOV-P1-TEST: repository/infrastructure integration tests should use real middleware via testcontainers: $relativePath"
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Test convention check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Test convention check passed." -ForegroundColor Green
exit 0
