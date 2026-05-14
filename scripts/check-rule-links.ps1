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

$rulesPath = Join-Path $Root "docs/governance/rules.md"
if (-not (Test-Path -LiteralPath $rulesPath -PathType Leaf)) {
    Add-Failure "缺少规则真相：docs/governance/rules.md"
} else {
    $rulesContent = Get-Content -LiteralPath $rulesPath -Raw -Encoding UTF8
    $rulesFrontMatter = Get-FrontMatter $rulesContent
    $definedRuleIds = [regex]::Matches($rulesContent, "GOV-P[0-2]-[0-9]{3}") | ForEach-Object { $_.Value } | Sort-Object -Unique
    $relatedRulesInRules = Parse-InlineList $rulesFrontMatter "related_rules"

    foreach ($ruleId in $definedRuleIds) {
        if ($ruleId -notin $relatedRulesInRules) {
            Add-Failure "rules.md 的 related_rules 未覆盖规则：$ruleId"
        }
    }

    $docsPath = Join-Path $Root "docs"
    $docFiles = Get-ChildItem -LiteralPath $docsPath -Recurse -File -Include "*.md", "*.yaml", "*.yml"
    foreach ($file in $docFiles) {
        $content = Get-Content -LiteralPath $file.FullName -Raw -Encoding UTF8
        $frontMatter = Get-FrontMatter $content
        if ($null -eq $frontMatter) {
            continue
        }

        $relatedRules = Parse-InlineList $frontMatter "related_rules"
        $bodyRuleRefs = [regex]::Matches($content, "GOV-P[0-2]-[0-9]{3}") | ForEach-Object { $_.Value } | Sort-Object -Unique

        foreach ($ruleId in $relatedRules) {
            if ($ruleId -match "^GOV-P[0-2]-" -and $ruleId -notin $definedRuleIds) {
                Add-Failure "文档引用了未定义规则：$($file.FullName) -> $ruleId"
            }
        }

        foreach ($ruleId in $bodyRuleRefs) {
            if ($ruleId -notin $definedRuleIds) {
                Add-Failure "文档正文引用了未定义规则：$($file.FullName) -> $ruleId"
            }
            if ($ruleId -notin $relatedRules) {
                Add-Failure "文档正文引用了规则但 related_rules 未覆盖：$($file.FullName) -> $ruleId"
            }
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
