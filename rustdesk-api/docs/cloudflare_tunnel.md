# Hướng dẫn cấu hình Cloudflare Tunnel trên Host OS (Không dùng Docker)

Tài liệu này hướng dẫn cách cài đặt và cấu hình **Cloudflare Tunnel (`cloudflared`)** chạy trực tiếp dưới dạng một dịch vụ (service/daemon) trên hệ điều hành của máy chủ (Host OS) để chuyển tiếp lưu lượng truy cập tới các dịch vụ RustDesk đang chạy trong Docker.

---

## Bước 1: Chuẩn bị cổng kết nối trong Docker
Đảm bảo bạn đang sử dụng tệp `docker-compose.yaml` mới nhất, trong đó các cổng dịch vụ nội bộ đã được mở ra ngoài localhost (Host OS):
* **`21114`**: RustDesk API Server.
* **`21118`**: hbbs Websocket (ID).
* **`21119`**: hbbr Websocket (Relay).

---

## Bước 2: Cài đặt cloudflared trên Host OS

### A. Đối với Linux (Ubuntu / Debian)
1. Tải và cài đặt khóa GPG của kho lưu trữ Cloudflare:
   ```bash
   sudo mkdir -p --mode=0755 /usr/share/keyrings
   curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
   ```
2. Thêm kho lưu trữ của Cloudflare vào hệ thống:
   ```bash
   echo 'deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflare-main.gpg cloudflare main' | sudo tee /etc/apt/sources.list.d/cloudflare-main.list
   ```
3. Cập nhật danh sách gói và cài đặt `cloudflared`:
   ```bash
   sudo apt-get update && sudo apt-get install cloudflared
   ```

### B. Đối với Windows
1. Tải xuống bản cài đặt `.msi` dành cho Windows từ trang chính thức:
   [Tải cloudflared cho Windows](https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-windows-amd64.msi)
2. Chạy tệp `.msi` để cài đặt. Cửa sổ dòng lệnh `cloudflared` sẽ khả dụng trên PowerShell hoặc CMD.

---

## Bước 3: Cấu hình Tunnel thông qua Cloudflare Zero Trust (Khuyên dùng)

Cách dễ nhất và trực quan nhất là quản lý Tunnel của bạn trực tiếp thông qua giao diện web **Cloudflare Zero Trust**:

1. Truy cập vào **[Cloudflare Zero Trust Dashboard](https://one.dash.cloudflare.com/)**.
2. Di chuyển đến mục **Networks** $\rightarrow$ **Tunnels** và bấm **Create a tunnel**.
3. Chọn loại Tunnel là **Cloudflared** và đặt tên cho Tunnel của bạn (ví dụ: `rustdesk-server`).
4. Tại tab **Install and run a connector**, chọn Hệ điều hành tương ứng với máy chủ của bạn (Windows/Linux) và làm theo hướng dẫn:
   * **Sao chép lệnh cài đặt dịch vụ** được Cloudflare cung cấp sẵn (lệnh này chứa token định danh duy nhất của Tunnel).
   * **Thực thi lệnh đó trên máy chủ của bạn** (với quyền `sudo` trên Linux hoặc Administrator trên Windows).
   * Trạng thái Tunnel trên trang Dashboard của Cloudflare sẽ chuyển sang **Active** (màu xanh).

---

## Bước 4: Thiết lập định tuyến (Public Hostnames)

Trên trang cấu hình Tunnel của Cloudflare Dashboard, di chuyển sang tab **Public Hostname** và thêm 3 bản ghi định tuyến tương ứng với 3 dịch vụ của RustDesk:

### 1. Định tuyến cho API và Web Admin/Web Client
* **Subdomain**: Nhập tên miền phụ (ví dụ: `rustdesk` nếu muốn truy cập qua `rustdesk.domain.com`).
* **Domain**: Chọn tên miền của bạn.
* **Type**: `HTTP`
* **URL**: `localhost:21114`

### 2. Định tuyến cho Websocket ID
* **Subdomain**: Nhập tên miền phụ giống hệt ở bước 1.
* **Domain**: Chọn tên miền của bạn.
* **Path**: `ws/id`
* **Type**: `HTTP` (Cloudflare tự động nâng cấp lên WebSocket)
* **URL**: `localhost:21118`

### 3. Định tuyến cho Websocket Relay
* **Subdomain**: Nhập tên miền phụ giống hệt ở bước 1.
* **Domain**: Chọn tên miền của bạn.
* **Path**: `ws/relay`
* **Type**: `HTTP` (Cloudflare tự động nâng cấp lên WebSocket)
* **URL**: `localhost:21119`

> [!IMPORTANT]
> **Kích hoạt tính năng WebSockets**:
> Đảm bảo rằng bạn đã kích hoạt tính năng WebSockets trong phần cài đặt tên miền của mình trên Cloudflare:
> Truy cập **Cloudflare Dashboard** $\rightarrow$ **Tên miền của bạn** $\rightarrow$ **Network** $\rightarrow$ Bật **WebSockets** (chuyển sang On).

---

## Bước 5: Chạy hệ thống
1. Tạo tệp cấu hình `.env` trên máy chủ từ tệp mẫu:
   ```bash
   cp .env.example .env
   ```
2. Cấu hình tên miền của bạn trong `.env` (ví dụ: `DOMAIN=rustdesk.domain.com`).
3. Khởi chạy Docker Compose:
   ```bash
   docker compose up --build -d
   ```

Hệ thống sẽ tự động khởi tạo. Mọi truy cập vào `https://rustdesk.domain.com` sẽ được Cloudflare Tunnel tiếp nhận, mã hóa SSL tự động và đẩy về các cổng dịch vụ thích hợp chạy trong Docker.
