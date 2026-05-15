param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
}

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

    Add-Failure "Invalid boolean value for ${FieldName}: $value"
    return $null
}

function Normalize-List {
    param($Values)

    if ($null -eq $Values) {
        return @()
    }

    $items = @()
    foreach ($value in $Values) {
        $items += [string]$value
    }
    return $items
}

function Get-AllowedValues {
    param(
        [string]$SchemaContent,
        [string]$SectionName
    )

    $pattern = "(?s)<!-- META-$([regex]::Escape($SectionName))-START -->(.*?)<!-- META-$([regex]::Escape($SectionName))-END -->"
    if ($SchemaContent -notmatch $pattern) {
        throw "Missing schema markers for $SectionName"
    }

    return [regex]::Matches($Matches[1], '`([^`]+)`') | ForEach-Object { $_.Groups[1].Value } | Sort-Object -Unique
}

function Test-RelativeTargetsExist {
    param(
        [string[]]$Targets,
        [string]$Label,
        [string]$StateId
    )

    foreach ($target in $Targets) {
        if ([string]::IsNullOrWhiteSpace($target)) {
            continue
        }

        $absolute = Join-Path $Root $target
        if (-not (Test-Path -LiteralPath $absolute)) {
            Add-Failure "$StateId has missing $Label target: $target"
        }
    }
}

$schemaPath = Join-Path $Root "docs/governance/metadata-schema.md"
$schemaContent = Get-Content -LiteralPath $schemaPath -Raw -Encoding UTF8
$allowedDocRoles = Get-AllowedValues $schemaContent "DOC_ROLE"
$allowedMemoryLevels = Get-AllowedValues $schemaContent "MEMORY_LEVEL"
$allowedStateScopes = Get-AllowedValues $schemaContent "STATE_SCOPE"
$allowedAuthorities = Get-AllowedValues $schemaContent "AUTHORITY_LEVEL"
$allowedStatuses = Get-AllowedValues $schemaContent "STATUS"

