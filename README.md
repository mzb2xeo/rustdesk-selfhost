# Dự Án RustDesk Server & API Server Đồng Bộ (Self-Hosted)

Dự án này là bộ tích hợp hoàn chỉnh giúp triển khai máy chủ điều khiển máy tính từ xa **RustDesk** tự lưu trữ (self-hosted) với đầy đủ tính năng:
- **RustDesk Server (hbbs & hbbr)**: Dịch vụ ID/Rendezvous và Relay chính thức.
- **RustDesk API Server**: Máy chủ dịch vụ tài khoản, đồng bộ sổ địa chỉ (Address Book), nhóm và giao diện quản trị Web (Web Admin).
- **Web Admin Frontend**: Giao diện người dùng và quản trị trực quan.

Dự án được cấu trúc dưới dạng một Git repository cha, liên kết với các repository con qua cơ chế **Git Submodule** nhằm mục đích dễ dàng cập nhật mã nguồn gốc từ các nguồn upstream phục vụ cho mục tiêu cập nhật tính năng mới và các bản vá lỗi bảo mật.

---

## 📂 Cấu trúc thư mục dự án

```text
D:\projects\RustDesk\
├── rustdesk-server/            # Submodule: ID/Relay Server chính thức (Rust)
├── rustdesk-api/               # Submodule: API Server Go (Fork từ lejianwen/rustdesk-api)
├── rustdesk-api-web/           # Submodule: Giao diện Web Admin Vue 3
├── rustdesk-server-lejianwen/  # Submodule: Phiên bản fork (Chỉ dùng để tham khảo phần xác thực đăng nhập)
├── data/                       # [Bị bỏ qua bởi Git] Dữ liệu vận hành (Database, Keys, Logs)
├── .env                        # [Bị bỏ qua bởi Git] Cấu hình môi trường chứa khóa bảo mật
├── .gitignore                  # Cấu hình bỏ qua tệp tin rác của Git
├── Dockerfile                  # Chỉ dẫn xây dựng container cho rustdesk-api (Go + Vue)
├── Dockerfile.server           # Chỉ dẫn xây dựng container cho hbbs/hbbr (Rust gốc)
└── docker-compose.yml          # Trình quản lý điều phối khởi chạy hệ thống Docker
```

## 🌐 Kiến trúc và Phương án Triển khai (Deployment Strategy)

Dự án hỗ trợ mô hình triển khai **Single-Node Dockerized Stack** giúp tối ưu hóa tài nguyên phần cứng, đồng thời đảm bảo bảo mật và dễ vận hành.

### 1. Sơ đồ các luồng Dịch vụ
* **Nginx Reverse Proxy**: Làm cổng ngõ duy nhất hứng traffic HTTPS (`443`) và HTTP (`80`). Nó thực hiện SSL termination và định tuyến:
  * Traffic API & Web Admin tới `rustdesk-api` (Port `21114`).
  * Traffic WebSocket kết nối client (đối với Web Client) tới `hbbs`/`hbbr` (Port `21118`/`21119`).
* **hbbs (ID Server)**: Dịch vụ điều khiển, quản lý đăng ký thiết bị và hỗ trợ đấm cổng (Punch Hole).
* **hbbr (Relay Server)**: Trung chuyển dữ liệu màn hình khi kết nối trực tiếp (Direct Connection) thất bại hoặc bị áp đặt (Force Relay).
* **rustdesk-api**: Xử lý logic đăng nhập, xác thực JWT, đồng bộ Sổ địa chỉ (Address Book), nhóm thiết bị và giao diện Web Admin.

### 2. Cấu hình Tường lửa (Firewall / Ports to open)
Để hệ thống hoạt động ổn định (đặc biệt là kết nối đấm cổng UDP và trung chuyển Relay), bạn cần mở các cổng sau trên tường lửa của máy chủ:

| Cổng | Giao thức | Dịch vụ | Chức năng |
|---|---|---|---|
| **80** / **443** | TCP | Nginx Proxy | Truy cập Web Admin, REST API & Web Client |
| **21115** | TCP | hbbs | Cổng điều khiển hbbs |
| **21116** | TCP | hbbs | Cổng truy vấn ID |
| **21116** | **UDP** | hbbs | Cổng Rendezvous / Punch Hole (Bắt buộc phải mở UDP) |
| **21117** | TCP | hbbr | Cổng điều khiển hbbr (Relay) |
| **21118** | TCP | hbbs | Cổng WebSocket ID (phục vụ Web Client) |
| **21119** | TCP | hbbr | Cổng WebSocket Relay (phục vụ Web Client) |

