param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"

function Get-GoPackageFiles {
    param([string[]]$RelativeRoots)

    $files = @()
    foreach ($relativeRoot in $RelativeRoots) {
        $path = Join-Path $Root $relativeRoot
        if (Test-Path -LiteralPath $path -PathType Container) {
            $files += Get-ChildItem -LiteralPath $path -Recurse -File -Filter "*.go" |
                Where-Object { $_.Name -notlike "*_test.go" }
        }
    }
    return $files
}

$goMod = Join-Path $Root "go.mod"
if (-not (Test-Path -LiteralPath $goMod -PathType Leaf)) {
    Write-Host "Go module absent."
    exit 2
}

$goFiles = Get-GoPackageFiles @("cmd", "internal", "pkg")
if ($goFiles.Count -eq 0) {
    Write-Host "Go module present, but no analyzable Go packages found."
    exit 3
}

Write-Host "Go module and analyzable Go packages present."
exit 0
