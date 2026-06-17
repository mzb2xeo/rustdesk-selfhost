# RustDesk Self-Hosted Stack

Bộ triển khai RustDesk tự lưu trữ gồm ID Server, Relay Server, API Server và Web Admin trong một Docker Compose stack. Mục tiêu của repo là đóng gói các thành phần cần thiết để vận hành RustDesk nội bộ, có tài khoản người dùng, sổ địa chỉ, cấu hình client tự động và luồng deploy thiết bị cho Windows.

> Đây là bản tích hợp/fork phục vụ triển khai riêng, không phải bản phát hành chính thức của RustDesk.

Tài liệu: **Tiếng Việt** | [English](README_EN.md)

## Nội Dung Chính

- `hbbs`: ID/Rendezvous server, quản lý đăng ký thiết bị và NAT traversal.
- `hbbr`: Relay server, chuyển tiếp phiên điều khiển khi không thể kết nối trực tiếp.
- `rustdesk-api`: API quản trị, xác thực, cấu hình server, address book và deploy token.
- `rustdesk-api-web`: Web Admin dùng để quản lý người dùng, thiết bị, sổ địa chỉ và tạo lệnh deploy.
- `reverse-proxy`: Nginx entrypoint cho API/Web Admin/WebSocket.

## Nguồn Upstream Và Các Bản Fork

| Thành phần | Mô tả | Nguồn fork | Phiên bản gốc | Thay đổi so với bản gốc |
|---|---|---|---|---|
| `hbbs` | RustDesk ID/Rendezvous server | `https://github.com/rustdesk/rustdesk-server` | `1.1.15` | Tích hợp kiểm tra `MUST_LOGIN`, dùng `JWT_SECRET` chung với API, hỗ trợ luồng bắt buộc đăng nhập/deploy trước khi thiết bị được đăng ký. |
| `hbbr` | RustDesk relay server | `https://github.com/rustdesk/rustdesk-server` | `1.1.15` | Đóng gói trong cùng Docker stack, dùng chung network namespace với `hbbs`, cấu hình relay theo domain nội bộ của stack. |
| `rustdesk-api` | API server | `https://github.com/lejianwen/rustdesk-api` | `2.7` | Bổ sung deploy token ngắn hạn, route tải PowerShell deploy script, auth bằng deploy token cho `/api/devices/deploy` và `/api/devices/cli`, đọc public key từ `hbbs`, cấu hình server tự động cho client. |
| `rustdesk-api-web` | Web Admin | `https://github.com/lejianwen/rustdesk-api` | `2.7` | Bổ sung trang `My -> Client Config`, hiển thị cấu hình client, tạo lệnh tự tải script và chạy deploy, tải script deploy, copy command/token metadata. |
| Docker/ops trong repo này | Local integration/custom fork | Local working tree | Theo các nguồn trên | Thêm `docker-compose.yml`, `Dockerfile`, `Dockerfile.server`, `nginx.conf`, `deploy-host.ps1`, volume dữ liệu chung và tài liệu vận hành cho triển khai self-hosted. |

Khi cập nhật fork, cần phân biệt rõ:

- Cập nhật `rustdesk-server` có thể ảnh hưởng giao thức đăng ký thiết bị, `MUST_LOGIN`, WebSocket và keypair.
- Cập nhật `rustdesk-api` có thể ảnh hưởng schema DB, token, route admin/API và address book.
- Cập nhật `rustdesk-api-web` có thể ảnh hưởng menu, i18n, endpoint gọi từ Web Admin.

## Cấu Trúc Repo

```text
.
├── docker-compose.yml          # Stack vận hành chính
├── nginx.conf                  # Reverse proxy cho API/WebSocket
├── Dockerfile                  # Build rustdesk-api + rustdesk-api-web
├── Dockerfile.server           # Build hbbs/hbbr
├── deploy-host.ps1             # Wrapper tải script deploy từ API và chạy trên Windows
├── rustdesk-server/            # Mã server RustDesk / hbbs / hbbr
├── rustdesk-api/               # API server Go
├── rustdesk-api-web/           # Web Admin Vue
└── data/                       # Dữ liệu runtime: DB, key, log
```

