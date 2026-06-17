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

function Test-RustDeskServiceInstalled {
    return $null -ne (Get-Service -Name "rustdesk" -ErrorAction SilentlyContinue)
}

function Install-RustDeskServiceIfNeeded {
    param([string]$Exe)
    if (Test-RustDeskServiceInstalled) {
        return $true
    }
    Write-Host "Installing RustDesk Windows service..." -ForegroundColor Yellow
    Write-DeployLog "Running rustdesk --install-service"
    try {
        $proc = Start-Process -FilePath $Exe -ArgumentList "--install-service" -Wait -PassThru -WindowStyle Hidden
        Write-DeployLog "rustdesk --install-service exit code: $($proc.ExitCode)"
    } catch {
        Write-DeployLog "rustdesk --install-service failed: $($_.Exception.Message)"
    }
    Start-Sleep -Seconds 5
    return (Test-RustDeskServiceInstalled)
}

function Test-RustDeskIpcPipeReady {
    return Test-Path "\\.\pipe\RustDesk\query"
}

function Wait-RustDeskIpcReady {
    param([int]$MaxAttempts = 30)
    for ($i = 0; $i -lt $MaxAttempts; $i++) {
        if (Test-RustDeskIpcPipeReady) {
            return $true
        }
        Start-Sleep -Seconds 2
    }
    return $false
}

function Start-RustDeskServerProcessIfNeeded {
    param([string]$Exe)
    if (Test-RustDeskIpcPipeReady) {
        return
    }
    Write-DeployLog "Starting rustdesk --server for IPC."
    Write-Host " -> Starting RustDesk server process for IPC..." -ForegroundColor Yellow
    Start-Process -FilePath $Exe -ArgumentList "--server" -WindowStyle Hidden | Out-Null
    if (-not (Wait-RustDeskIpcReady)) {
        Write-Error "RustDesk IPC pipe is not available. Install the Windows service or start RustDesk before setting password."
        exit 1
    }
}

function Ensure-RustDeskIpcReady {
    param([string]$Exe)

    if (Test-RustDeskIpcPipeReady) {
        return
    }

    $serviceInstalled = Install-RustDeskServiceIfNeeded -Exe $Exe
    if ($serviceInstalled) {
        Write-DeployLog "Starting RustDesk Windows service."
        try {
            $svc = Get-Service -Name "rustdesk" -ErrorAction Stop
            if ($svc.Status -ne "Running") {
                Start-Service -Name "rustdesk" -ErrorAction Stop
            }
            Start-Sleep -Seconds 5
            if (Wait-RustDeskIpcReady) {
                return
            }
            Write-DeployLog "RustDesk service is running but IPC pipe is not ready yet."
        } catch {
            Write-DeployLog "Failed to start RustDesk service: $($_.Exception.Message)"
        }
    }

    Start-RustDeskServerProcessIfNeeded -Exe $Exe
}

function Start-RustDeskRuntime {
    param([string]$Exe = $rustdeskExe)
    Ensure-RustDeskIpcReady -Exe $Exe
}

