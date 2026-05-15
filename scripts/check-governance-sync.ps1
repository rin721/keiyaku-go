param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
}

function Read-Text {
    param([string]$RelativePath)

    $path = Join-Path $Root $RelativePath
    if (-not (Test-Path -LiteralPath $path -PathType Leaf)) {
        Add-Failure "Missing sync-check dependency: $RelativePath"
        return ""
    }
    return Get-Content -LiteralPath $path -Raw -Encoding UTF8
}

function Require-Contains {
    param(
        [string]$RelativePath,
        [string[]]$Patterns
    )

    $content = Read-Text $RelativePath
    foreach ($pattern in $Patterns) {
        if ($content -notmatch [regex]::Escape($pattern)) {
            Add-Failure "$RelativePath is missing synchronized reference: $pattern"
        }
    }
}

function Get-FrontMatter {
    param([string]$Content)

    if ($Content -notmatch "(?s)^---\r?\n(.*?)\r?\n---") {
        return $null
    }
    return $Matches[1]
}

Require-Contains "AGENTS.md" @(
    "docs/governance/metadata-schema.md",
    "docs/governance/change-management.md",
    "docs/governance/governance-map.json"
)

Require-Contains "docs/governance/README.md" @(
    "metadata-schema.md",
    "change-management.md",
    "automation-matrix.md",
    "exceptions.yaml",
    "exceptions.template.yaml",
    "governance-map.json",
    "20260515-governance-state-model.md"
)

Require-Contains "docs/governance/ai-execution.md" @(
    "metadata-schema.md",
    "automation-matrix.md",
    "exceptions.yaml",
    "governance-map.json"
)

Require-Contains "docs/governance/automation-matrix.md" @(
    "check-governance-taxonomy.ps1",
    "check-governance-sync.ps1",
    "dependency-injection.md",
    "check-governance-map.ps1",
    "export-governance-map.ps1",
    "governance-map.json"
)

Require-Contains "docs/governance/metadata-schema.md" @(
    "governance_map",
    "memory_level",
    "state_scope",
    "source_of_truth",
    "rollback_target",
    "verification_target"
)

Require-Contains "docs/review/governance-change-checklist.md" @(
    "rules.md",
    "ai-execution.md",
    "metadata-schema.md",
    "automation-matrix.md",
    "exceptions.yaml",
    "stop-condition",
    "governance-map.json"
)

Require-Contains "docs/adr/README.md" @(
    "accepted",
    "ssot_decision"
)

foreach ($adrPath in @(
    "docs/adr/20260515-governance-ssot-structure.md",
    "docs/adr/20260515-default-backend-direction.md",
    "docs/adr/20260515-governance-state-model.md"
)) {
    $content = Read-Text $adrPath
    $frontMatter = Get-FrontMatter $content
    if ($frontMatter -and $frontMatter -notmatch "(?m)^status:\s+accepted\s*$") {
        Add-Failure "Accepted ADR has the wrong status: $adrPath"
    }
}

Require-Contains "docs/adr/0000-template.md" @(
    "governance-map.json",
    "automation-matrix.md"
)

$historicalDoc = Read-Text "docs/architecture/governance.md"
$historicalFrontMatter = Get-FrontMatter $historicalDoc
if ($historicalFrontMatter -and $historicalFrontMatter -notmatch "(?m)^status:\s+historical\s*$") {
    Add-Failure "Historical governance document is not marked historical: docs/architecture/governance.md"
}
if ($historicalDoc -match "## P0" -or $historicalDoc -match "## P1") {
    Add-Failure "Historical governance document still contains active rule sections: docs/architecture/governance.md"
}

$exceptions = Read-Text "docs/governance/exceptions.yaml"
if ($exceptions -match "status:\s+example") {
    Add-Failure "Production exception registry must not keep example entries: docs/governance/exceptions.yaml"
}

if (-not (Test-Path -LiteralPath (Join-Path $Root "docs/governance/governance-map.json") -PathType Leaf)) {
    Add-Failure "Missing governance map artifact: docs/governance/governance-map.json"
}

if ($failures.Count -gt 0) {
    Write-Host "Governance sync check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Governance sync check passed." -ForegroundColor Green
exit 0