Repo hiện tại không có `.gitmodules`; các thư mục thành phần là working tree riêng. Kiểm tra remote bằng:

```bash
git -C rustdesk-server remote -v
git -C rustdesk-api remote -v
git -C rustdesk-api-web remote -v
```

## Kiến Trúc

```text
Client RustDesk
   │
   ├── 21116/tcp+udp ──> hbbs: ID / rendezvous / NAT punch
   ├── 21117/tcp     ──> hbbr: relay
   └── HTTPS/WSS     ──> reverse-proxy ──> rustdesk-api / Web Admin / WebSocket

hbbs/hbbr ── ./data/server
rustdesk-api ── ./data/api
rustdesk-api ── read-only ./data/server/id_ed25519.pub
```

### Cổng Dịch Vụ

| Cổng | Giao thức | Thành phần | Mục đích |
|---|---|---|---|
| `8082` | TCP | Nginx | HTTP entrypoint trong compose hiện tại |
| `21114` | TCP | API qua namespace `hbbs` | API/Web Admin khi truy cập trực tiếp |
| `21115` | TCP | `hbbs` | Control port |
| `21116` | TCP/UDP | `hbbs` | ID/Rendezvous, NAT punch |
| `21117` | TCP | `hbbr` | Relay |
| `21118` | TCP | `hbbs` | WebSocket ID cho Web Client |
| `21119` | TCP | `hbbr` | WebSocket Relay cho Web Client |

Với môi trường production, thường đặt reverse proxy ngoài hoặc load balancer TLS phía trước `8082`, sau đó trỏ domain public về API/Web Admin.

## Biến Môi Trường

Tạo file `.env` ở thư mục gốc:

```env
DOMAIN=rd.example.com
DOMAIN_API=rustdesk.example.com
TZ=Asia/Ho_Chi_Minh
JWT_SECRET=change_me_to_a_long_random_secret
MUST_LOGIN=N
```

Ý nghĩa chính:

- `DOMAIN`: domain/IP mà RustDesk client dùng cho `hbbs` và `hbbr`.
- `DOMAIN_API`: domain public của API/Web Admin, dùng để sinh URL cấu hình client và deploy script.
- `JWT_SECRET`: khóa ký JWT, phải đồng bộ giữa `hbbs` và `rustdesk-api`.
- `MUST_LOGIN=Y`: yêu cầu client đăng nhập/deploy trước khi được đăng ký thiết bị.

## Khởi Chạy

```bash
docker compose up -d --build
```

Kiểm tra trạng thái:

```bash
docker compose ps
docker compose logs -f rustdesk-api
```

Web Admin:

```text
http://<server-ip>:8082/_admin/
```

Mật khẩu admin khởi tạo được ghi trong log của container `rustdesk-api`.

## Deploy Client Windows

Luồng khuyến nghị cho người dùng cuối:

1. Đăng nhập Web Admin.
2. Vào `My` -> `Client Config`.
3. Bấm `Tạo lệnh triển khai`.
4. Sao chép lệnh `Tự tải script và chạy deploy`.
5. Mở PowerShell bằng quyền Administrator trên máy Windows cần deploy.
6. Dán lệnh và chạy.

Script sẽ:

- Tải RustDesk client nếu máy chưa cài.
- Ghi cấu hình `ID Server`, `Relay Server`, `API Server`, public key.
- Xác minh `RustDesk2.toml` (host/relay/api/key) và kiểm tra kết nối mạng tới ID/Relay/API.
- Đọc device ID bằng `rustdesk.exe --get-id` (trên Windows cần pipe: `| Out-String`).
- Gọi API `/api/devices/deploy` trực tiếp bằng deploy token ngắn hạn.
- Sinh mật khẩu tĩnh ngẫu nhiên cho unattended access.
- Đồng bộ thiết bị vào address book `My Devices`.
- Thu hồi deploy token sau khi hoàn tất.

