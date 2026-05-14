param(
    [string]$Root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

$ErrorActionPreference = "Stop"
$failures = New-Object System.Collections.Generic.List[string]
$today = (Get-Date).Date

function Add-Failure {
    param([string]$Message)
    $failures.Add($Message) | Out-Null
}

$exceptionsPath = Join-Path $Root "docs/governance/exceptions.yaml"
if (-not (Test-Path -LiteralPath $exceptionsPath -PathType Leaf)) {
    Add-Failure "Missing exceptions registry: docs/governance/exceptions.yaml"
} else {
    $content = Get-Content -LiteralPath $exceptionsPath -Raw -Encoding UTF8
    $blocks = [regex]::Split($content, "(?m)^\s*-\s+id:\s+")

    foreach ($block in $blocks | Select-Object -Skip 1) {
        $id = ($block -split "\r?\n", 2)[0].Trim()
        $kind = if ($block -match "(?m)^\s+kind:\s+(.+)$") { $Matches[1].Trim() } else { "" }
        $status = if ($block -match "(?m)^\s+status:\s+(.+)$") { $Matches[1].Trim() } else { "" }
        $expiry = if ($block -match "(?m)^\s+expiry:\s+(.+)$") { $Matches[1].Trim() } else { "" }
        $reviewAt = if ($block -match "(?m)^\s+review_at:\s+(.+)$") { $Matches[1].Trim() } else { "" }
        $adrRef = if ($block -match "(?m)^\s+adr_ref:\s+(.+)$") { $Matches[1].Trim() } else { "" }

        if ($status -ne "active") {
            continue
        }

        if ($kind -eq "break_glass") {
            if ([string]::IsNullOrWhiteSpace($expiry) -or $expiry -eq "null") {
                Add-Failure "Active break-glass entry must have expiry: $id"
            }
            if ([string]::IsNullOrWhiteSpace($adrRef) -or $adrRef -eq "null") {
                Add-Failure "Active break-glass entry must explain ADR status in adr_ref: $id"
            }
        }

        foreach ($dateField in @(@("expiry", $expiry), @("review_at", $reviewAt))) {
            $name = $dateField[0]
            $value = $dateField[1]
            if ([string]::IsNullOrWhiteSpace($value) -or $value -eq "null") {
                continue
            }
            try {
                $dateValue = [DateTime]::Parse($value).Date
                if ($dateValue -lt $today) {
                    Add-Failure "Active exception has expired $name date: $id ($value)"
                }
            } catch {
                Add-Failure "Invalid $name date on exception $id`: $value"
            }
        }
    }
}

if ($failures.Count -gt 0) {
    Write-Host "Exception expiry check failed:" -ForegroundColor Red
    foreach ($failure in $failures) {
        Write-Host " - $failure" -ForegroundColor Red
    }
    exit 1
}

Write-Host "Exception expiry check passed." -ForegroundColor Green
exit 0
