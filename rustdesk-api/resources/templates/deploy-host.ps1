# RustDesk automated host deployment (Windows)

param (
    [string]$DeployToken = "{{.DeployToken}}",
    [string]$ApiUrl = "{{.ApiUrl}}",
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

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (WINDOWS)     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Deploy log: $logPath" -ForegroundColor Cyan
Write-DeployLog "Starting RustDesk deployment."
Write-DeployLog "UserName=$env:USERNAME ApiUrl=$ApiUrl ConfigLength=$($ConfigString.Length)"

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

Write-Host "[2/6] Applying server config..." -ForegroundColor Yellow
Write-DeployLog "Running rustdesk --config"
$configOutput = & $rustdeskExe --config $ConfigString 2>&1
if ($configOutput) {
    Write-DeployLog "rustdesk --config output: $configOutput"
}
if ($LASTEXITCODE -and $LASTEXITCODE -ne 0) {
    Write-Warning "rustdesk --config exit code: $LASTEXITCODE"
}

Start-RustDeskRuntime

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
Write-DeployLog "Device ID: $id"

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

Write-Host "[6/6] Syncing address book..." -ForegroundColor Yellow
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
