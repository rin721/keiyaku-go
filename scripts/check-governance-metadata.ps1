param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
}

function Get-FrontMatter {
    param([string]$Content)

    if ($Content -notmatch "(?s)^---\r?\n(.*?)\r?\n---") {
        return $null
    }
    return $Matches[1]
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

$requiredFields = @(
    "doc_role",
    "scope",
    "authority_level",
    "owners",
    "status",
    "related_rules",
    "read_when",
    "update_when"
)

$files = @()
$files += Get-ChildItem -LiteralPath $Root -File -Filter "AGENTS.md"
$claude = Join-Path $Root "CLAUDE.md"
if (Test-Path -LiteralPath $claude -PathType Leaf) {
    $files += Get-Item -LiteralPath $claude
}
$docsPath = Join-Path $Root "docs"
if (Test-Path -LiteralPath $docsPath -PathType Container) {
    $files += Get-ChildItem -LiteralPath $docsPath -Recurse -File -Include "*.md", "*.yaml", "*.yml"
}

foreach ($file in $files) {
    $relativePath = Get-RelativePath $Root $file.FullName
    $content = Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8
    $frontMatter = Get-FrontMatter $content

    if ($null -eq $frontMatter) {
        Add-Failure "Missing metadata front matter: $relativePath"
        continue
    }

    foreach ($field in $requiredFields) {
        if ($frontMatter -notmatch "(?m)^$([regex]::Escape($field))\s*:") {
            Add-Failure "Missing metadata field '$field': $relativePath"
        }
    }

    if ($frontMatter -notmatch "(?m)^(version|effective_date)\s*:") {
        Add-Failure "Missing metadata field 'version' or 'effective_date': $relativePath"
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Governance metadata check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Governance metadata check passed." -ForegroundColor Green
exit 0
