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

## 💻 Hướng dẫn Deploy Client cho Người dùng (Device Deployment)

Tính năng **Device Deployment** (hoặc Deploy Client) được áp dụng khi máy chủ cấu hình bắt buộc phải xác thực thiết bị trước khi cho phép đăng ký (tránh việc ID lạ tự ý đăng ký vào hệ thống). Khi tính năng này hoạt động, Client sẽ hiển thị cảnh báo yêu cầu deploy.

Để triển khai RustDesk Client cho từng thiết bị, bạn có hai lựa chọn dưới đây:

### Cách 1: Tự động Triển khai & Cấu hình qua PowerShell (Khuyên dùng cho Windows)
Đây là phương pháp nhanh nhất và hoàn toàn tự động dành cho các máy chủ chạy hệ điều hành Windows:
1. Đăng nhập vào giao diện Web Admin.
2. Điều hướng đến menu **Cấu hình & Tải Client (Client Setup & Downloads)** ở thanh điều hướng bên trái.
3. Ở thẻ **Tự động Cài đặt & Cấu hình (Windows)**, bạn có thể tích chọn **Nhúng tài khoản hiện tại vào script** để tự động gán thiết bị vào tài khoản đang đăng nhập.
4. Sao chép dòng lệnh PowerShell một hàng được tạo sẵn.
5. Mở ứng dụng **PowerShell dưới quyền Administrator (Run as Administrator)** trên máy trạm Windows, dán lệnh vào và nhấn **Enter**.
6. Script sẽ tự động:
   - Tải về phiên bản RustDesk Client chính thức phù hợp và cài đặt silent.
   - Ghi cấu hình máy chủ riêng gồm ID Server, Relay Server, API Server và Public Key vào hệ thống.
   - Khởi chạy dịch vụ và tự động sinh ngẫu nhiên một mật khẩu tĩnh bảo mật dài 12 ký tự cho unattended access.
   - Gửi yêu cầu đăng nhập và deploy thiết bị lên API Server để gán sở hữu cho tài khoản của bạn.
   - Mã hóa và đồng bộ hóa mật khẩu tĩnh lên **Sổ địa chỉ (Address Book)** cá nhân của bạn dưới tên máy tính trạm.

### Cách 2: Triển khai thủ công qua dòng lệnh CLI
Nếu bạn muốn cấu hình thủ công hoặc triển khai cho Linux/macOS, vui lòng thực hiện các bước sau:

#### Bước A: Lấy API Token của User
1. Đăng nhập vào giao diện Web Admin bằng tài khoản người dùng hoặc admin.
2. Truy cập mục **Thông tin cá nhân (Profile)** để lấy **API Token** tương ứng với tài khoản đó.

#### Bước B: Thực hiện lệnh Deploy trên thiết bị Client
Chạy file thực thi của RustDesk bằng dòng lệnh (CLI) với quyền Administrator (trên Windows) hoặc root (trên Linux/macOS).

##### Trên Windows (PowerShell/CMD với quyền Admin):
1. Mở PowerShell hoặc Command Prompt dưới quyền Administrator.
2. Điều hướng tới thư mục cài đặt RustDesk (mặc định là `C:\Program Files\RustDesk\rustdesk.exe`).
3. Thực thi lệnh deploy:
   ```cmd
   rustdesk.exe --deploy --token <API_TOKEN_CỦA_USER>
   ```
4. **Tùy chọn thiết lập ID**: Thêm tham số `--id` để đặt ID tùy chọn nếu muốn:
   ```cmd
   rustdesk.exe --deploy --token <API_TOKEN_CỦA_USER> --id <ID_TỰ_CHỌN>
   ```

##### Trên Linux / macOS (Terminal):
Chạy lệnh bằng quyền `sudo` hoặc root:
```bash
sudo rustdesk --deploy --token <API_TOKEN_CỦA_USER>
```
Hoặc cấu hình kèm ID:
```bash
sudo rustdesk --deploy --token <API_TOKEN_CỦA_USER> --id <ID_TỰ_CHỌN>
```

#### Bước C: Xác nhận kết quả
- Màn hình Command Line sẽ in ra dòng chữ: `Device deployed.`
- Trên giao diện Web Admin, thiết bị sẽ xuất hiện trong mục **Thiết bị cá nhân (My Peer)**.
- Trạng thái kết nối của Client chuyển sang **Ready (Sẵn sàng)** và sẵn sàng cho việc kết nối từ xa.

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

---
---