function Set-RustDeskPermanentPassword {
    param(
        [string]$Exe,
        [string]$Password
    )

    Ensure-RustDeskIpcReady -Exe $Exe

    $lastOutput = ""
    for ($i = 0; $i -lt 10; $i++) {
        $lastOutput = (& $Exe --password "$Password" 2>&1 | Out-String).Trim()
        Write-DeployLog "rustdesk --password attempt $($i + 1): $lastOutput"
        if ($lastOutput -match 'Done') {
            $optionOutput = (& $Exe --option verification-method use-permanent-password 2>&1 | Out-String).Trim()
            if ($optionOutput) {
                Write-DeployLog "verification-method output: $optionOutput"
            }
            Write-DeployLog "verification-method set to use-permanent-password"
            return
        }
        if ($lastOutput -match 'Installation and administrative privileges required') {
            Write-Error "Deploy script must run in an elevated Administrator PowerShell session to set permanent password."
            exit 1
        }
        if ($lastOutput -match 'cannot find the file|os error 2') {
            Ensure-RustDeskIpcReady -Exe $Exe
        }
        Start-Sleep -Seconds 2
    }

    Write-Error "Failed to set permanent password. Output: $lastOutput"
    exit 1
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

function Escape-RustDeskTomlSingleQuoted {
    param([string]$Value)
    return ($Value -replace "'", "''")
}

function Set-RustDeskTomlOption {
    param(
        [string]$Content,
        [string]$Key,
        [string]$Value
    )
    $escaped = Escape-RustDeskTomlSingleQuoted $Value
    $line = "$Key = '$escaped'"
    $pattern = "(?m)^\s*$([regex]::Escape($Key))\s*=.*$"
    if ($Content -match $pattern) {
        return [regex]::Replace($Content, $pattern, $line)
    }
    if ($Content -match '(?m)^\s*\[options\]\s*$') {
        return [regex]::Replace($Content, '(?m)^\s*\[options\]\s*$', "[options]`n$line")
    }
    return ($Content.TrimEnd() + "`n`n[options]`n$line`n")
}

function Get-RustDeskLocalConfigPaths {
    return @(
        "$env:APPDATA\RustDesk\config\RustDesk_local.toml",
        "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk_local.toml"
    )
}

function Set-RustDeskClientLogin {
    param(
        [string]$AccessToken,
        [string]$UserInfoJson
    )

    $defaultToml = @"
remote_id = ''
kb_layout_type = ''
size = [
    0,
    0,
    0,
    0,
]
fav = []

[options]

[ui_flutter]
"@

    foreach ($tomlPath in (Get-RustDeskLocalConfigPaths)) {
        $parent = Split-Path -Parent $tomlPath
        if (-not (Test-Path -LiteralPath $parent -ErrorAction SilentlyContinue)) {
            try {
                New-Item -ItemType Directory -Path $parent -Force | Out-Null
            } catch {
                Write-DeployLog "Client login skipped for $tomlPath : cannot create directory"
                continue
            }
        }

        $content = $defaultToml
        if (Test-Path -LiteralPath $tomlPath -ErrorAction SilentlyContinue) {
            try {
                $content = Get-Content -LiteralPath $tomlPath -Raw -ErrorAction Stop
            } catch {
                Write-DeployLog "Client login skipped for $tomlPath : unreadable"
                continue
            }
        }

        $content = Set-RustDeskTomlOption -Content $content -Key "access_token" -Value $AccessToken
        $content = Set-RustDeskTomlOption -Content $content -Key "user_info" -Value $UserInfoJson
        try {
            Set-Content -LiteralPath $tomlPath -Value $content -Encoding UTF8
            Write-DeployLog "Client login stored at $tomlPath"
            Write-Host " -> Account login stored: $tomlPath" -ForegroundColor Green
        } catch {
            Write-DeployLog "Client login failed for $tomlPath : $($_.Exception.Message)"
        }
    }
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
    Write-Host "[1/8] Installing RustDesk..." -ForegroundColor Yellow
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

Install-RustDeskServiceIfNeeded -Exe $rustdeskExe

Stop-RustDeskRuntime

Write-Host "[2/8] Applying server config..." -ForegroundColor Yellow
Write-DeployLog "Running rustdesk --config"
$configOutput = (& $rustdeskExe --config $ConfigString 2>&1 | Out-String).Trim()
if ($configOutput) {
    Write-DeployLog "rustdesk --config output: $configOutput"
}
if ($LASTEXITCODE -and $LASTEXITCODE -ne 0) {
    Write-Warning "rustdesk --config exit code: $LASTEXITCODE"
}

Start-RustDeskRuntime

Write-Host "[3/8] Verifying server config..." -ForegroundColor Yellow
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

Write-Host "[4/8] Reading device ID..." -ForegroundColor Yellow
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

Write-Host "[5/8] Registering device with API..." -ForegroundColor Yellow
$deployBody = @{ id = $id } | ConvertTo-Json
$deployResponse = Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/deploy" -Method Post -Headers $headers -Body $deployBody
if ($deployResponse.result -ne "OK") {
    Write-Error "Device deploy failed: $($deployResponse | ConvertTo-Json -Compress)"
    exit 1
}

Write-Host "[6/8] Setting host password..." -ForegroundColor Yellow
if ($PasswordMode -eq "custom" -and $CustomPassword) {
    $hostPassword = $CustomPassword
    Write-DeployLog "Setting custom host password from deploy token."
} else {
    $idTail = if ($id.Length -ge 5) { $id.Substring($id.Length - 5) } else { $id.PadLeft(5, '0') }
    $hostPassword = "Rd@$idTail"
    Write-DeployLog "Setting structured host password: $hostPassword"
}
Write-Host " -> Host password: $hostPassword" -ForegroundColor Green
Set-RustDeskPermanentPassword -Exe $rustdeskExe -Password $hostPassword

Write-Host "[7/8] Syncing address book..." -ForegroundColor Yellow
$deployedAt = [int][DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$deployNote = "Deploy: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
$cliBody = @{
    id = $id
    address_book_name = "My Devices"
    address_book_tag = "deploy"
    address_book_password = $hostPassword
    address_book_alias = $env:COMPUTERNAME
    address_book_note = $deployNote
    deployed_at = $deployedAt
} | ConvertTo-Json
Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/cli" -Method Post -Headers $headers -Body $cliBody | Out-Null

Write-Host "[8/8] Signing in RustDesk account..." -ForegroundColor Yellow
try {
    $loginBody = @{ id = $id } | ConvertTo-Json
    $loginResponse = Invoke-RestMethod -Uri "$cleanApiUrl/api/deploy/client-login" -Method Post -Headers $headers -Body $loginBody
    if ($loginResponse.type -eq "access_token" -and $loginResponse.access_token) {
        $userInfo = @{ name = $loginResponse.user.name }
        if ($loginResponse.user.email) { $userInfo.email = $loginResponse.user.email }
        $userInfoJson = ($userInfo | ConvertTo-Json -Compress)
        Set-RustDeskClientLogin -AccessToken $loginResponse.access_token -UserInfoJson $userInfoJson
        Write-Host " -> RustDesk account auto-login configured." -ForegroundColor Green
        Get-Process -Name "rustdesk" -ErrorAction SilentlyContinue | Where-Object { $_.SessionId -eq (Get-Process -Id $PID).SessionId } | Stop-Process -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        Start-Process -FilePath $rustdeskExe
        Write-DeployLog "RustDesk UI restarted to apply account login."
    } else {
        Write-Warning "Client auto-login skipped: unexpected API response."
        Write-DeployLog "Client auto-login skipped: unexpected API response."
    }
} catch {
    Write-Warning "Client auto-login failed: $($_.Exception.Message)"
    Write-DeployLog "Client auto-login failed: $($_.Exception.Message)"
}

try {
    Invoke-RestMethod -Uri "$cleanApiUrl/api/deploy/revoke" -Method Post -Headers $headers -Body "{}" | Out-Null
} catch {}

Write-Host "Deployment completed successfully." -ForegroundColor Green
Write-Host "Device ID: $id" -ForegroundColor Green
Write-DeployLog "Deployment completed successfully."
