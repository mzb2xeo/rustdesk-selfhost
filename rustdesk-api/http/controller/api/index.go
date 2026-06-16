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

// DeployPowershell returns automated powershell configuration script
func (i *Index) DeployPowershell(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")

	apiServer := global.Config.Rustdesk.ApiServer
	if apiServer == "" || strings.Contains(apiServer, "127.0.0.1") || strings.Contains(apiServer, "localhost") {
		scheme := "http"
		if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		apiServer = scheme + "://" + c.Request.Host
	}

	idServer := global.Config.Rustdesk.IdServer
	if idServer == "" {
		host := c.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		idServer = host + ":21116"
	}

	relayServer := global.Config.Rustdesk.RelayServer
	if relayServer == "" {
		host := c.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		relayServer = host + ":21117"
	}

	key := global.Config.Rustdesk.Key

	script := powershellTemplate
	script = strings.ReplaceAll(script, "{{.Username}}", username)
	script = strings.ReplaceAll(script, "{{.Password}}", password)
	script = strings.ReplaceAll(script, "{{.ApiUrl}}", apiServer)
	script = strings.ReplaceAll(script, "{{.IdServer}}", idServer)
	script = strings.ReplaceAll(script, "{{.RelayServer}}", relayServer)
	script = strings.ReplaceAll(script, "{{.Key}}", key)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, script)
}

