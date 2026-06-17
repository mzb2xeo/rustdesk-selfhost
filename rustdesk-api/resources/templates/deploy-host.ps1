# RustDesk automated host deployment (Windows)

param (
    [string]$DeployToken = "{{.DeployToken}}",
    [string]$ApiUrl = "{{.ApiUrl}}",
    [string]$IdServer = "{{.IdServer}}",
    [string]$RelayServer = "{{.RelayServer}}",
    [string]$Key = "{{.Key}}",
    [string]$ConfigString = "{{.ConfigString}}",
    [string]$PasswordMode = "{{.PasswordMode}}",
    [string]$CustomPassword = "{{.CustomPassword}}"
)

$ErrorActionPreference = "Stop"
$logPath = Join-Path $env:TEMP "rustdesk-deploy.log"
function Write-DeployLog($message) {
    $line = "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $message"
    Add-Content -Path $logPath -Value $line -Encoding UTF8
}

function Stop-RustDeskRuntime {
    Write-DeployLog "Stopping RustDesk service and processes."
    Stop-Service -Name "rustdesk" -ErrorAction SilentlyContinue
    Get-Process -Name "rustdesk" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
}

function Start-RustDeskRuntime {
    Write-DeployLog "Starting RustDesk service."
    Start-Service -Name "rustdesk" -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 5
}

function Get-RustDeskDeviceId {
    param([string]$Exe)

    # Windows GUI build often does not print to the interactive console; capture via Out-String.
    try {
        $out = (& $Exe --get-id 2>&1 | Out-String).Trim()
        if ($out -match '(\d{9,})') {
            return $Matches[1]
        }
    } catch {}

    # Fallback: redirect stdout to a temp file (works in cmd-style capture).
    $tmp = Join-Path $env:TEMP "rustdesk-get-id.txt"
    try {
        cmd /c "`"$Exe`" --get-id > `"$tmp`" 2>&1"
        if (Test-Path $tmp) {
            $fileOut = (Get-Content $tmp -Raw -ErrorAction SilentlyContinue)
            if ($fileOut -match '(\d{9,})') {
                return $Matches[1]
            }
        }
    } finally {
        Remove-Item $tmp -ErrorAction SilentlyContinue
    }

    # Fallback: newer builds may copy ID to clipboard.
    try {
        & $Exe --get-id 2>&1 | Out-Null
        Start-Sleep -Milliseconds 800
        $clip = Get-Clipboard -ErrorAction SilentlyContinue
        if ($clip -and "$clip" -match '^\d{9,}$') {
            return "$clip".Trim()
        }
    } catch {}

    # Fallback: read plaintext id from local config if present.
    $configPaths = @(
        "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk.toml",
        "C:\Windows\System32\config\systemprofile\AppData\Roaming\RustDesk\config\RustDesk.toml",
        "$env:APPDATA\RustDesk\config\RustDesk.toml"
    )
    foreach ($tomlPath in $configPaths) {
        if (-not (Test-Path $tomlPath)) { continue }
        $content = Get-Content $tomlPath -Raw -ErrorAction SilentlyContinue
        if ($content -match "id\s*=\s*'(\d+)'") { return $Matches[1] }
        if ($content -match 'id\s*=\s*"(\d+)"') { return $Matches[1] }
    }

    return ""
}

function Normalize-RustDeskHost {
    param([string]$Value)
    $v = $Value.Trim()
    if ($v -match '^(.+):21116$') { return $Matches[1] }
    if ($v -match '^(.+):21117$') { return $Matches[1] }
    if ($v -match '^([^:]+):\d+$') { return $Matches[1] }
    return $v
}

