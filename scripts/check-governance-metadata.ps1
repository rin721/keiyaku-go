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
    "state_id",
    "doc_role",
    "memory_level",
    "state_scope",
    "scope",
    "authority_level",
    "owners",
    "status",
    "related_rules",
    "source_of_truth",
    "derived_from",
    "read_when",
    "update_when",
    "conflict_policy",
    "rollback_target",
    "verification_target"
)

function Parse-InlineList {
    param([string]$FrontMatter, [string]$FieldName)

    if ($FrontMatter -notmatch "(?m)^$([regex]::Escape($FieldName))\s*:\s*\[(.*)\]\s*$") {
        return @()
    }

    $raw = $Matches[1].Trim()
    if ([string]::IsNullOrWhiteSpace($raw)) {
        return @()
    }

    return $raw.Split(",") | ForEach-Object { $_.Trim() } | Where-Object { $_ }
}

function Parse-SingleValue {
    param([string]$FrontMatter, [string]$FieldName)

    if ($FrontMatter -match "(?m)^$([regex]::Escape($FieldName))\s*:\s*(.+)\s*$") {
        return $Matches[1].Trim()
    }
    return ""
}

$files = @()
$files += Get-ChildItem -LiteralPath $Root -File -Filter "AGENTS.md"
$claude = Join-Path $Root "CLAUDE.md"
if (Test-Path -LiteralPath $claude -PathType Leaf) {
    $files += Get-Item -LiteralPath $claude
}
$docsPath = Join-Path $Root "docs"
if (Test-Path -LiteralPath $docsPath -PathType Container) {
    $files += Get-ChildItem -LiteralPath $docsPath -Recurse -File | Where-Object { $_.Extension -in @(".md", ".yaml", ".yml") }
}

$stateIds = @{}
$taskEntryPoints = @()

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

    $stateId = Parse-SingleValue $frontMatter "state_id"
    if ([string]::IsNullOrWhiteSpace($stateId)) {
        Add-Failure "Empty state_id: $relativePath"
    } elseif ($stateIds.ContainsKey($stateId)) {
        Add-Failure "Duplicate state_id '$stateId': $relativePath and $($stateIds[$stateId])"
    } else {
        $stateIds[$stateId] = $relativePath
    }

    foreach ($listField in @("owners", "read_when", "update_when", "source_of_truth", "rollback_target", "verification_target")) {
        $values = Parse-InlineList $frontMatter $listField
        if ($values.Count -eq 0) {
            Add-Failure "Metadata field '$listField' must not be empty: $relativePath"
        }
    }

    foreach ($scalarField in @("memory_level", "state_scope", "conflict_policy")) {
        if ([string]::IsNullOrWhiteSpace((Parse-SingleValue $frontMatter $scalarField))) {
            Add-Failure "Metadata field '$scalarField' must not be empty: $relativePath"
        }
    }

    $taskEntrypoint = Parse-SingleValue $frontMatter "task_entrypoint"
    if (-not [string]::IsNullOrWhiteSpace($taskEntrypoint)) {
        if ($taskEntrypoint -notin @("true", "false")) {
            Add-Failure "task_entrypoint must be 'true' or 'false': $relativePath"
        } elseif ($taskEntrypoint -eq "true") {
            $taskEntryPoints += $relativePath
        }
    }
}

if ($taskEntryPoints.Count -ne 1) {
    Add-Failure "Exactly one governance document must set task_entrypoint: true"
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
