# deploy-host.ps1
# Huong dan chay:
#  1. Tao token trong Web Admin -> My -> Client Config -> "Tao lenh trien khai"
#  2. Chay: .\deploy-host.ps1 -DeployToken "<TOKEN>" -ApiUrl "https://rustdesk.example.com"

param (
    [Parameter(Mandatory=$true)]
    [string]$DeployToken,

    [Parameter(Mandatory=$true)]
    [string]$ApiUrl
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (WINDOWS)     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Run PowerShell as Administrator."
    exit 1
}

$cleanApiUrl = $ApiUrl.TrimEnd('/')
$scriptUrl = "$cleanApiUrl/api/deploy/powershell?deploy_token=$DeployToken"

Write-Host "Downloading deployment script..." -ForegroundColor Yellow
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    $script = (Invoke-WebRequest -UseBasicParsing -Uri $scriptUrl).Content
    Invoke-Expression $script
} catch {
    Write-Error "Deploy failed. Token may be invalid or expired. Generate a new command from Web Admin. Details: $_"
    exit 1
}
