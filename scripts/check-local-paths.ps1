param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
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

$files = @()
$agents = Join-Path $Root "AGENTS.md"
if (Test-Path -LiteralPath $agents -PathType Leaf) {
    $files += Get-Item -LiteralPath $agents
}

$claude = Join-Path $Root "CLAUDE.md"
if (Test-Path -LiteralPath $claude -PathType Leaf) {
    $files += Get-Item -LiteralPath $claude
}

$docsPath = Join-Path $Root "docs"
if (Test-Path -LiteralPath $docsPath -PathType Container) {
    $files += Get-ChildItem -LiteralPath $docsPath -Recurse -File -Include "*.md", "*.yaml", "*.yml"
}

$windowsDrivePath = "(?i)(^|[\s\(\[`"'])/?[A-Z]:[\\/][^\s\)\]]+"
$localUnixPath = "(?i)(^|[\s\(\[`"'])/(Users|home|mnt|workspaces|workspace|tmp|var/folders)/[^\s\)\]]+"

foreach ($file in $files) {
    $relativePath = Get-RelativePath $Root $file.FullName
    $lineNumber = 0
    foreach ($line in Get-Content -LiteralPath $file.FullName -Encoding UTF8) {
        $lineNumber++
        if ($line -match $windowsDrivePath -or $line -match $localUnixPath) {
            Add-Failure "Local absolute path found: ${relativePath}:${lineNumber}"
        }
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Local path check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Local path check passed." -ForegroundColor Green
exit 0