const powershellTemplate = `# deploy-host.ps1
# Cấu hình tự động cài đặt và triển khai RustDesk Host cho Windows

param (
    [string]$Username = "{{.Username}}",
    [string]$Password = "{{.Password}}",
    [string]$ApiUrl = "{{.ApiUrl}}",
    [string]$IdServer = "{{.IdServer}}",
    [string]$RelayServer = "{{.RelayServer}}",
    [string]$Key = "{{.Key}}"
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (WINDOWS)     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Kiểm tra quyền Admin
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Script cần được chạy với quyền Administrator để cài đặt và cấu hình RustDesk."
    Write-Warning "Vui lòng mở PowerShell dưới quyền Administrator và chạy lại script."
    exit 1
}

# 2. Nhập thông tin nếu thiếu
if (-not $Username) {
    $Username = Read-Host "Nhập tài khoản RustDesk (Email/Username)"
}
if (-not $Password) {
    $Password = Read-Host "Nhập mật khẩu tài khoản RustDesk" -AsSecureString
    # Convert SecureString to plain text
    $BSTR = [System.Runtime.InteropServices.Marshal]::SecureToBSTR($Password)
    $Password = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
}

# 3. Tải và cài đặt RustDesk nếu chưa có
$rustdeskExe = "C:\Program Files\RustDesk\rustdesk.exe"
if (-not (Test-Path $rustdeskExe)) {
    Write-Host "[1/7] RustDesk chưa được cài đặt. Tiến hành tải và cài đặt..." -ForegroundColor Yellow
    $tempExe = "$env:TEMP\rustdesk-installer.exe"
    
    # Mặc định tải từ github
    $downloadUrl = "https://github.com/rustdesk/rustdesk/releases/download/1.2.3-2/rustdesk-1.2.3-2-x86_64.exe"
    try {
        $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/rustdesk/rustdesk/releases/latest"
        $asset = $releases.assets | Where-Object { $_.name -like "*x86_64.exe" -and $_.name -notlike "*crd*" } | Select-Object -First 1
        if ($asset) {
            $downloadUrl = $asset.browser_download_url
        }
    } catch {
        Write-Warning "Không thể tự động lấy link tải mới nhất từ Github, sử dụng link mặc định."
    }
    
    Write-Host " -> Đang tải RustDesk từ: $downloadUrl" -ForegroundColor Yellow
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempExe -UseBasicParsing
    Write-Host " -> Đang cài đặt RustDesk silent..." -ForegroundColor Yellow
    Start-Process -FilePath $tempExe -ArgumentList "--silent-install" -NoNewWindow -Wait
    
    # Đợi cài đặt và file exe xuất hiện
    $waitCount = 0
    while (-not (Test-Path $rustdeskExe) -and $waitCount -lt 15) {
        Start-Sleep -Seconds 1
        $waitCount++
    }
    if (-not (Test-Path $rustdeskExe)) {
        Write-Error "Cài đặt RustDesk thất bại hoặc không tìm thấy file cài đặt tại C:\Program Files\RustDesk\rustdesk.exe"
        exit 1
    }
}

# Dừng service để ghi đè config ổn định
Stop-Service -Name "rustdesk" -ErrorAction SilentlyContinue

# 4. Ghi đè cấu hình Server vào RustDesk.toml
Write-Host "[2/7] Đang ghi cấu hình server và key..." -ForegroundColor Yellow
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
    if (-not $found) {
        $newLines += "$key = '$value'"
    }
    return $newLines
}

foreach ($tomlPath in $configPaths) {
    $dir = Split-Path $tomlPath
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Force -Path $dir | Out-Null
    }
    $content = @()
    if (Test-Path $tomlPath) {
        $content = Get-Content $tomlPath
    }
    $content = Set-TomlKey $content "id-server" $IdServer
    $content = Set-TomlKey $content "relay-server" $RelayServer
    $content = Set-TomlKey $content "api-server" $ApiUrl
    $content = Set-TomlKey $content "key" $Key
    $content | Set-Content $tomlPath -Force
}

# Khởi động lại service RustDesk để sinh ID và UUID
Start-Service -Name "rustdesk" -ErrorAction SilentlyContinue
Start-Sleep -Seconds 5

# 5. Đọc ID và UUID từ file cấu hình TOML mới sinh
Write-Host "[3/7] Đang đọc cấu hình thiết bị cục bộ..." -ForegroundColor Yellow
$tomlPath = "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk.toml"
if (-not (Test-Path $tomlPath)) {
    $tomlPath = "$env:APPDATA\RustDesk\config\RustDesk.toml"
}

if (-not (Test-Path $tomlPath)) {
    Write-Error "Không tìm thấy tệp cấu hình RustDesk cục bộ để đọc ID và UUID."
    exit 1
}

$content = Get-Content $tomlPath -Raw
$id = [regex]::Match($content, 'id\s*=\s*''([^'']+)''').Groups[1].Value
if (-not $id) {
    $id = [regex]::Match($content, 'id\s*=\s*"([^"]+)"').Groups[1].Value
}
$uuid = [regex]::Match($content, 'uuid\s*=\s*''([^'']+)''').Groups[1].Value
if (-not $uuid) {
    $uuid = [regex]::Match($content, 'uuid\s*=\s*"([^"]+)"').Groups[1].Value
}

if (-not $id -or -not $uuid) {
    Write-Error "Không thể đọc ID hoặc UUID từ tệp RustDesk.toml. Đảm bảo dịch vụ RustDesk đang chạy."
    exit 1
}

Write-Host " -> Nhận diện Device ID: $id" -ForegroundColor Green
Write-Host " -> Nhận diện Device UUID: $uuid" -ForegroundColor Green

# 6. Tạo mật khẩu cố định ngẫu nhiên và an toàn cho máy con
Write-Host "[4/7] Tạo mật khẩu cố định an toàn cho Host..." -ForegroundColor Yellow
$charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
$randomPassword = -join ((1..12) | ForEach-Object { $charset[(Get-Random -Maximum $charset.Length)] })

Write-Host " -> Đang đặt mật khẩu bằng CLI..." -ForegroundColor Yellow
try {
    Start-Process -FilePath $rustdeskExe -ArgumentList "--password", $randomPassword -NoNewWindow -Wait
    Write-Host " -> Đã thiết lập mật khẩu cố định thành công." -ForegroundColor Green
} catch {
    Write-Error "Lỗi khi cấu hình mật khẩu qua RustDesk CLI."
    exit 1
}

# 7. Xác thực tài khoản với API Server để lấy Access Token
Write-Host "[5/7] Đang đăng nhập vào API Server..." -ForegroundColor Yellow
$cleanApiUrl = $ApiUrl.TrimEnd('/')
$loginBody = @{
    username = $Username
    password = $Password
    id = $id
    uuid = $uuid
    deviceInfo = @{
        os = "windows"
        type = "app"
    }
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$cleanApiUrl/api/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.access_token
    if (-not $token) {
         throw "Không nhận được access_token từ API."
    }
    Write-Host " -> Đăng nhập thành công!" -ForegroundColor Green
} catch {
    Write-Error "Đăng nhập API Server thất bại. Chi tiết: $_"
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# 8. Gọi Deploy API để gán sở hữu thiết bị cho tài khoản
Write-Host "[6/7] Đang đăng ký thiết bị với tài khoản..." -ForegroundColor Yellow
$deployBody = @{
    id = $id
    uuid = $uuid
    pk = ""
} | ConvertTo-Json

try {
    $deployResponse = Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/deploy" -Method Post -Headers $headers -Body $deployBody
    Write-Host " -> Kết quả Deploy: $($deployResponse.result)" -ForegroundColor Green
} catch {
    Write-Error "Gửi yêu cầu Deploy lên API thất bại. Chi tiết: $_"
    exit 1
}

# 9. Mã hóa Base64 mật khẩu tự sinh và đồng bộ lên Sổ địa chỉ (Address Book)
Write-Host "[7/7] Mã hóa mật khẩu và đồng bộ lên Sổ địa chỉ (Address Book)..." -ForegroundColor Yellow
$bytes = [System.Text.Encoding]::UTF8.GetBytes($randomPassword)
$base64Password = [Convert]::ToBase64String($bytes)

$cliBody = @{
    id = $id
    uuid = $uuid
    address_book_name = "My Devices"
    address_book_password = $base64Password
    address_book_alias = $env:COMPUTERNAME
} | ConvertTo-Json

try {
    $cliResponse = Invoke-RestMethod -Uri "$cleanApiUrl/api/devices/cli" -Method Post -Headers $headers -Body $cliBody
    Write-Host " -> Đồng bộ lên Sổ địa chỉ thành công!" -ForegroundColor Green
} catch {
    Write-Error "Đồng bộ mật khẩu lên Sổ địa chỉ thất bại. Chi tiết: $_"
    exit 1
}

Write-Host "==========================================================" -ForegroundColor Green
Write-Host "Triển khai và cấu hình hoàn tất thành công!" -ForegroundColor Green
Write-Host "Mật khẩu ngẫu nhiên đã được đặt và đồng bộ lên Sổ địa chỉ." -ForegroundColor Green
Write-Host "Thiết bị có thể được truy cập từ xa không cần mật khẩu từ tài khoản của bạn." -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green
`
