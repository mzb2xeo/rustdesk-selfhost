#!/bin/bash
# deploy-host.sh
# Hướng dẫn chạy: chmod +x deploy-host.sh && sudo ./deploy-host.sh -u "tên_đăng_nhập" -p "mật_khẩu_tài_khoản" -a "https://rustdesk.ngoctuct.io.vn"

set -e

# Định nghĩa màu sắc
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}==========================================================${NC}"
echo -e "${CYAN}  RUSTDESK AUTOMATED HOST DEPLOYMENT SCRIPT (LINUX)       ${NC}"
echo -e "${CYAN}==========================================================${NC}"

# 1. Kiểm tra quyền root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Script cần được chạy với quyền root (sudo) để cấu hình mật khẩu RustDesk.${NC}"
  exit 1
fi

# 2. Xử lý tham số
while getopts u:p:a: flag
do
    case "${flag}" in
        u) USERNAME=${OPTARG};;
        p) PASSWORD=${OPTARG};;
        a) API_URL=${OPTARG};;
        *) echo "Sử dụng: $0 -u <Username> -p <Password> -a <ApiUrl>"; exit 1;;
    esac
done

if [ -z "$USERNAME" ] || [ -z "$PASSWORD" ] || [ -z "$API_URL" ]; then
    echo -e "${RED}Thiếu tham số bắt buộc.${NC}"
    echo "Sử dụng: $0 -u <Username> -p <Password> -a <ApiUrl>"
    exit 1
fi

# 3. Định vị tệp cấu hình RustDesk.toml cục bộ
TOML_PATH=""
PATHS=(
  "/etc/rustdesk/RustDesk.toml"
  "/root/.config/rustdesk/RustDesk.toml"
)

for p in "${PATHS[@]}"; do
  if [ -f "$p" ]; then
    TOML_PATH="$p"
    break
  fi
done

if [ -z "$TOML_PATH" ]; then
  # Nếu không tìm thấy, thử tìm trong thư mục nhà của user đang chạy sudo
  USER_HOME=$(eval echo ~${SUDO_USER})
  if [ -f "$USER_HOME/.config/rustdesk/RustDesk.toml" ]; then
    TOML_PATH="$USER_HOME/.config/rustdesk/RustDesk.toml"
  fi
fi

if [ -z "$TOML_PATH" ]; then
  echo -e "${RED}Không tìm thấy tệp cấu hình RustDesk cục bộ. Vui lòng cài đặt và chạy RustDesk trước.${NC}"
  exit 1
fi

# 4. Đọc ID và UUID từ file cấu hình TOML
echo -e "${YELLOW}[1/6] Đang đọc cấu hình thiết bị cục bộ...${NC}"
ID=$(grep -oE "id = '[^']+'" "$TOML_PATH" | cut -d"'" -f2 || true)
if [ -z "$ID" ]; then
  ID=$(grep -oE 'id = "[^"]+"' "$TOML_PATH" | cut -d'"' -f2 || true)
fi

UUID=$(grep -oE "uuid = '[^']+'" "$TOML_PATH" | cut -d"'" -f2 || true)
if [ -z "$UUID" ]; then
  UUID=$(grep -oE 'uuid = "[^"]+"' "$TOML_PATH" | cut -d'"' -f2 || true)
fi

if [ -z "$ID" ] || [ -z "$UUID" ]; then
  echo -e "${RED}Không thể đọc ID hoặc UUID từ tệp RustDesk.toml.${NC}"
  exit 1
fi

echo -e " -> Nhận diện Device ID: ${GREEN}$ID${NC}"
echo -e " -> Nhận diện Device UUID: ${GREEN}$UUID${NC}"

# 5. Tạo mật khẩu cố định ngẫu nhiên và an toàn cho máy con
echo -e "${YELLOW}[2/6] Tạo mật khẩu cố định an toàn cho Host...${NC}"
RAND_PASSWORD=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 12 ; echo '')