API ưu tiên đọc template deploy từ `data/templates/deploy-host.ps1`. Trong Docker, đường dẫn này tương ứng với `./data/api/templates/deploy-host.ps1` trên host, nên có thể sửa template script mà không cần build lại binary Go. Nếu file này không tồn tại, API dùng template mặc định trong `resources/templates/deploy-host.ps1`, rồi mới fallback về template embed trong code.

Có thể chạy wrapper thủ công nếu đã có token:

```powershell
.\deploy-host.ps1 -DeployToken "<DEPLOY_TOKEN>" -ApiUrl "https://rustdesk.example.com"
```

## Deploy Client Thủ Công

Windows:

```cmd
"C:\Program Files\RustDesk\rustdesk.exe" --deploy --token <USER_API_TOKEN>
"C:\Program Files\RustDesk\rustdesk.exe" --get-id | more
```

PowerShell (đọc ID — bản GUI Windows thường không in ra console trực tiếp):

```powershell
& "C:\Program Files\RustDesk\rustdesk.exe" --get-id | Out-String
```

Linux/macOS:

```bash
sudo rustdesk --deploy --token <USER_API_TOKEN>
rustdesk --get-id
```

Khi deploy thành công, thiết bị sẽ xuất hiện trong Web Admin/My Devices hoặc được sync vào address book nếu chạy qua script tự động.

## Vận Hành

Build lại sau khi sửa code:

```bash
docker compose up -d --build
```

Xem log:

```bash
docker compose logs -f hbbs
docker compose logs -f hbbr
docker compose logs -f rustdesk-api
docker compose logs -f reverse-proxy
```

Backup tối thiểu:

```text
data/server/   # keypair hbbs/hbbr
data/api/      # database API
.env           # domain, JWT secret, flags
```

Không mất `data/server/id_ed25519` nếu không muốn tất cả client phải cấu hình lại key.

## Cập Nhật Fork / Upstream

Trước khi cập nhật:

```bash
git status --short
git -C rustdesk-server status --short
git -C rustdesk-api status --short
git -C rustdesk-api-web status --short
```

Cập nhật từng phần:

```bash
git -C rustdesk-server pull --ff-only
git -C rustdesk-api pull --ff-only
git -C rustdesk-api-web pull --ff-only
docker compose up -d --build
```

Sau cập nhật, cần kiểm tra tối thiểu:

- Web Admin đăng nhập được.
- Client nhận đúng `ID Server`, `Relay Server`, `API Server`, `key`.
- Deploy token sinh được trong `My -> Client Config`.
- Script Windows chạy được tới bước sync address book.
- `MUST_LOGIN=Y` không cho thiết bị lạ tự đăng ký nếu đang bật chế độ bắt buộc deploy.

## Kiểm Thử Local

API:

```bash
cd rustdesk-api
go test ./...
```

Frontend:

```bash
cd rustdesk-api-web
npm run build
```

Toàn bộ stack:

```bash
docker compose up -d --build
docker compose ps
```

## Ghi Chú Bảo Mật

- Không commit `.env`, `data/`, database, private key hoặc token.
- Dùng `JWT_SECRET` dài, ngẫu nhiên và không dùng lại giữa môi trường test/production.
- Chỉ sinh deploy token khi cần, vì token cho phép gán thiết bị vào tài khoản.
- Nên đặt TLS ở reverse proxy public trước khi cấp lệnh deploy cho người dùng ngoài mạng nội bộ.
- Sau khi đổi domain/API URL, tạo lại lệnh deploy mới để script tải đúng endpoint.

## Giấy Phép Và Trách Nhiệm

Các thành phần RustDesk và các fork liên quan giữ giấy phép của upstream tương ứng. Repo tích hợp này chỉ mô tả cách đóng gói, cấu hình và các chỉnh sửa vận hành nội bộ. Khi phân phối lại binary hoặc image, cần rà soát license của từng thành phần upstream/fork.
