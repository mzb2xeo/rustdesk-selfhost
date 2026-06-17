# Smoke test for RustDesk host after deploy (Windows)
# Usage:
#   .\test-rustdesk-host.ps1
#   .\test-rustdesk-host.ps1 -ExpectedHost rd.example.com -Password "Rd@86079"

param(
    [string]$RustDeskExe = "C:\Program Files\RustDesk\rustdesk.exe",
    [string]$ExpectedHost = "",
    [string]$ExpectedApi = "",
    [string]$Password = ""
)

$ErrorActionPreference = "Continue"
$passed = 0
$failed = 0

function Write-Check {
    param([string]$Name, [bool]$Ok, [string]$Detail = "")
    if ($Ok) {
        Write-Host "[PASS] $Name" -ForegroundColor Green
        if ($Detail) { Write-Host "       $Detail" -ForegroundColor DarkGray }
        $script:passed++
    } else {
        Write-Host "[FAIL] $Name" -ForegroundColor Red
        if ($Detail) { Write-Host "       $Detail" -ForegroundColor Yellow }
        $script:failed++
    }
}

Write-Host "RustDesk host smoke test" -ForegroundColor Cyan
Write-Host "Exe: $RustDeskExe"
Write-Host ""

Write-Check "RustDesk installed" (Test-Path $RustDeskExe) $RustDeskExe

$svc = Get-Service -Name "rustdesk" -ErrorAction SilentlyContinue
Write-Check "Windows service exists" ($null -ne $svc) $(if ($svc) { "Status=$($svc.Status)" } else { "Run: rustdesk --install-service (Admin)" })

$pipeReady = Test-Path "\\.\pipe\RustDesk\query"
Write-Check "IPC pipe ready" $pipeReady "\\.\pipe\RustDesk\query"

if (Test-Path $RustDeskExe) {
    $id = (& $RustDeskExe --get-id 2>&1 | Out-String).Trim()
    if ($id -match '(\d{9,})') { $id = $Matches[1] }
    Write-Check "Device ID readable" ($id -match '^\d{9,}$') $(if ($id) { "ID=$id" } else { "Start service or rustdesk --server" })
}

$configPaths = @(
    "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk2.toml",
    "$env:APPDATA\RustDesk\config\RustDesk2.toml"
)
$configFound = $false
foreach ($toml in $configPaths) {
    if (-not (Test-Path -LiteralPath $toml -ErrorAction SilentlyContinue)) { continue }
    $configFound = $true
    $content = Get-Content -LiteralPath $toml -Raw
    Write-Check "Config file exists" $true $toml
    if ($ExpectedHost) {
        $hostOk = $content -match "custom-rendezvous-server\s*=\s*'$([regex]::Escape($ExpectedHost))'"
        Write-Check "ID/Relay host" $hostOk "expected=$ExpectedHost"
    }
    if ($ExpectedApi) {
        $api = $ExpectedApi.TrimEnd('/')
        $apiOk = $content -match "api-server\s*=\s*'$([regex]::Escape($api))'"
        Write-Check "API server" $apiOk "expected=$api"
    }
    break
}
if (-not $configFound) {
    Write-Check "Config file exists" $false "RustDesk2.toml not found"
}

$localToml = "$env:APPDATA\RustDesk\config\RustDesk_local.toml"
if (Test-Path $localToml) {
    $local = Get-Content $localToml -Raw
    $hasToken = $local -match "access_token\s*=\s*'[^']+'"
    Write-Check "Client logged in" $hasToken $localToml
} else {
    Write-Check "Client logged in" $false "RustDesk_local.toml missing"
}

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
Write-Check "Running as Administrator" $isAdmin "Required for --password"

if ($Password -and (Test-Path $RustDeskExe) -and $isAdmin) {
    if (-not $pipeReady) {
        Write-Host "       Starting rustdesk --server for password test..." -ForegroundColor Yellow
        Start-Process -FilePath $RustDeskExe -ArgumentList "--server" -WindowStyle Hidden | Out-Null
        Start-Sleep -Seconds 8
        $pipeReady = Test-Path "\\.\pipe\RustDesk\query"
    }
    if ($pipeReady) {
        $pwdOut = (& $RustDeskExe --password "$Password" 2>&1 | Out-String).Trim()
        Write-Check "Permanent password" ($pwdOut -match 'Done') $pwdOut
    } else {
        Write-Check "Permanent password" $false "IPC pipe not ready"
    }
} elseif ($Password) {
    Write-Check "Permanent password" $false "Skipped (need Admin + Password)"
}

Write-Host ""
Write-Host "Result: $passed passed, $failed failed" -ForegroundColor $(if ($failed -eq 0) { "Green" } else { "Yellow" })
if ($failed -gt 0) { exit 1 }
exit 0