# RustDesk Server & Synced API Server (Self-Hosted) Project

This project provides a complete self-hosted integration stack to run **RustDesk** remote control services with full features:
- **RustDesk Server (hbbs & hbbr)**: Official ID/Rendezvous and Relay services.
- **RustDesk API Server**: An API service backend handling accounts, Address Book synchronization, device groups, and a Web Admin panel.
- **Web Admin Frontend**: An intuitive user interface and control panel.

The project is structured as a parent Git repository linking component repositories via **Git Submodules**, making it easy to fetch updates and security patches from upstream sources.

---

## 📂 Project Structure

```text
D:\projects\RustDesk\
├── rustdesk-server/            # Submodule: Official ID/Relay Server (Rust)
├── rustdesk-api/               # Submodule: Go API Server (Forked from lejianwen/rustdesk-api)
├── rustdesk-api-web/           # Submodule: Vue 3 Web Admin Dashboard
├── rustdesk-server-lejianwen/  # Submodule: Reference implementation for authentication
├── data/                       # [Git Ignored] Operation data (Databases, Keys, Logs)
├── .env                        # [Git Ignored] Environment configurations containing credentials
├── .gitignore                  # Git ignore rules
├── Dockerfile                  # Container build instructions for rustdesk-api (Go + Vue)
├── Dockerfile.server           # Container build instructions for hbbs/hbbr (Rust Upstream)
└── docker-compose.yml          # Docker compose file managing all services
```

## 🌐 Architecture & Deployment Strategy

The project supports a **Single-Node Dockerized Stack** model to optimize hardware resources while ensuring security and simple maintenance.

### 1. Service Flow Architecture
* **Nginx Reverse Proxy**: Acts as the single entrance for HTTP (`80`) and HTTPS (`443`) traffic, handles SSL termination, and routes traffic:
  - API & Web Admin traffic is routed to `rustdesk-api` (Port `21114`).
  - WebSocket client traffic (for Web Client connections) is routed to `hbbs`/`hbbr` (Port `21118`/`21119`).
* **hbbs (ID Server)**: Controls connection setup, NAT traversal (Punch Hole), and client registrations.
* **hbbr (Relay Server)**: Relays screen data when direct client connections fail or force-relay is configured.
* **rustdesk-api**: Processes login logic, JWT signing, Address Book synchronization, device groups, and serves the Web Admin UI.

### 2. Firewall Settings (Ports to Open)
For smooth operation (especially rendezvous UDP traffic and relay connections), please open these ports on your host firewall:

| Port | Protocol | Service | Purpose |
|---|---|---|---|
| **80** / **443** | TCP | Nginx Proxy | Web Admin, REST API & Web Client access |
| **21115** | TCP | hbbs | hbbs control port |
| **21116** | TCP | hbbs | ID query port |
| **21116** | **UDP** | hbbs | Rendezvous/NAT Punch Hole (Mandatory UDP) |
| **21117** | TCP | hbbr | hbbr relay control port |
| **21118** | TCP | hbbs | WebSocket ID port (for Web Client) |
| **21119** | TCP | hbbr | WebSocket Relay port (for Web Client) |

### 3. Encryption Keys & Access Security
* **Public Key Encryption**:
  - Upon first boot, `hbbs` automatically generates a secure keypair under `./data/server/`.
  - `rustdesk-api` mounts this folder as read-only (`/data:ro`) to read the public key file `id_ed25519.pub` and automatically supply it to logging-in PC clients, eliminating manual client key setups.
* **JWT & Forced Authentication (MUST_LOGIN)**:
  - The `JWT_SECRET` key is shared between `hbbs` and `rustdesk-api` to verify user logins.
  - When `MUST_LOGIN=Y` is enabled in `.env`, clients must authenticate with their account credentials (synchronized via the API Server) before being assigned an ID or starting remote sessions.

---

## 🛠️ Installation & Setup Guide

### 1. Prepare Environment variables
Copy or create a `.env` file in the root directory:
```env
# Public IP or Domain of your host server (e.g. 192.168.1.100 or rustdesk.example.com)
DOMAIN=127.0.0.1

# Timezone
TZ=Asia/Ho_Chi_Minh

# JWT secret key for signing login tokens. Change this to a secure random string!
JWT_SECRET=super_secret_jwt_sign_key_change_me
```

### 2. Launch using Docker Compose
Execute the following command in the root folder to build and run all services in the background:
```bash
docker compose up -d --build
```
*Note: The first launch might take several minutes as Docker downloads the Rust build image and compiles the upstream `hbbs`/`hbbr` source code.*