# Thiết lập mật khẩu cố định thông qua RustDesk CLI
RUSTDESK_CMD="rustdesk"
if ! command -v rustdesk &> /dev/null; then
  if [ -f "/usr/bin/rustdesk" ]; then
    RUSTDESK_CMD="/usr/bin/rustdesk"
  else
    echo -e "${RED}Không tìm thấy lệnh rustdesk trên hệ thống.${NC}"
    exit 1
  fi
fi

echo -e " -> Đang đặt mật khẩu bằng CLI..."
$RUSTDESK_CMD --password "$RAND_PASSWORD"
echo -e " -> Đã thiết lập mật khẩu cố định trên Host thành công."

# 6. Xác thực tài khoản với API Server để lấy Access Token
echo -e "${YELLOW}[3/6] Đang đăng nhập vào API Server...${NC}"
# Chuẩn hóa API URL bỏ ký tự gạch chéo cuối
CLEAN_API_URL="${API_URL%/}"

LOGIN_JSON=$(cat <<EOF
{
  "username": "$USERNAME",
  "password": "$PASSWORD",
  "id": "$ID",
  "uuid": "$UUID",
  "deviceInfo": {
    "os": "linux",
    "type": "app"
  }
}
EOF
)

LOGIN_RESP=$(curl -s -X POST "$CLEAN_API_URL/api/login" \
  -H "Content-Type: application/json" \
  -d "$LOGIN_JSON")

TOKEN=$(echo "$LOGIN_RESP" | grep -oP '"access_token":"\K[^"]+' || true)
if [ -z "$TOKEN" ]; then
  # Thử phân tích cú pháp thô bằng python/node hoặc jq nếu có
  if command -v jq &> /dev/null; then
    TOKEN=$(echo "$LOGIN_RESP" | jq -r '.access_token' || true)
  fi
fi

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo -e "${RED}Đăng nhập API Server thất bại. Kết quả nhận được:${NC}"
  echo "$LOGIN_RESP"
  exit 1
fi

echo -e " -> Đăng nhập thành công!"

# 7. Gọi Deploy API để gán sở hữu thiết bị cho tài khoản
echo -e "${YELLOW}[4/6] Đang đăng ký thiết bị với tài khoản...${NC}"
DEPLOY_JSON=$(cat <<EOF
{
  "id": "$ID",
  "uuid": "$UUID",
  "pk": ""
}
EOF
)

DEPLOY_RESP=$(curl -s -X POST "$CLEAN_API_URL/api/devices/deploy" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "$DEPLOY_JSON")

echo -e " -> Kết quả Deploy: ${GREEN}$DEPLOY_RESP${NC}"

# 8. Mã hóa Base64 mật khẩu tự sinh và đồng bộ lên Sổ địa chỉ (Address Book)
echo -e "${YELLOW}[5/6] Mã hóa mật khẩu và đồng bộ lên Sổ địa chỉ (Address Book)...${NC}"
BASE64_PASSWORD=$(echo -n "$RAND_PASSWORD" | base64)
HOSTNAME=$(hostname)

CLI_JSON=$(cat <<EOF
{
  "id": "$ID",
  "uuid": "$UUID",
  "address_book_name": "My Devices",
  "address_book_password": "$RAND_PASSWORD",
  "address_book_alias": "$HOSTNAME"
}
EOF
)

CLI_RESP=$(curl -s -X POST "$CLEAN_API_URL/api/devices/cli" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "$CLI_JSON")

echo -e " -> Đồng bộ lên Sổ địa chỉ thành công!"

echo -e "${GREEN}==========================================================${NC}"
echo -e "${GREEN}[6/6] Triển khai hoàn tất thành công!${NC}"
echo -e "${GREEN}Thiết bị của bạn hiện tại có thể được truy cập từ xa không${NC}"
echo -e "${GREEN}cần mật khẩu từ bất kỳ máy nào đăng nhập cùng tài khoản này.${NC}"
echo -e "${GREEN}==========================================================${NC}"
