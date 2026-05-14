param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
}

$rulesPath = Join-Path $Root "docs/governance/rules.md"
if (-not (Test-Path -LiteralPath $rulesPath -PathType Leaf)) {
    Add-Failure "Missing rules SSOT: docs/governance/rules.md"
} else {
    $rules = Get-Content -LiteralPath $rulesPath -Raw -Encoding UTF8
    $requiredRules = @(
        "GOV-P0-001",
        "GOV-P0-002",
        "GOV-P0-003",
        "GOV-P0-004",
        "GOV-P1-001",
        "GOV-P1-002",
        "GOV-P1-003",
        "GOV-P1-004",
        "GOV-P1-005",
        "GOV-P1-006"
    )
    foreach ($rule in $requiredRules) {
        if ($rules -notmatch [regex]::Escape($rule)) {
            Add-Failure "Missing rule id in rules.md: $rule"
        }
    }

    $docsPath = Join-Path $Root "docs"
    $docFiles = Get-ChildItem -LiteralPath $docsPath -Recurse -File -Include "*.md", "*.yaml", "*.yml"
    $allRefs = New-Object System.Collections.Generic.HashSet[string]
    foreach ($file in $docFiles) {
        $content = Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8
        foreach ($match in [regex]::Matches($content, "GOV-(P[0-2]|NAV|META|CI)-[0-9]{3}")) {
            $allRefs.Add($match.Value) | Out-Null
        }
    }

    foreach ($ref in $allRefs) {
        if ($ref -match "^GOV-P[0-2]-" -and $rules -notmatch [regex]::Escape($ref)) {
            Add-Failure "Referenced rule id is not defined in rules.md: $ref"
        }
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Rule link check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Rule link check passed." -ForegroundColor Green
exit 0
