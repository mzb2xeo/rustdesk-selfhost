# RustDesk automated host deployment (Windows)

param (
    [string]$DeployToken = "{{.DeployToken}}",
    [string]$ApiUrl = "{{.ApiUrl}}",
    [string]$IdServer = "{{.IdServer}}",
    [string]$RelayServer = "{{.RelayServer}}",
    [string]$Key = "{{.Key}}"
)

$ErrorActionPreference = "Stop"
$logPath = Join-Path $env:TEMP "rustdesk-deploy.log"
function Write-DeployLog($message) {
    $line = "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $message"
    Add-Content -Path $logPath -Value $line -Encoding UTF8
}

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (WINDOWS)     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Deploy log: $logPath" -ForegroundColor Cyan
Write-DeployLog "Starting RustDesk deployment."
Write-DeployLog "UserName=$env:USERNAME UserProfile=$env:USERPROFILE AppData=$env:APPDATA Temp=$env:TEMP"
Write-DeployLog "ApiUrl=$ApiUrl IdServer=$IdServer RelayServer=$RelayServer KeyLength=$($Key.Length)"

if (-not $DeployToken) {
    Write-Error "Deploy token is missing. Generate a new command from Web Admin."
    exit 1
}

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Run PowerShell as Administrator."
    exit 1
}

$rustdeskExe = "C:\Program Files\RustDesk\rustdesk.exe"
if (-not (Test-Path $rustdeskExe)) {
    Write-Host "[1/6] Installing RustDesk..." -ForegroundColor Yellow
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
    while (-not (Test-Path $rustdeskExe) -and $waitCount -lt 15) {
        Start-Sleep -Seconds 1
        $waitCount++
    }
    if (-not (Test-Path $rustdeskExe)) {
        Write-Error "RustDesk install failed."
        exit 1
    }
}

Stop-Service -Name "rustdesk" -ErrorAction SilentlyContinue

Write-Host "[2/6] Writing server config..." -ForegroundColor Yellow
$configPaths = @(
    "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk.toml",
    "C:\Windows\System32\config\systemprofile\AppData\Roaming\RustDesk\config\RustDesk.toml",
    "C:\ProgramData\RustDesk\config\RustDesk.toml",
    "$env:APPDATA\RustDesk\config\RustDesk.toml"
)
$optionsConfigPaths = @(
    "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk2.toml",
    "C:\Windows\System32\config\systemprofile\AppData\Roaming\RustDesk\config\RustDesk2.toml",
    "C:\ProgramData\RustDesk\config\RustDesk2.toml",
    "$env:APPDATA\RustDesk\config\RustDesk2.toml"
)

function Set-TomlKey($lines, $key, $value) {
    $newLines = @()
    $found = $false
    foreach ($line in $lines) {
        if ($line -match "^$key\s*=") {
            $newLines += "$key = '$value'"
            $found = $true
        } else {
            $newLines += $line
        }
    }
    if (-not $found) { $newLines += "$key = '$value'" }
    return $newLines
}

function Set-TomlSectionKey($lines, $section, $key, $value) {
    $newLines = @()
    $inSection = $false
    $sectionFound = $false
    $keyFound = $false
    foreach ($line in $lines) {
        if ($line -match "^\s*\[$section\]\s*$") {
            $sectionFound = $true
            $inSection = $true
            $newLines += $line
            continue
        }
        if ($inSection -and $line -match "^\s*\[.+\]\s*$") {
            if (-not $keyFound) {
                $newLines += "$key = '$value'"
                $keyFound = $true
            }
            $inSection = $false
            $newLines += $line
            continue
        }
        if ($inSection -and $line -match "^$key\s*=") {
            $newLines += "$key = '$value'"
            $keyFound = $true
        } else {
            $newLines += $line
        }
    }
    if (-not $sectionFound) {
        if ($newLines.Count -gt 0 -and $newLines[-1] -ne "") { $newLines += "" }
        $newLines += "[$section]"
        $newLines += "$key = '$value'"
    } elseif ($inSection -and -not $keyFound) {
        $newLines += "$key = '$value'"
    }
    return $newLines
}