### 3. Đồng bộ hóa Khóa và Bảo mật
* **Khóa công khai (Public Key Encryption)**:
  * Khi `hbbs` khởi chạy lần đầu, nó sẽ tự động sinh cặp khóa bảo mật (trong `./data/server/`).
  * `rustdesk-api` được mount chung thư mục này dưới dạng Read-Only (`/data:ro`) để tự động đọc file public key `id_ed25519.pub` và trả về cấu hình cho các client PC khi đăng nhập mà không cần cấu hình thủ công.
* **Xác thực JWT & Bắt buộc Đăng nhập (MUST_LOGIN)**:
  * Biến `JWT_SECRET` được đồng bộ hóa giữa `hbbs` và `rustdesk-api` để kiểm tra chữ ký token đăng nhập.
  * Nếu đặt `MUST_LOGIN=Y` ở file `.env`, client RustDesk bắt buộc phải đăng nhập bằng tài khoản (đồng bộ qua API Server) trước khi có thể lấy ID hoặc bắt đầu kết nối từ xa.

---

## 🛠️ Hướng dẫn cài đặt và Khởi chạy

### 1. Chuẩn bị biến môi trường
Sao chép mẫu hoặc chỉnh sửa trực tiếp tệp tin `.env` ở thư mục gốc:
```env
# Địa chỉ IP hoặc Domain thực tế của máy chủ (ví dụ: 192.168.1.100 hoặc rustdesk.example.com)
DOMAIN=127.0.0.1

# Múi giờ hệ thống
TZ=Asia/Ho_Chi_Minh

# Khóa JWT bí mật dùng để mã hóa và ký Token đăng nhập. Hãy thay đổi thành chuỗi bảo mật ngẫu nhiên!
JWT_SECRET=super_secret_jwt_sign_key_change_me
```

### 2. Khởi chạy toàn bộ hệ thống bằng Docker Compose
Chạy lệnh sau tại thư mục gốc để tự động build và chạy dịch vụ ngầm:
```bash
docker compose up -d --build
```
Lưu ý: Quá trình chạy lần đầu tiên có thể mất vài phút vì Docker cần tải ảnh môi trường Rust và tiến hành biên dịch mã nguồn `hbbs`/`hbbr` từ đầu.

Sau khi khởi chạy hoàn tất:
- Truy cập trang quản trị Web Admin tại: `http://<IP_Server_Của_Bạn>:21114/_admin/`
- Tài khoản mặc định ban đầu là `admin` và mật khẩu ngẫu nhiên được hiển thị trong nhật ký log khởi động của container `rustdesk-api`.

---

## 🔄 Quy trình cập nhật mã nguồn (Update bảo mật)

Vì các dự án con được liên kết qua Git Submodule, bạn có thể dễ dàng cập nhật mã nguồn của từng dự án từ tác giả gốc để áp dụng các bản vá bảo mật mới:

### 1. Cập nhật tất cả các submodules về phiên bản mới nhất
Chạy lệnh này từ thư mục gốc để kéo các commit mới nhất trên nhánh master của các repository nguồn:
```bash
git submodule update --remote --merge
```

### 2. Cập nhật thủ công một submodule cụ thể
Ví dụ, bạn muốn cập nhật bản vá bảo mật mới nhất cho máy chủ Rust từ repository chính thức:
```bash
cd rustdesk-server
git checkout master
git pull origin master
cd ..
# Sau đó commit thay đổi pointer của submodule này lên git cha
git add rustdesk-server
git commit -m "Update rustdesk-server submodule to latest release"
```

### 3. Build lại hệ thống sau khi cập nhật mã nguồn
Sau khi kéo mã nguồn mới về, chạy lệnh sau để build lại các image Docker mới:
```bash
docker compose up -d --build
```

---

## 📝 Quản lý Git Repository Cha

Để đẩy toàn bộ cấu trúc dự án này lên kho lưu trữ cá nhân của bạn trên GitHub (vẫn giữ nguyên liên kết submodule):

```bash
# 1. Khởi tạo Git nếu chưa làm
git init

# 2. Add các tệp cấu hình và khai báo submodule
git add .gitignore README.md docker-compose.yml Dockerfile Dockerfile.server .gitmodules

# 3. Add các thư mục submodule (Lưu ý: không add nội dung con trực tiếp mà add dưới dạng gitlink)
git add rustdesk-server rustdesk-api rustdesk-api-web rustdesk-server-lejianwen

# 4. Commit lần đầu
git commit -m "Initial commit: Unified RustDesk Stack with submodules"

# 5. Đẩy lên Github cá nhân
git remote add origin <URL_GITHUB_REPOSITRY_CỦA_BẠN>
git branch -M main
git push -u origin main
```