function Get-RustDeskTomlValue {
    param(
        [string]$Content,
        [string]$Name
    )
    if ($Content -match "(?m)^\s*$([regex]::Escape($Name))\s*=\s*'([^']*)'") {
        return $Matches[1]
    }
    if ($Content -match "(?m)^\s*$([regex]::Escape($Name))\s*=\s*`"([^`"]*)`"") {
        return $Matches[1]
    }
    return ""
}

function Test-RustDeskTomlContent {
    param(
        [string]$Content,
        [string]$ExpectedHost,
        [string]$ExpectedRelay,
        [string]$ExpectedApi,
        [string]$ExpectedKey
    )

    $actualHost = Get-RustDeskTomlValue -Content $Content -Name "custom-rendezvous-server"
    $actualRelay = Get-RustDeskTomlValue -Content $Content -Name "relay-server"
    $actualApi = (Get-RustDeskTomlValue -Content $Content -Name "api-server").TrimEnd('/')
    $actualKey = Get-RustDeskTomlValue -Content $Content -Name "key"
    $actualRendezvous = Get-RustDeskTomlValue -Content $Content -Name "rendezvous_server"

    $hostOk = ($actualHost -eq $ExpectedHost)
    $relayOk = ($actualRelay -eq $ExpectedRelay)
    $apiOk = ($actualApi -eq $ExpectedApi)
    $keyOk = ($actualKey -eq $ExpectedKey)
    $rendezvousOk = ($actualRendezvous -eq "$ExpectedHost`:21116") -or ($actualRendezvous -eq $ExpectedHost)

    return [PSCustomObject]@{
        Passed = ($hostOk -and $relayOk -and $apiOk -and $keyOk)
        HostOk = $hostOk
        RelayOk = $relayOk
        ApiOk = $apiOk
        KeyOk = $keyOk
        RendezvousOk = $rendezvousOk
        ActualHost = $actualHost
        ActualRelay = $actualRelay
        ActualApi = $actualApi
        ActualKey = $actualKey
        ActualRendezvous = $actualRendezvous
    }
}

function Get-RustDeskConfigPaths {
    return @(
        "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk2.toml",
        "C:\Windows\System32\config\systemprofile\AppData\Roaming\RustDesk\config\RustDesk2.toml",
        "$env:APPDATA\RustDesk\config\RustDesk2.toml"
    )
}

function Test-RustDeskConfigApplied {
    param(
        [string]$ExpectedHost,
        [string]$ExpectedRelay,
        [string]$ExpectedApi,
        [string]$ExpectedKey
    )

    $reports = @()
    foreach ($tomlPath in (Get-RustDeskConfigPaths)) {
        if (-not (Test-Path -LiteralPath $tomlPath -ErrorAction SilentlyContinue)) {
            $reports += [PSCustomObject]@{ Path = $tomlPath; Passed = $false; Reason = "missing" }
            continue
        }
        try {
            $content = Get-Content -LiteralPath $tomlPath -Raw -ErrorAction Stop
        } catch {
            $reports += [PSCustomObject]@{ Path = $tomlPath; Passed = $false; Reason = "unreadable: $($_.Exception.Message)" }
            continue
        }
        $result = Test-RustDeskTomlContent -Content $content -ExpectedHost $ExpectedHost -ExpectedRelay $ExpectedRelay -ExpectedApi $ExpectedApi -ExpectedKey $ExpectedKey
        $reports += [PSCustomObject]@{
            Path = $tomlPath
            Passed = $result.Passed
            Reason = if ($result.Passed) { "ok" } else { "mismatch" }
            Details = $result
        }
        if ($result.Passed) {
            return [PSCustomObject]@{
                Passed = $true
                Path = $tomlPath
                Details = $result
                Reports = $reports
            }
        }
    }

    return [PSCustomObject]@{
        Passed = $false
        Path = ""
        Details = $null
        Reports = $reports
    }
}

function Test-RustDeskNetwork {
    param(
        [string]$HostName,
        [string]$ApiUrl
    )

    $checks = @()
    foreach ($port in @(21116, 21117)) {
        $ok = $false
        try {
            $ok = (Test-NetConnection -ComputerName $HostName -Port $port -WarningAction SilentlyContinue).TcpTestSucceeded
        } catch {}
        $checks += [PSCustomObject]@{ Target = "${HostName}:$port"; Passed = $ok }
    }

    $apiOk = $false
    try {
        $resp = Invoke-WebRequest -Uri "$($ApiUrl.TrimEnd('/'))/api/version" -UseBasicParsing -TimeoutSec 15
        $apiOk = ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 500)
    } catch {}
    $checks += [PSCustomObject]@{ Target = "$($ApiUrl.TrimEnd('/'))/api/version"; Passed = $apiOk }

    return [PSCustomObject]@{
        Passed = (($checks | Where-Object { -not $_.Passed }).Count -eq 0)
        Checks = $checks
    }
}

function Write-RustDeskConfigVerifyReport {
    param($VerifyResult)
    foreach ($report in $VerifyResult.Reports) {
        if ($report.Reason -eq "missing") {
            Write-DeployLog "Config verify: missing $($report.Path)"
            continue
        }
        if ($report.Reason -eq "unreadable") {
            Write-DeployLog "Config verify: $($report.Path) $($report.Reason)"
            continue
        }
        $d = $report.Details
        Write-DeployLog "Config verify: $($report.Path) host=$($d.ActualHost) relay=$($d.ActualRelay) api=$($d.ActualApi) rendezvous=$($d.ActualRendezvous) passed=$($report.Passed)"
    }
}

function Assert-RustDeskConfigApplied {
    param(
        [string]$ExpectedHost,
        [string]$ExpectedRelay,
        [string]$ExpectedApi,
        [string]$ExpectedKey,
        [int]$MaxAttempts = 15
    )

    for ($i = 0; $i -lt $MaxAttempts; $i++) {
        $verify = Test-RustDeskConfigApplied -ExpectedHost $ExpectedHost -ExpectedRelay $ExpectedRelay -ExpectedApi $ExpectedApi -ExpectedKey $ExpectedKey
        if ($verify.Passed) {
            Write-DeployLog "Config verified at $($verify.Path)"
            Write-Host " -> Config verified: $($verify.Path)" -ForegroundColor Green
            return $verify
        }
        Start-Sleep -Seconds 2
    }

    $final = Test-RustDeskConfigApplied -ExpectedHost $ExpectedHost -ExpectedRelay $ExpectedRelay -ExpectedApi $ExpectedApi -ExpectedKey $ExpectedKey
    Write-RustDeskConfigVerifyReport -VerifyResult $final
    $details = ($final.Reports | Where-Object { $_.Details } | Select-Object -Last 1).Details
    $msg = "RustDesk config verification failed."
    if ($details) {
        $msg += " Expected host=$ExpectedHost relay=$ExpectedRelay api=$ExpectedApi."
        $msg += " Last seen host=$($details.ActualHost) relay=$($details.ActualRelay) api=$($details.ActualApi)."
    }
    Write-Error $msg
    exit 1
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (WINDOWS)     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Deploy log: $logPath" -ForegroundColor Cyan
Write-DeployLog "Starting RustDesk deployment."
Write-DeployLog "UserName=$env:USERNAME ApiUrl=$ApiUrl ConfigLength=$($ConfigString.Length)"

$expectedHost = Normalize-RustDeskHost $IdServer
$expectedRelay = Normalize-RustDeskHost $(if ($RelayServer) { $RelayServer } else { $IdServer })
$expectedApi = $ApiUrl.TrimEnd('/')
$expectedKey = $Key.Trim()
Write-DeployLog "Expected host=$expectedHost relay=$expectedRelay api=$expectedApi"

if (-not $DeployToken) {
    Write-Error "Deploy token is missing. Generate a new command from Web Admin."
    exit 1
}
if (-not $ConfigString) {
    Write-Error "Server config string is missing. Regenerate deploy command from Web Admin."
    exit 1
}

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Run PowerShell as Administrator."
    exit 1
}

$rustdeskExe = "C:\Program Files\RustDesk\rustdesk.exe"
if (-not (Test-Path $rustdeskExe)) {
    Write-Host "[1/7] Installing RustDesk..." -ForegroundColor Yellow
    $tempExe = "$env:TEMP\rustdesk-installer.exe"
    $downloadUrl = "https://github.com/rustdesk/rustdesk/releases/download/1.2.3-2/rustdesk-1.2.3-2-x86_64.exe"
    try {
        $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/rustdesk/rustdesk/releases/latest"
        $asset = $releases.assets | Where-Object { $_.name -like "*x86_64.exe" -and $_.name -notlike "*crd*" } | Select-Object -First 1
        if ($asset) { $downloadUrl = $asset.browser_download_url }
    } catch {}
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempExe -UseBasicParsing
    Start-Process -FilePath $tempExe -ArgumentList "--silent-install" -NoNewWindow -Wait
    $waitCount = 0
    while (-not (Test-Path $rustdeskExe) -and $waitCount -lt 30) {
        Start-Sleep -Seconds 2
        $waitCount++
    }
    if (-not (Test-Path $rustdeskExe)) {
        Write-Error "RustDesk install failed."
        exit 1
    }
    Write-DeployLog "RustDesk installed."
}

Stop-RustDeskRuntime

Write-Host "[2/7] Applying server config..." -ForegroundColor Yellow
Write-DeployLog "Running rustdesk --config"
$configOutput = (& $rustdeskExe --config $ConfigString 2>&1 | Out-String).Trim()
if ($configOutput) {
    Write-DeployLog "rustdesk --config output: $configOutput"
}
if ($LASTEXITCODE -and $LASTEXITCODE -ne 0) {
    Write-Warning "rustdesk --config exit code: $LASTEXITCODE"
}

Start-RustDeskRuntime

Write-Host "[3/7] Verifying server config..." -ForegroundColor Yellow
$null = Assert-RustDeskConfigApplied -ExpectedHost $expectedHost -ExpectedRelay $expectedRelay -ExpectedApi $expectedApi -ExpectedKey $expectedKey
$network = Test-RustDeskNetwork -HostName $expectedHost -ApiUrl $expectedApi
foreach ($check in $network.Checks) {
    $status = if ($check.Passed) { "ok" } else { "failed" }
    Write-DeployLog "Network verify $($check.Target): $status"
    if (-not $check.Passed) {
        Write-Warning "Network check failed: $($check.Target)"
    }
}
if (-not $network.Passed) {
    Write-Warning "Some network checks failed. Deployment continues, but remote access may not work until ports/API are reachable."
}

Write-Host "[4/7] Reading device ID..." -ForegroundColor Yellow
$id = ""
for ($i = 0; $i -lt 30; $i++) {
    $id = Get-RustDeskDeviceId -Exe $rustdeskExe
    if ($id) { break }
    Start-Sleep -Seconds 2
}
if (-not $id) {
    Write-Error "Cannot read device ID from RustDesk CLI."
    exit 1
}
Write-DeployLog "Device ID: $id"

$cleanApiUrl = $ApiUrl.TrimEnd('/')
$headers = @{
    "Authorization" = "Bearer $DeployToken"
    "Content-Type"  = "application/json"
}

Write-Host "[5/7] Registering device with API..." -ForegroundColor Yellow
$deployBody = @{ id = $id } | ConvertTo-Json
$deployResponse = Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/deploy" -Method Post -Headers $headers -Body $deployBody
if ($deployResponse.result -ne "OK") {
    Write-Error "Device deploy failed: $($deployResponse | ConvertTo-Json -Compress)"
    exit 1
}

Write-Host "[6/7] Setting host password..." -ForegroundColor Yellow
if ($PasswordMode -eq "custom" -and $CustomPassword) {
    $hostPassword = $CustomPassword
    Write-DeployLog "Setting custom host password from deploy token."
} else {
    $idTail = if ($id.Length -ge 5) { $id.Substring($id.Length - 5) } else { $id.PadLeft(5, '0') }
    $hostPassword = "Rd@$idTail"
    Write-DeployLog "Setting structured host password: $hostPassword"
}
Write-Host " -> Host password: $hostPassword" -ForegroundColor Green
Stop-RustDeskRuntime
Start-Process -FilePath $rustdeskExe -ArgumentList "--password", $hostPassword -NoNewWindow -Wait
Start-RustDeskRuntime

Write-Host "[7/7] Syncing address book..." -ForegroundColor Yellow
$base64Password = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($hostPassword))
$deployedAt = [int][DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$deployNote = "Deploy: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
$cliBody = @{
    id = $id
    address_book_name = "My Devices"
    address_book_tag = "deploy"
    address_book_password = $base64Password
    address_book_alias = $env:COMPUTERNAME
    address_book_note = $deployNote
    deployed_at = $deployedAt
} | ConvertTo-Json
Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/cli" -Method Post -Headers $headers -Body $cliBody | Out-Null

try {
    Invoke-RestMethod -Uri "$cleanApiUrl/api/deploy/revoke" -Method Post -Headers $headers -Body "{}" | Out-Null
} catch {}

Write-Host "Deployment completed successfully." -ForegroundColor Green
Write-Host "Device ID: $id" -ForegroundColor Green
Write-DeployLog "Deployment completed successfully."