foreach ($tomlPath in $configPaths) {
    $dir = Split-Path $tomlPath
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
    Write-DeployLog "Writing legacy config: $tomlPath"
    $content = @()
    if (Test-Path $tomlPath) { $content = Get-Content $tomlPath }
    $content = Set-TomlKey $content "custom-rendezvous-server" ($IdServer -replace ':21116$','')
    $content = Set-TomlKey $content "relay-server" ($RelayServer -replace ':21117$','')
    $content = Set-TomlKey $content "api-server" $ApiUrl
    $content = Set-TomlKey $content "key" $Key
    $content | Set-Content $tomlPath -Force -Encoding UTF8
    Write-DeployLog "Wrote legacy config: $tomlPath"
    Write-DeployLog "Legacy config content begin: $tomlPath"
    Get-Content $tomlPath -Raw | Add-Content -Path $logPath -Encoding UTF8
    Write-DeployLog "Legacy config content end: $tomlPath"
}

foreach ($tomlPath in $optionsConfigPaths) {
    $dir = Split-Path $tomlPath
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
    Write-DeployLog "Writing options config: $tomlPath"
    $content = @()
    if (Test-Path $tomlPath) { $content = Get-Content $tomlPath }
    $content = Set-TomlSectionKey $content "options" "custom-rendezvous-server" ($IdServer -replace ':21116$','')
    $content = Set-TomlSectionKey $content "options" "relay-server" ($RelayServer -replace ':21117$','')
    $content = Set-TomlSectionKey $content "options" "api-server" $ApiUrl
    $content = Set-TomlSectionKey $content "options" "key" $Key
    $content | Set-Content $tomlPath -Force -Encoding UTF8
    Write-DeployLog "Wrote options config: $tomlPath"
    Write-DeployLog "Options config content begin: $tomlPath"
    Get-Content $tomlPath -Raw | Add-Content -Path $logPath -Encoding UTF8
    Write-DeployLog "Options config content end: $tomlPath"
}

Start-Service -Name "rustdesk" -ErrorAction SilentlyContinue
Start-Sleep -Seconds 5

Write-Host "[3/6] Reading device ID..." -ForegroundColor Yellow
$id = ""
for ($i = 0; $i -lt 30; $i++) {
    $idOutput = & $rustdeskExe --get-id 2>$null | Select-Object -First 1
    if ($idOutput) { $id = $idOutput.Trim() }
    if ($id) { break }
    Start-Sleep -Seconds 2
}
if (-not $id) {
    Write-Error "Cannot read device ID from RustDesk CLI."
    exit 1
}

$cleanApiUrl = $ApiUrl.TrimEnd('/')
$headers = @{
    "Authorization" = "Bearer $DeployToken"
    "Content-Type"  = "application/json"
}

Write-Host "[4/6] Registering device with API..." -ForegroundColor Yellow
$deployBody = @{ id = $id } | ConvertTo-Json
$deployResponse = Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/deploy" -Method Post -Headers $headers -Body $deployBody
if ($deployResponse.result -ne "OK") {
    Write-Error "Device deploy failed: $($deployResponse | ConvertTo-Json -Compress)"
    exit 1
}

Write-Host "[5/6] Setting host password..." -ForegroundColor Yellow
$charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
$randomPassword = -join ((1..12) | ForEach-Object { $charset[(Get-Random -Maximum $charset.Length)] })
Start-Process -FilePath $rustdeskExe -ArgumentList "--password", $randomPassword -NoNewWindow -Wait

Write-Host "[6/6] Syncing address book..." -ForegroundColor Yellow
$base64Password = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($randomPassword))
$cliBody = @{
    id = $id
    address_book_name = "My Devices"
    address_book_password = $base64Password
    address_book_alias = $env:COMPUTERNAME
} | ConvertTo-Json
Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/cli" -Method Post -Headers $headers -Body $cliBody | Out-Null

try {
    Invoke-RestMethod -Uri "$cleanApiUrl/api/deploy/revoke" -Method Post -Headers $headers -Body "{}" | Out-Null
} catch {}

Write-Host "Deployment completed successfully." -ForegroundColor Green
