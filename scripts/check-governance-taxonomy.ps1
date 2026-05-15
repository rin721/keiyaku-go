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

$schemaPath = Join-Path $Root "docs/governance/metadata-schema.md"
if (-not (Test-Path -LiteralPath $schemaPath -PathType Leaf)) {
    Add-Failure "Missing metadata schema: docs/governance/metadata-schema.md"
} else {
    $schemaContent = Get-Content -LiteralPath $schemaPath -Raw -Encoding UTF8
    $allowedDocRoles = Get-AllowedValues $schemaContent "DOC_ROLE"
    $allowedMemoryLevels = Get-AllowedValues $schemaContent "MEMORY_LEVEL"
    $allowedStateScopes = Get-AllowedValues $schemaContent "STATE_SCOPE"
    $allowedScopes = Get-AllowedValues $schemaContent "SCOPE"
    $allowedAuthorities = Get-AllowedValues $schemaContent "AUTHORITY_LEVEL"
    $allowedStatuses = Get-AllowedValues $schemaContent "STATUS"
    $allowedReadWhen = Get-AllowedValues $schemaContent "READ_WHEN"
    $allowedUpdateWhen = Get-AllowedValues $schemaContent "UPDATE_WHEN"

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

    foreach ($file in $files) {
        $relativePath = Get-RelativePath $Root $file.FullName
        $content = Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8
        $frontMatter = Get-FrontMatter $content
        if ($null -eq $frontMatter) {
            continue
        }

        $docRole = Parse-SingleValue $frontMatter "doc_role"
        $memoryLevel = Parse-SingleValue $frontMatter "memory_level"
        $stateScope = Parse-SingleValue $frontMatter "state_scope"
        $scope = Parse-SingleValue $frontMatter "scope"
        $authority = Parse-SingleValue $frontMatter "authority_level"
        $status = Parse-SingleValue $frontMatter "status"
        $readWhen = Parse-InlineList $frontMatter "read_when"
        $updateWhen = Parse-InlineList $frontMatter "update_when"
        $taskEntrypoint = Parse-SingleValue $frontMatter "task_entrypoint"

        if ($docRole -notin $allowedDocRoles) {
            Add-Failure "Invalid doc_role: $relativePath -> $docRole"
        }
        if ($memoryLevel -notin $allowedMemoryLevels) {
            Add-Failure "Invalid memory_level: $relativePath -> $memoryLevel"
        }
        if ($stateScope -notin $allowedStateScopes) {
            Add-Failure "Invalid state_scope: $relativePath -> $stateScope"
        }
        if ($scope -notin $allowedScopes) {
            Add-Failure "Invalid scope: $relativePath -> $scope"
        }
        if ($authority -notin $allowedAuthorities) {
            Add-Failure "Invalid authority_level: $relativePath -> $authority"
        }
        if ($status -notin $allowedStatuses) {
            Add-Failure "Invalid status: $relativePath -> $status"
        }
        foreach ($item in $readWhen) {
            if ($item -notin $allowedReadWhen) {
                Add-Failure "Invalid read_when tag: $relativePath -> $item"
            }
        }
        foreach ($item in $updateWhen) {
            if ($item -notin $allowedUpdateWhen) {
                Add-Failure "Invalid update_when tag: $relativePath -> $item"
            }
        }

        switch ($docRole) {
            "ai_entry" {
                if ($authority -ne "entry") {
                    Add-Failure "AI entry documents must use authority_level: entry -> $relativePath"
                }
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "AI entry documents must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "navigation" {
                if ($authority -ne "ssot_navigation") {
                    Add-Failure "Navigation SSOT must use authority_level: ssot_navigation -> $relativePath"
                }
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "Navigation SSOT must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "governance_rules" {
                if ($authority -ne "ssot_rules") {
                    Add-Failure "Rules SSOT must use authority_level: ssot_rules -> $relativePath"
                }
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "Rules SSOT must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "ai_execution" {
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "AI execution documents must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "governance_process" {
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "Governance process documents must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "automation_spec" {
                if ($stateScope -ne "global") {
                    Add-Failure "Automation specs must use state_scope global -> $relativePath"
                }
            }
            "metadata_schema" {
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "Metadata schema documents must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "adr_index" {
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "ADR index documents must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "adr" {
                if ($authority -ne "ssot_decision") {
                    Add-Failure "ADR decision records must use authority_level: ssot_decision -> $relativePath"
                }
                if ($memoryLevel -ne "L0" -or $stateScope -ne "global") {
                    Add-Failure "ADR decision records must use memory_level L0 and state_scope global -> $relativePath"
                }
            }
            "template" {
                if ($authority -ne "template") {
                    Add-Failure "Template documents must use authority_level: template -> $relativePath"
                }
            }
            "convention" {
                if ($memoryLevel -ne "L1" -or $stateScope -ne "module") {
                    Add-Failure "Convention documents must use memory_level L1 and state_scope module -> $relativePath"
                }
            }
            "review_checklist" {
                if ($memoryLevel -ne "L1" -or $stateScope -ne "module") {
                    Add-Failure "Review checklists must use memory_level L1 and state_scope module -> $relativePath"
                }
            }
            "exception_registry" {
                if ($stateScope -ne "global") {
                    Add-Failure "Exception registries must use state_scope global -> $relativePath"
                }
            }
            "historical_reference" {
                if ($status -notin @("historical", "deprecated")) {
                    Add-Failure "Historical references must use status historical or deprecated -> $relativePath"
                }
            }
        }

        if ($authority -eq "ssot_decision" -and $docRole -ne "adr") {
            Add-Failure "Only ADR documents may use authority_level: ssot_decision -> $relativePath"
        }
        if (-not [string]::IsNullOrWhiteSpace($taskEntrypoint) -and $taskEntrypoint -notin @("true", "false")) {
            Add-Failure "task_entrypoint must be 'true' or 'false': $relativePath"
        }
        if ($taskEntrypoint -eq "true" -and $docRole -ne "ai_entry") {
            Add-Failure "Only ai_entry documents may set task_entrypoint: true -> $relativePath"
        }
        if ($relativePath -eq "docs/architecture/governance.md" -and $status -ne "historical") {
            Add-Failure "Historical governance document must use status historical -> $relativePath"
        }
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Governance taxonomy check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Governance taxonomy check passed." -ForegroundColor Green
exit 0