Once running:
- Access the Web Admin dashboard at: `http://<Your_Server_IP>:21114/_admin/`
- Default login username is `admin`, and the random password will be shown in the start logs of the `rustdesk-api` container.

---

## 💻 Client / Device Deployment Guide

The **Device Deployment** feature applies when the server requires device authentication before registering them (preventing unknown client IDs from arbitrary server registry). When activated, client terminals will display a warning requiring deployment.

To deploy RustDesk host clients on target devices, you have two options:

### Method 1: Automated Script via PowerShell (Recommended for Windows)
This is the fastest and fully automated method for Windows-based clients:
1. Log in to the Web Admin dashboard.
2. Open the **Client Setup & Downloads** menu in the sidebar.
3. Under the **Automated Setup & Config (Windows)** card, check **Embed current account into script** if you want to assign the device to your account automatically.
4. Copy the generated single-line PowerShell command.
5. Open **PowerShell as Administrator (Run as Administrator)** on the target Windows machine, paste the command, and press **Enter**.
6. The script will automatically:
   - Download the official RustDesk client and perform a silent installation.
   - Configure your private server settings (ID Server, Relay Server, API Server, and Public Key).
   - Start the client service and generate a secure, random 12-character static password for unattended access.
   - Send authentication details and deploy requests to the API Server, registering the host under your account.
   - Encrypt and sync the password to your personal **Address Book** under the host computer name.

### Method 2: Manual CLI Deployment
To perform a manual deployment or set up Linux/macOS clients, please follow these steps:

#### Step A: Obtain User API Token
1. Log in to the Web Admin dashboard with the target user account.
2. Go to **Profile** to retrieve the **API Token** assigned to the account.

#### Step B: Execute Deploy Command on Client Device
Run the RustDesk client executable from the terminal with Administrator/root privileges.

##### On Windows (PowerShell/CMD as Admin):
1. Open PowerShell or Command Prompt as Administrator.
2. Navigate to the RustDesk installation folder (default is `C:\Program Files\RustDesk\rustdesk.exe`).
3. Run the deploy command:
   ```cmd
   rustdesk.exe --deploy --token <USER_API_TOKEN>
   ```
4. **Optional custom ID**: Append the `--id` parameter to assign a custom ID:
   ```cmd
   rustdesk.exe --deploy --token <USER_API_TOKEN> --id <CUSTOM_ID>
   ```

##### On Linux / macOS (Terminal):
Run with `sudo` or root privileges:
```bash
sudo rustdesk --deploy --token <USER_API_TOKEN>
```
Or with a custom ID:
```bash
sudo rustdesk --deploy --token <USER_API_TOKEN> --id <CUSTOM_ID>
```

#### Step C: Verify Connection
- The command line will output: `Device deployed.`
- In the Web Admin dashboard, the client device will immediately appear under **My Devices** (My Peer).
- The client status will change to **Ready**, allowing passwordless remote control connections.

---

## 🔄 Upstream Repository Updates (Security Patches)

Since component projects are linked as Git Submodules, you can easily pull updates and security patches from their original authors:

### 1. Update all submodules to the latest commits
Run the following from the project root:
```bash
git submodule update --remote --merge
```

### 2. Manually update a specific submodule
For instance, to update the upstream RustDesk server:
```bash
cd rustdesk-server
git checkout master
git pull origin master
cd ..
# Commit the updated submodule pointer to the parent repository
git add rustdesk-server
git commit -m "Update rustdesk-server submodule to latest release"
```

### 3. Rebuild the stack
After pulling the updates, rebuild and restart your Docker containers:
```bash
docker compose up -d --build
```

---

## 📝 Parent Git Repository Management

To push this self-hosted unified repository structure to your private GitHub repository:

```bash
# 1. Initialize Git repository
git init

# 2. Add configuration files and submodule registers
git add .gitignore README.md docker-compose.yml Dockerfile Dockerfile.server .gitmodules

# 3. Add submodules as gitlinks (Do NOT add their files directly)
git add rustdesk-server rustdesk-api rustdesk-api-web rustdesk-server-lejianwen

# 4. Commit initial project setup
git commit -m "Initial commit: Unified RustDesk Stack with submodules"

# 5. Push to your private remote GitHub repository
git remote add origin <YOUR_GITHUB_REPOSITORY_URL>
git branch -M main
git push -u origin main
```