$mapPath = Join-Path $Root "docs/governance/governance-map.json"
if (-not (Test-Path -LiteralPath $mapPath -PathType Leaf)) {
    Add-Failure "Missing governance map: docs/governance/governance-map.json"
} else {
    $map = Get-Content -LiteralPath $mapPath -Raw -Encoding UTF8 | ConvertFrom-Json

    foreach ($field in @("state_id", "title", "file_path", "doc_role", "memory_level", "state_scope", "scope", "authority_level", "owners", "status", "source_of_truth", "derived_from", "read_when", "update_when", "conflict_policy", "rollback_target", "verification_target", "generated_at")) {
        if (-not $map.metadata.PSObject.Properties.Name.Contains($field)) {
            Add-Failure "governance-map.json metadata is missing field: $field"
        }
    }

    if ($map.metadata.doc_role -notin $allowedDocRoles) {
        Add-Failure "governance-map.json metadata has invalid doc_role: $($map.metadata.doc_role)"
    }
    if ($map.metadata.memory_level -notin $allowedMemoryLevels) {
        Add-Failure "governance-map.json metadata has invalid memory_level: $($map.metadata.memory_level)"
    }
    if ($map.metadata.state_scope -notin $allowedStateScopes) {
        Add-Failure "governance-map.json metadata has invalid state_scope: $($map.metadata.state_scope)"
    }
    if ($map.metadata.authority_level -notin $allowedAuthorities) {
        Add-Failure "governance-map.json metadata has invalid authority_level: $($map.metadata.authority_level)"
    }
    if ($map.metadata.status -notin $allowedStatuses) {
        Add-Failure "governance-map.json metadata has invalid status: $($map.metadata.status)"
    }

    Test-RelativeTargetsExist (Normalize-List $map.metadata.source_of_truth) "source_of_truth" $map.metadata.state_id
    Test-RelativeTargetsExist (Normalize-List $map.metadata.derived_from) "derived_from" $map.metadata.state_id
    Test-RelativeTargetsExist (Normalize-List $map.metadata.rollback_target) "rollback_target" $map.metadata.state_id
    Test-RelativeTargetsExist (Normalize-List $map.metadata.verification_target) "verification_target" $map.metadata.state_id

    $statesById = @{}
    foreach ($state in $map.states) {
        if ($statesById.ContainsKey($state.state_id)) {
            Add-Failure "Duplicate state_id in governance-map.json: $($state.state_id)"
            continue
        }
        $statesById[$state.state_id] = $state
    }

    $documentStates = @{}
    $taskEntryPoints = @()

    foreach ($file in Get-GovernanceFiles $Root) {
        $content = Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8
        $frontMatter = Get-FrontMatter $content
        if ($null -eq $frontMatter) {
            continue
        }

        $relativePath = Get-RelativePath $Root $file.FullName
        $stateId = Parse-SingleValue $frontMatter "state_id"

        if ([string]::IsNullOrWhiteSpace($stateId)) {
            continue
        }

        if ($documentStates.ContainsKey($stateId)) {
            Add-Failure "Duplicate state_id in governance documents: $stateId"
            continue
        }

        $expected = [ordered]@{
            file_path = $relativePath
            doc_role = Parse-SingleValue $frontMatter "doc_role"
            memory_level = Parse-SingleValue $frontMatter "memory_level"
            state_scope = Parse-SingleValue $frontMatter "state_scope"
            authority_level = Parse-SingleValue $frontMatter "authority_level"
            status = Parse-SingleValue $frontMatter "status"
            source_of_truth = @(Parse-InlineList $frontMatter "source_of_truth")
            derived_from = @(Parse-InlineList $frontMatter "derived_from")
            read_when = @(Parse-InlineList $frontMatter "read_when")
            update_when = @(Parse-InlineList $frontMatter "update_when")
            conflict_policy = Parse-SingleValue $frontMatter "conflict_policy"
            rollback_target = @(Parse-InlineList $frontMatter "rollback_target")
            verification_target = @(Parse-InlineList $frontMatter "verification_target")
            last_updated = $file.LastWriteTimeUtc.ToString("o")
        }

        $documentStates[$stateId] = $expected

        $taskEntrypoint = Parse-Boolean $frontMatter "task_entrypoint"
        if ($taskEntrypoint -eq $true) {
            $taskEntryPoints += $stateId
        }

        if (-not $statesById.ContainsKey($stateId)) {
            Add-Failure "governance-map.json is missing state entry: $stateId"
            continue
        }

        $indexedState = $statesById[$stateId]
        foreach ($field in $expected.Keys) {
            $expectedValue = $expected[$field]
            $actualValue = $indexedState.$field

            if ($expectedValue -is [System.Array]) {
                $actualList = Normalize-List $actualValue
                if (($expectedValue -join "|") -ne ($actualList -join "|")) {
                    Add-Failure "$stateId has stale governance-map field '$field'"
                }
            } else {
                if ([string]$expectedValue -ne [string]$actualValue) {
                    Add-Failure "$stateId has stale governance-map field '$field'"
                }
            }
        }

        Test-RelativeTargetsExist $expected.source_of_truth "source_of_truth" $stateId
        Test-RelativeTargetsExist $expected.derived_from "derived_from" $stateId
        Test-RelativeTargetsExist $expected.rollback_target "rollback_target" $stateId
        Test-RelativeTargetsExist $expected.verification_target "verification_target" $stateId
    }

    if ($taskEntryPoints.Count -ne 1) {
        Add-Failure "Exactly one governance document must set task_entrypoint: true"
    }

    foreach ($indexedStateId in $statesById.Keys) {
        if (-not $documentStates.ContainsKey($indexedStateId)) {
            Add-Failure "governance-map.json contains orphan state entry: $indexedStateId"
        }
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Governance map check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Governance map check passed." -ForegroundColor Green
exit 0
