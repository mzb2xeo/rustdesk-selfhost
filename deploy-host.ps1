# deploy-host.ps1
# Hướng dẫn chạy: .\deploy-host.ps1 -Username "tên_đăng_nhập" -Password "mật_khẩu_tài_khoản" -ApiUrl "https://rustdesk.ngoctuct.io.vn"

param (
    [Parameter(Mandatory=$true)]
    [string]$Username,

    [Parameter(ParameterSetName="Credentials", Mandatory=$true)]
    [string]$Password,

    [Parameter(Mandatory=$true)]
    [string]$ApiUrl
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (WINDOWS)     " -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Kiểm tra quyền Admin
$isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Warning "Script cần được chạy với quyền Administrator để cấu hình mật khẩu RustDesk."
    Write-Warning "Vui lòng mở PowerShell dưới quyền Administrator và chạy lại script."
    exit 1
}

# 2. Định vị tệp cấu hình RustDesk.toml cục bộ
$tomlPath = "C:\Windows\ServiceProfiles\LocalService\AppData\Roaming\RustDesk\config\RustDesk.toml"
if (-not (Test-Path $tomlPath)) {
    $tomlPath = "$env:APPDATA\RustDesk\config\RustDesk.toml"
}

if (-not (Test-Path $tomlPath)) {
    Write-Error "Không tìm thấy tệp cấu hình RustDesk cục bộ. Vui lòng cài đặt và khởi chạy RustDesk trước."
    exit 1
}

# 3. Đọc ID và UUID từ file cấu hình TOML
Write-Host "[1/6] Đang đọc cấu hình thiết bị cục bộ..." -ForegroundColor Yellow
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
    Write-Error "Không thể đọc ID hoặc UUID từ tệp RustDesk.toml. Đảm bảo RustDesk đã chạy và nhận dạng được ID."
    exit 1
}

Write-Host " -> Nhận diện Device ID: $id" -ForegroundColor Green
Write-Host " -> Nhận diện Device UUID: $uuid" -ForegroundColor Green

# 4. Tạo mật khẩu cố định ngẫu nhiên và an toàn cho máy con
Write-Host "[2/6] Tạo mật khẩu cố định an toàn cho Host..." -ForegroundColor Yellow
$charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
$randomPassword = -join ((1..12) | ForEach-Object { $charset[(Get-Random -Maximum $charset.Length)] })

# Thiết lập mật khẩu cố định thông qua RustDesk CLI
$rustdeskExe = "C:\Program Files\RustDesk\rustdesk.exe"
if (-not (Test-Path $rustdeskExe)) {
    $rustdeskExe = "rustdesk.exe"
}

Write-Host " -> Đang đặt mật khẩu bằng CLI..." -ForegroundColor Yellow
try {
    Start-Process -FilePath $rustdeskExe -ArgumentList "--password", $randomPassword -NoNewWindow -Wait
    Write-Host " -> Đã thiết lập mật khẩu cố định trên Host thành công." -ForegroundColor Green
} catch {
    Write-Error "Lỗi khi cấu hình mật khẩu qua RustDesk CLI. Hãy chắc chắn đường dẫn tới rustdesk.exe đúng."
    exit 1
}

# 5. Xác thực tài khoản với API Server để lấy Access Token
Write-Host "[3/6] Đang đăng nhập vào API Server..." -ForegroundColor Yellow
# Chuẩn hóa ApiUrl bỏ ký tự gạch chéo cuối nếu có
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

# 6. Gọi Deploy API để gán sở hữu thiết bị cho tài khoản
Write-Host "[4/6] Đang đăng ký thiết bị với tài khoản..." -ForegroundColor Yellow
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

# 7. Mã hóa Base64 mật khẩu tự sinh và đồng bộ lên Sổ địa chỉ (Address Book)
Write-Host "[5/6] Mã hóa mật khẩu và đồng bộ lên Sổ địa chỉ (Address Book)..." -ForegroundColor Yellow
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
Write-Host "[6/6] Triển khai hoàn tất thành công!" -ForegroundColor Green
Write-Host "Thiết bị của bạn hiện tại có thể được truy cập từ xa không" -ForegroundColor Green
Write-Host "cần mật khẩu từ bất kỳ máy nào đăng nhập cùng tài khoản này." -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green
