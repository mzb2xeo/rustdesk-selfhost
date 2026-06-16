package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rustdesk-api/global"
	requstform "rustdesk-api/http/request/api"
	"rustdesk-api/http/response"
	"rustdesk-api/model"
	"rustdesk-api/service"
	"strings"
	"time"
)

type Index struct {
}

// Index Home Page
// @Tags Home Page
// @Summary Home Page
// @Description front page
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router / [get]
func (i *Index) Index(c *gin.Context) {
	response.Success(
		c,
		"Hello Gwen",
	)
}

// Heartbeat
// @Tags Home Page
// @Summary heartbeat
// @Description heartbeat
// @Accept  json
// @Produce  json
// @Success 200 {object} nil
// @Failure 500 {object} response.Response
// @Router /heartbeat [post]
func (i *Index) Heartbeat(c *gin.Context) {
	info := &requstform.PeerInfoInHeartbeat{}
	err := c.ShouldBindJSON(info)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if info.Uuid == "" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	peer := service.AllService.PeerService.FindById(info.Id)
	if peer == nil || peer.RowId == 0 {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	//If it is within 40s, it will not be updated.
	if time.Now().Unix()-peer.LastOnlineTime >= 30 {
		upp := &model.Peer{RowId: peer.RowId, LastOnlineTime: time.Now().Unix(), LastOnlineIp: c.ClientIP()}
		service.AllService.PeerService.Update(upp)
	}
	c.JSON(http.StatusOK, gin.H{})
}

// Version version
// @Tags Home Page
// @Summary version
// @Description Version
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /version [get]
func (i *Index) Version(c *gin.Context) {
	//Read resources/version file
	v := service.AllService.AppService.GetAppVersion()
	response.Success(
		c,
		v,
	)
}

// DeployPowershell returns automated powershell configuration script (deploy token required).
func (i *Index) DeployPowershell(c *gin.Context) {
	deployToken := strings.TrimSpace(c.Query("deploy_token"))
	if deployToken == "" {
		c.String(http.StatusBadRequest, "deploy_token is required")
		return
	}
	if _, err := service.AllService.DeployTokenService.FindValid(deployToken); err != nil {
		c.String(http.StatusUnauthorized, "invalid or expired deploy token")
		return
	}

	apiServer := resolvePublicApiServer(c)
	idServer := resolvePublicIdServer(c)
	relayServer := resolvePublicRelayServer(c)
	key := global.Config.Rustdesk.Key

	script := powershellTemplate
	script = strings.ReplaceAll(script, "{{.DeployToken}}", deployToken)
	script = strings.ReplaceAll(script, "{{.ApiUrl}}", apiServer)
	script = strings.ReplaceAll(script, "{{.IdServer}}", idServer)
	script = strings.ReplaceAll(script, "{{.RelayServer}}", relayServer)
	script = strings.ReplaceAll(script, "{{.Key}}", key)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, script)
}

// DeployRevoke consumes a deploy token after successful setup.
func (i *Index) DeployRevoke(c *gin.Context) {
	authType, _ := c.Get("authType")
	if authType != "deploy" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	token, _ := c.Get("token")
	if token == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	_ = service.AllService.DeployTokenService.Consume(token.(string))
	c.JSON(http.StatusOK, gin.H{})
}

func resolvePublicApiServer(c *gin.Context) string {
	apiServer := global.Config.Rustdesk.ApiServer
	if apiServer == "" || strings.Contains(apiServer, "127.0.0.1") || strings.Contains(apiServer, "localhost") {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		apiServer = scheme + "://" + c.Request.Host
	}
	return strings.TrimRight(apiServer, "/")
}

func resolvePublicIdServer(c *gin.Context) string {
	idServer := global.Config.Rustdesk.IdServer
	if idServer == "" {
		host := c.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		idServer = host + ":21116"
	}
	return idServer
}

func resolvePublicRelayServer(c *gin.Context) string {
	relayServer := global.Config.Rustdesk.RelayServer
	if relayServer == "" {
		host := c.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		relayServer = host + ":21117"
	}
	return relayServer
}

const powershellTemplate = `# RustDesk automated host deployment (Windows)

param (
    [string]$DeployToken = "{{.DeployToken}}",
    [string]$ApiUrl = "{{.ApiUrl}}",
    [string]$IdServer = "{{.IdServer}}",
    [string]$RelayServer = "{{.RelayServer}}",
    [string]$Key = "{{.Key}}"
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (WINDOWS)     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

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
    "$env:APPDATA\RustDesk\config\RustDesk.toml"
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

foreach ($tomlPath in $configPaths) {
    $dir = Split-Path $tomlPath
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
    $content = @()
    if (Test-Path $tomlPath) { $content = Get-Content $tomlPath }
    $content = Set-TomlKey $content "custom-rendezvous-server" ($IdServer -replace ':21116$','')
    $content = Set-TomlKey $content "relay-server" ($RelayServer -replace ':21117$','')
    $content = Set-TomlKey $content "api-server" $ApiUrl
    $content = Set-TomlKey $content "key" $Key
    $content | Set-Content $tomlPath -Force
}

Start-Service -Name "rustdesk" -ErrorAction SilentlyContinue
Start-Sleep -Seconds 5

Write-Host "[3/6] Registering device with RustDesk CLI..." -ForegroundColor Yellow
$deployOutput = & $rustdeskExe --deploy --token $DeployToken 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Error "RustDesk CLI deploy failed: $deployOutput"
    exit 1
}

Write-Host "[4/6] Reading device ID..." -ForegroundColor Yellow
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

Write-Host "[5/6] Setting host password..." -ForegroundColor Yellow
$charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
$randomPassword = -join ((1..12) | ForEach-Object { $charset[(Get-Random -Maximum $charset.Length)] })
Start-Process -FilePath $rustdeskExe -ArgumentList "--password", $randomPassword -NoNewWindow -Wait

$cleanApiUrl = $ApiUrl.TrimEnd('/')
$headers = @{
    "Authorization" = "Bearer $DeployToken"
    "Content-Type"  = "application/json"
}

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
`
