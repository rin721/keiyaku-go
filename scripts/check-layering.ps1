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

function Test-ImportPattern {
    param(
        [string]$Content,
        [string]$Pattern
    )

    return $Content -match $Pattern
}

$pkgGoFiles = Get-GoFiles @("pkg")
foreach ($file in $pkgGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $importsInternal = (Test-ImportPattern $content '"[^"]*/internal(/[^"]*)?"') -or (Test-ImportPattern $content '"internal/[^"]+"')
    if ($importsInternal) {
        Add-Failure "GOV-P0-002: pkg must not import internal packages: $($file.FullName)"
    }
}

$domainGoFiles = Get-GoFiles @("internal/domain")
foreach ($file in $domainGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $importsForbiddenLayer = (Test-ImportPattern $content '"[^"]*/internal/(api|handler|application|repository|infrastructure)(/[^"]*)?"') -or
        (Test-ImportPattern $content '"internal/(api|handler|application|repository|infrastructure)(/[^"]*)?"')
    if ($importsForbiddenLayer) {
        Add-Failure "GOV-P1-002: domain must not import upper layers or adapters: $($file.FullName)"
    }
}

$applicationGoFiles = Get-GoFiles @("internal/application")
foreach ($file in $applicationGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $importsTransport = (Test-ImportPattern $content '"[^"]*/internal/(api|handler)(/[^"]*)?"') -or
        (Test-ImportPattern $content '"internal/(api|handler)(/[^"]*)?"')
    if ($importsTransport) {
        Add-Failure "GOV-P1-002: application must not import transport packages: $($file.FullName)"
    }
}

$adapterGoFiles = Get-GoFiles @("internal/repository", "internal/infrastructure")
foreach ($file in $adapterGoFiles) {
    $content = Get-Content -LiteralPath $file.FullName -Raw
    $importsTransport = (Test-ImportPattern $content '"[^"]*/internal/(api|handler)(/[^"]*)?"') -or
        (Test-ImportPattern $content '"internal/(api|handler)(/[^"]*)?"')
    if ($importsTransport) {
        Add-Failure "GOV-P0-002: repository/infrastructure must not import transport packages: $($file.FullName)"
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Layering check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Layering check passed." -ForegroundColor Green
exit 0
