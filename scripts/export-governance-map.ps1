param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path,
    [string]$OutputPath = "docs/governance/governance-map.json"
)

$ErrorActionPreference = "Stop"

function Get-GovernanceFiles {
    param([string]$RepositoryRoot)

    $files = @()
    $files += Get-ChildItem -LiteralPath $RepositoryRoot -File -Filter "AGENTS.md"

    $claude = Join-Path $RepositoryRoot "CLAUDE.md"
    if (Test-Path -LiteralPath $claude -PathType Leaf) {
        $files += Get-Item -LiteralPath $claude
    }

    $docsPath = Join-Path $RepositoryRoot "docs"
    if (Test-Path -LiteralPath $docsPath -PathType Container) {
        $files += Get-ChildItem -LiteralPath $docsPath -Recurse -File | Where-Object { $_.Extension -in @(".md", ".yaml", ".yml") }
    }

    return $files | Sort-Object FullName
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

function Get-FrontMatter {
    param([string]$Content)

    if ($Content -notmatch "(?s)^---\r?\n(.*?)\r?\n---") {
        return $null
    }

    return $Matches[1]
}

function Get-Body {
    param([string]$Content)

    if ($Content -notmatch "(?s)^---\r?\n.*?\r?\n---\r?\n?(.*)$") {
        return $Content
    }

    return $Matches[1]
}

function Parse-InlineList {
    param(
        [string]$FrontMatter,
        [string]$FieldName
    )

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
    param(
        [string]$FrontMatter,
        [string]$FieldName
    )

    if ($FrontMatter -match "(?m)^$([regex]::Escape($FieldName))\s*:\s*(.+)\s*$") {
        return $Matches[1].Trim()
    }

    return ""
}

function Parse-Boolean {
    param(
        [string]$FrontMatter,
        [string]$FieldName
    )

    $value = Parse-SingleValue $FrontMatter $FieldName
    if ([string]::IsNullOrWhiteSpace($value)) {
        return $null
    }

    if ($value -eq "true") {
        return $true
    }

    if ($value -eq "false") {
        return $false
    }

    throw "Invalid boolean value for ${FieldName}: $value"
}

function Get-Title {
    param(
        [string]$RelativePath,
        [string]$Content
    )

    if ($RelativePath.EndsWith(".md")) {
        $body = Get-Body $Content
        foreach ($line in ($body -split "\r?\n")) {
            if ($line -match "^#\s+(.+)$") {
                return $Matches[1].Trim()
            }
        }
    }

    return [System.IO.Path]::GetFileName($RelativePath)
}

function Build-StateObject {
    param(
        [string]$RelativePath,
        [string]$Content,
        [string]$FrontMatter,
        [DateTime]$LastWriteTimeUtc
    )

    $state = [ordered]@{
        state_id = Parse-SingleValue $FrontMatter "state_id"
        title = Get-Title $RelativePath $Content
        file_path = $RelativePath
        doc_role = Parse-SingleValue $FrontMatter "doc_role"
        memory_level = Parse-SingleValue $FrontMatter "memory_level"
        state_scope = Parse-SingleValue $FrontMatter "state_scope"
        scope = Parse-SingleValue $FrontMatter "scope"
        authority_level = Parse-SingleValue $FrontMatter "authority_level"
        owners = @(Parse-InlineList $FrontMatter "owners")
        status = Parse-SingleValue $FrontMatter "status"
        related_rules = @(Parse-InlineList $FrontMatter "related_rules")
        source_of_truth = @(Parse-InlineList $FrontMatter "source_of_truth")
        derived_from = @(Parse-InlineList $FrontMatter "derived_from")
        depends_on = @(Parse-InlineList $FrontMatter "depends_on")
        impacts = @(Parse-InlineList $FrontMatter "impacts")
        read_when = @(Parse-InlineList $FrontMatter "read_when")
        update_when = @(Parse-InlineList $FrontMatter "update_when")
        conflict_policy = Parse-SingleValue $FrontMatter "conflict_policy"
        rollback_target = @(Parse-InlineList $FrontMatter "rollback_target")
        verification_target = @(Parse-InlineList $FrontMatter "verification_target")
        last_updated = $LastWriteTimeUtc.ToString("o")
    }

    foreach ($optionalListField in @("supersedes", "superseded_by")) {
        if ($FrontMatter -match "(?m)^$([regex]::Escape($optionalListField))\s*:") {
            $state[$optionalListField] = @(Parse-InlineList $FrontMatter $optionalListField)
        }
    }

    foreach ($optionalScalarField in @("task_entrypoint", "change_reason")) {
        if ($FrontMatter -match "(?m)^$([regex]::Escape($optionalScalarField))\s*:") {
            if ($optionalScalarField -eq "task_entrypoint") {
                $state[$optionalScalarField] = Parse-Boolean $FrontMatter $optionalScalarField
            } else {
                $state[$optionalScalarField] = Parse-SingleValue $FrontMatter $optionalScalarField
            }
        }
    }

    return [PSCustomObject]$state
}

$states = foreach ($file in Get-GovernanceFiles $Root) {
    $content = Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8
    $frontMatter = Get-FrontMatter $content
    if ($null -eq $frontMatter) {
        continue
    }

    $relativePath = Get-RelativePath $Root $file.FullName
    Build-StateObject $relativePath $content $frontMatter $file.LastWriteTimeUtc
}

$outputAbsolutePath = Join-Path $Root $OutputPath
$outputDirectory = Split-Path -Parent $outputAbsolutePath
if (-not (Test-Path -LiteralPath $outputDirectory -PathType Container)) {
    New-Item -ItemType Directory -Path $outputDirectory | Out-Null
}

$metadata = [ordered]@{
    state_id = "IDX-GOVMAP-001"
    title = "Keiyaku-Go Governance Map"
    file_path = $OutputPath.Replace("\", "/")
    doc_role = "governance_map"
    memory_level = "L0"
    state_scope = "global"
    scope = "repo"
    authority_level = "derived"
    owners = @("tech-lead")
    status = "active"
    version = "2.0"
    source_of_truth = @("docs/governance/README.md")
    derived_from = @(
        "docs/governance/README.md",
        "docs/governance/rules.md",
        "docs/governance/change-management.md",
        "docs/governance/metadata-schema.md"
    )
    read_when = @("all_tasks", "governance_change")
    update_when = @("routing_changed", "governance_structure_changed", "metadata_standard_changed", "automation_changed")
    conflict_policy = "index_must_yield_to_ssot"
    rollback_target = @("docs/governance/README.md", "docs/governance/metadata-schema.md")
    verification_target = @("scripts/check-governance.ps1", "scripts/check-governance-map.ps1")
    generated_at = (Get-Date).ToUniversalTime().ToString("o")
}

$document = [ordered]@{
    metadata = $metadata
    states = @($states)
}

$json = $document | ConvertTo-Json -Depth 10
Set-Content -LiteralPath $outputAbsolutePath -Value $json -Encoding UTF8

Write-Host "Governance map exported to $OutputPath" -ForegroundColor Green
exit 0
