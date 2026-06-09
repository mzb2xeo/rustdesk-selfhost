# RustDesk API Gap Implementation Tasks

Tai lieu nay ghi lai cac API con thieu sau khi doi chieu:

- Client: `D:\projects\rustdeskclient`
- API backend: `rustdesk-api`
- Web Admin: `rustdesk-api-web`
- RustDesk server: `rustdesk-server`

## Nguyen tac thuc hien

- Uu tien API dang anh huong truc tiep den client va Web Admin.
- Tat ca API thay doi du lieu phai xac thuc bang `middleware.RustAuth()` hoac middleware admin phu hop.
- Kiem tra quyen so huu thiet bi/audit record, khong chi kiem tra token hop le.
- Khong tin tuong `id`, `uuid`, `guid` hoac `user_id` do client gui len.
- Them migration/model, controller, service, router va test cho tung nhom API.
- Cap nhat Swagger sau khi hoan thanh endpoint.

### Phase 1 - Sua Web Admin command update

### TASK-001: Dang ky route `cmdUpdate`

- [x] Them route:
  - `POST /api/admin/rustdesk/cmdUpdate`
  - Handler: `admin.Rustdesk.CmdUpdate`
- [x] Xac nhan route nam sau `middleware.BackendUserAuth()`.
- [x] Xac nhan quyen admin truoc khi cho phep sua command.
- [x] Test tao command, sua command, doc lai danh sach va xoa command.
- [x] Kiem tra thao tac sua command tren Web Admin khong con tra ve `404`.

Vi tri lien quan:

- `rustdesk-api/http/router/admin.go`
- `rustdesk-api/http/controller/admin/rustdesk.go`
- `rustdesk-api-web/src/api/rustdesk.js`
- `rustdesk-api-web/src/views/rustdesk/control.vue`

Tieu chi nghiem thu:

- Web Admin cap nhat command thanh cong.
- User khong co quyen admin khong the cap nhat command.

## Phase 2 - Hoan thien Audit API cho client

### TASK-002: Them audit connection GUID

- [x] Them truong GUID duy nhat vao audit connection.
- [x] Tao GUID khi nhan `POST /api/audit/conn` voi `action = new`.
- [x] Them index database cho GUID.
- [x] Dam bao migration khong lam mat audit record hien tai.

Vi tri lien quan:

- `rustdesk-api/model/audit.go`
- `rustdesk-api/http/request/api/audit.go`
- `rustdesk-api/http/controller/api/audit.go`
- `rustdesk-api/service/audit.go`

### TASK-003: Them API truy van active audit connection

- [x] Them route co xac thuc:
  - `GET /api/audit/conn/active?id={peer_id}&session_id={session_id}&conn_type={type}`
- [x] Chi tra ve GUID cua connection dang active va thuoc user hien tai.
- [x] Tra ve JSON string GUID de phu hop client.
- [x] Khong cho phep user truy van audit connection cua user khac.
- [x] Them test: tim thay, khong tim thay, sai user, token khong hop le.

Client tham chieu:

- `D:\projects\rustdeskclient\flutter\lib\models\model.dart`

### TASK-004: Them API cap nhat audit note

- [x] Them truong `note` vao audit connection.
- [x] Them route co xac thuc:
  - `PUT /api/audit`
- [x] Payload:

```json
{
  "guid": "audit-guid",
  "note": "connection note"
}
```

- [x] Chi cho phep cap nhat record thuoc user hien tai.
- [x] Gioi han do dai note va validate payload.
- [x] Them test cap nhat thanh cong, GUID sai va sai user.

Client tham chieu:

- `D:\projects\rustdeskclient\flutter\lib\common\widgets\dialog.dart`

### TASK-005: Them audit alarm

- [x] Xac dinh danh sach alarm type client dang gui.
- [x] Tao model va migration cho audit alarm.
- [x] Them route:
  - `POST /api/audit/alarm`
- [x] Luu `id`, `uuid`, `typ`, `info`, IP va thoi gian.
- [x] Them danh sach alarm trong Web Admin neu can quan ly qua UI.
- [x] Them gioi han tan suat de tranh spam alarm.

Client tham chieu:

- `D:\projects\rustdeskclient\src\server\connection.rs`
- `D:\projects\rustdeskclient\src\server\login_failure_check.rs`

Tieu chi nghiem thu Phase 2:

- Client lay duoc audit GUID sau khi ket noi.
- Client cap nhat note thanh cong khi ket thuc ket noi.
- Alarm tu client duoc luu ma khong cho phep gia mao user.

## Phase 3 - Device deployment va CLI management

### TASK-006: Them `POST /api/devices/deploy`

- [x] Xac dinh contract chinh xac tu client:

```json
{
  "id": "device-id",
  "uuid": "base64-uuid",
  "pk": "base64-public-key"
}
```

- [x] Xac thuc `Authorization: Bearer <token>`.
- [x] Kiem tra token co quyen deploy device.
- [x] Dang ky hoac cap nhat device theo `id`, `uuid` va public key.
- [x] Xu ly tranh chiem device ID cua user khac.
- [x] Tra ve response client mong doi:

```json
{
  "result": "OK"
}
```

- [x] Them test deploy moi, deploy lai, ID trung, token sai va payload sai.

Client tham chieu:

- `D:\projects\rustdeskclient\src\ui_interface.rs`

### TASK-007: Them `POST /api/devices/cli`

- [x] Liet ke day du tham so CLI client gui.
- [x] Xac thuc bearer token va quyen quan ly device.
- [x] Ho tro cac thao tac client CLI dang yeu cau.
- [x] Validate device ID, group, address book va owner.
- [x] Tra ve loi ro rang thay vi response rong khi thao tac that bai.
- [x] Them test cho tung thao tac CLI duoc ho tro.

Client tham chieu:

- `D:\projects\rustdeskclient\src\core_main.rs`

Tieu chi nghiem thu Phase 3:

- Lenh deploy cua client tra ve `OK` va device xuat hien trong Web Admin.
- CLI khong the sua device cua user khac.

## Phase 4 - Danh gia va bo sung API nang cao

### TASK-008: Danh gia `POST /api/record`

- [ ] Xac dinh co bat tinh nang upload recording tren client hay khong.
- [ ] Neu can, thiet ke storage, authentication, quota va cleanup policy.
- [ ] Ho tro cac request type cua client: `new`, chunk upload va ket thuc/xoa.
- [ ] Khong bat tinh nang neu chua co gioi han dung luong va bao mat storage.

Client tham chieu:

- `D:\projects\rustdeskclient\src\hbbs_http\record_upload.rs`

### TASK-009: Danh gia heartbeat strategy response

- [ ] Xac dinh co can day policy tu server xuong client hay khong.
- [ ] Neu can, bo sung response cho `POST /api/heartbeat`:
  - `strategy`
  - `modified_at`
  - `disconnect`
- [ ] Them versioning va validate strategy options.
- [ ] Dam bao heartbeat van tuong thich voi client cu.

Client tham chieu:

- `D:\projects\rustdeskclient\src\hbbs_http\sync.rs`

### TASK-010: Xac dinh pham vi plugin signing

- [ ] Quyet dinh co ho tro `/lic/web/api/plugin-sign` hay khong.
- [ ] Neu khong ho tro, ghi ro day la tinh nang ngoai pham vi/community.
- [ ] Khong tao endpoint gia neu khong co quy trinh ky va quan ly khoa an toan.

## Phase 5 - Don dep contract Web Admin

### TASK-011: Ra soat API khai bao nhung backend chua co

- [ ] Xac minh endpoint nao con duoc UI su dung:
  - `/address_book/detail/:id`
  - `/user/myPeer`
  - `/address_book_collection_rule/batchCreate`
  - `/my/address_book_collection_rule/batchCreate`
  - `/file/oss_token`
- [ ] Neu UI khong su dung, xoa API wrapper khoi `rustdesk-api-web`.
- [ ] Neu UI can su dung, them controller/router va test tuong ung.
- [ ] Khong bat lai file upload routes neu chua co storage configuration an toan.

## Phase 6 - Dong bo rustdesk-server voi client

Phat hien khi doi chieu:

- Client co version `1.4.7`, trong khi `rustdesk-server` dang khai bao version `1.1.15`.
- Client va server khong dung chung mot checkout `hbb_common`.
- Thu muc `D:\projects\rustdeskclient\libs\hbb_common` dang rong, thieu ca
  `Cargo.toml` va protobuf source, nen khong the build/kiem chung client sach.
- Client dang tham chieu cac protobuf field va enum ma
  `rustdesk-server/libs/hbb_common/protos/rendezvous.proto` chua dinh nghia.

### TASK-012: Khoi phuc va khoa version `hbb_common` cua client

- [x] Khoi tao submodule `D:\projects\rustdeskclient\libs\hbb_common`.
- [x] Ghi lai commit SHA cua `hbb_common` ma client dang yeu cau:
  `387603f47cbb15c0d3dc3d67ae3396d3eb707daf`.
- [x] Dam bao `git submodule update --init --recursive` hoat dong trong moi
  truong build.
- [x] Them kiem tra CI de fail som neu `libs/hbb_common/Cargo.toml` hoac protobuf
  source bi thieu.
- [x] Chay `cargo check -p hbb_common` cho client sau khi khoi phuc submodule
  bang Rust `1.96` tren Linux.

Tieu chi nghiem thu:

- Checkout moi co the build client ma khong can file local khong duoc commit.
- Commit `hbb_common` cua client duoc ghi ro va co the tai lap.

### TASK-013: Dong bo `rendezvous.proto` giua client va server

- [x] Lay protobuf source thuc te tu commit `hbb_common` cua client.
- [x] So sanh va cap nhat
  `rustdesk-server/libs/hbb_common/protos/rendezvous.proto`.
- [x] Giu nguyen field number cu; chi them field moi voi field number dung tu
  client de dam bao backward compatibility.
- [x] Dong bo it nhat cac field/enum client dang tham chieu:
  - `RegisterPk.no_register_device`
  - `RegisterPkResponse.Result.NOT_DEPLOYED`
  - `PunchHoleRequest.udp_port`
  - `PunchHoleRequest.force_relay`
  - `PunchHoleRequest.socket_addr_v6`
  - `PunchHole.force_relay`
  - `PunchHoleResponse.is_udp`
  - `PunchHoleResponse.socket_addr_v6`
  - `RelayResponse.socket_addr_v6`
- [x] Ra soat them tat ca field moi sau khi submodule client duoc khoi phuc.
- [x] Generate lai protobuf code cho ca client va server thong qua build/check.
- [x] Them protocol compatibility test giua client proto va server proto.

Vi tri lien quan:

- `rustdesk-server/libs/hbb_common/protos/rendezvous.proto`
- `D:\projects\rustdeskclient\libs\hbb_common\protos\rendezvous.proto`
- `D:\projects\rustdeskclient\src\client.rs`
- `D:\projects\rustdeskclient\src\rendezvous_mediator.rs`

Tieu chi nghiem thu:

- Server doc duoc request tu client `1.4.7` ma khong bo qua cac field can thiet.
- Client cu van dang ky, punch hole va relay duoc sau khi server cap nhat.

Ket qua kiem chung:

- SHA256 `rendezvous.proto` client/server giong nhau:
  `1A79E328400D8D7D99D42A4DD562BA062EDC301DEDBEE3467007A7F7E844E7A9`.
- `cargo test -p hbb_common --test rendezvous_compat` dat `2/2` tren Rust
  `1.85` va `1.96`.
- Docker release build `hbbs`/`hbbr` dat tren Rust `1.96`.

### TASK-019: Nang rustdesk-server len Rust `1.96`

- [x] Nang Rust builder trong `Dockerfile.server` len `rust:1.96-alpine`.
- [x] Dat `rust-version = "1.96"` cho server va `hbb_common`.
- [x] Nang Rust toolchain trong workflow build server len `1.96`.
- [x] Build release `hbbs` va `hbbr` thanh cong tren Linux/Rust `1.96`.
- [x] Chay protocol compatibility test tren Rust `1.96`.
- [ ] Xu ly cac warning moi/ro hon tren Rust `1.96`:
  - shared reference den mutable static va function-to-integer cast trong
    `libs/hbb_common/src/platform/mod.rs`;
  - [x] thay `array.map(...)` khong duoc su dung bang vong lap trong
    `libs/hbb_common/src/config.rs`;
  - [x] bo `mut` thua trong lenh `punch-requests` cua hbbs.
- [x] Kiem tra native Windows bang MSYS2 UCRT64 va Rust GNU toolchain.

Hien trang:

- Native Windows MSVC check ngay 2026-06-09 van bi chan boi
  `LNK1104: cannot open file 'libcmt.lib'`; khong dung MSVC cho check hien tai.
- Da cai `stable-x86_64-pc-windows-gnu` va
  `x86_64-pc-windows-gnu` bang rustup.
- Native Windows GNU check dung MSYS2 UCRT64 GCC 16.1 va Rust 1.96 dat.
- Protocol compatibility test tren native Windows GNU dat `2/2`.
- `cargo check --bin hbbs` tren Linux image test Rust 1.96 dat.

Lenh native Windows GNU:

```powershell
$env:PATH='C:\msys64\ucrt64\bin;' + $env:PATH
$env:CARGO_TARGET_DIR='target-gnu-host'
rtk cargo +stable-x86_64-pc-windows-gnu check --bin hbbs
```

### TASK-020: Tao image Rust test dung lai

- [x] Tao `Dockerfile.test` chua dependency build server va client.
- [x] Tao `docker-compose.test.yml` mount hai repo va dung named volumes cho
  Cargo registry, Cargo git va target cache.
- [x] Build image `rustdesk-rust-test:1.96`.
- [x] Chay protocol compatibility test bang image test.
- [x] Chay `cargo check -p hbb_common` cua client bang image test.

Ket qua:

- Lan dau protocol test tao cache va dat `2/2`.
- Lan chay lai hoan tat trong khoang `1.6s`; Cargo build/test mat `0.51s`.
- Client `cargo check -p hbb_common` dat sau khi tao client target cache.
- Full client check dung `linux-pkg-config`; image bo sung `libyuv.pc` vi
  package Debian `libyuv-dev` khong cung cap file nay.

Cach dung:

```powershell
docker compose -f docker-compose.test.yml build rust-test
docker compose -f docker-compose.test.yml run --rm rust-test
docker compose -f docker-compose.test.yml run --rm -w /workspace/rustdeskclient rust-test cargo check -p hbb_common
docker compose -f docker-compose.test.yml run --rm -w /workspace/rustdeskclient rust-test cargo check -p rustdesk --features linux-pkg-config
```

### TASK-014: Ho tro UDP NAT punch va IPv6 theo client

- [x] Xu ly `PunchHoleRequest.udp_port` tren hbbs.
- [x] Tra ve `PunchHoleResponse.is_udp` khi UDP punch duoc chon.
- [x] Chuyen tiep `socket_addr_v6` giua hai client khi co IPv6.
- [x] Chuyen tiep `RelayResponse.socket_addr_v6` neu relay negotiation can.
- [x] Ton trong `force_relay` do client gui, thay vi chi dua vao server command
  `always-use-relay`.
- [x] Xac nhan client da co fallback TCP/relay khi UDP hoac IPv6 that bai.
- [ ] Test cac truong hop:
  - TCP direct
  - UDP direct
  - IPv6 direct
  - force relay
  - WebSocket relay

Hien trang:

- Server chuyen tiep `udp_port`, `force_relay`, `upnp_port` va
  `socket_addr_v6` tu requester sang peer.
- Server chuyen tiep IPv6/UPnP trong punch response, local-address response va
  giu `RelayResponse.socket_addr_v6`.
- Server dat `PunchHoleResponse.is_udp` khi nhan `PunchHoleSent` qua UDP.
- Client chay song song request co UDP va request TCP, thu IPv6 cung relay va
  chon ket noi thanh cong dau tien; khong can them fallback o hbbs.
- `cargo check --bin hbbs` dat va protocol compatibility test dat `2/2`.
- Integration test voi hai client that cho TCP/UDP/IPv6/relay se duoc kiem
  thu tay va cap nhat ket qua sau; khong dong muc nay bang automated check.

### TASK-015: Dong bo luong bat buoc deploy device

- [x] Quyet dinh ro che do van hanh:
  - cho phep client tu dang ky nhu server hien tai; hoac
  - bat buoc deploy truoc khi client duoc dang ky.
- [x] Ton trong `RegisterPk.no_register_device` trong che do hien tai: tra `OK`
  de client xac nhan key, nhung khong tao/cap nhat peer trong hbbs.
- [ ] Neu bat buoc deploy:
  - Khi `RegisterPk`, hbbs phai kiem tra device trong database/API.
  - Tra `RegisterPkResponse.Result.NOT_DEPLOYED` neu chua deploy.
  - Ton trong `RegisterPk.no_register_device`.
  - Sau `POST /api/devices/deploy`, client co the dang ky lai thanh cong.
  - Device bi xoa/disable phai bi tu choi dang ky lai.
- [ ] Thiet ke ket noi hbbs voi API/database ma khong tin tuong du lieu tu client.
- [ ] Them cache va timeout de API/database loi khong lam treo hbbs.
- [ ] Test end-to-end: chua deploy, deploy, disable, delete va deploy lai.

Hien trang:

- Client da co `NEEDS_DEPLOY`, retry throttling va xu ly `NOT_DEPLOYED`.
- Che do hien tai cho phep client tu dang ky; server chua tra `NOT_DEPLOYED`.
- Khong bat deploy-required truoc khi co API deploy va nguon du lieu device dang
  tin cay cho hbbs. `rustdesk-api` hien chua co route `/api/devices/deploy`.
- `no_register_device` duoc xu ly tren ca UDP va TCP RegisterPk ma khong ghi peer
  vao SQLite/in-memory map.
- `cargo check --bin hbbs` tren image test Rust 1.96 dat sau thay doi; protocol
  compatibility test dat `2/2`.

Phu thuoc:

- `TASK-006`: `POST /api/devices/deploy`
- `TASK-007`: `POST /api/devices/cli`
- `TASK-013`: dong bo protobuf

### TASK-016: Sua cau hinh JWT giua API va hbbs

- [x] Truyen cung `RUSTDESK_API_JWT_KEY=${JWT_SECRET}` vao service `hbbs`.
- [x] Xac nhan JWT do `rustdesk-api` tao co claims va HS256 signature ma hbbs
  verify duoc.
- [x] Khong cho phep bat `must-login` neu JWT secret rong.
- [x] Khi secret rong, hbbs phai fail startup hoac tu choi bat `must-login`;
  khong duoc chi kiem tra token khong rong.
- [x] Khong log JWT secret. Xoa hoac cam su dung `generate_token()` dang
  `println!` secret trong `rustdesk-server/src/jwt.rs`.
- [ ] Test:
  - token API hop le
  - token het han
  - token ky bang secret khac
  - chuoi token gia nhung khong rong
  - logout/revoke token

Hien trang:

- API va hbbs cung nhan `JWT_SECRET` qua `RUSTDESK_API_JWT_KEY`.
- API tao HS256 claims `user_id` va `exp`, phu hop verifier hbbs.
- Hbbs fail startup neu `must-login` duoc yeu cau voi secret rong, va tu choi
  lenh runtime bat `must-login` trong truong hop nay.
- `go test ./lib/jwt` dat `2/2`; `cargo check --bin hbbs` dat.
- Chua co integration test dua token API qua hbbs, token revoke khong the duoc
  hbbs nhan biet neu chi verify JWT stateless.
- Hbbr khong can access token, nhung van phai tiep tuc kiem tra server key.

### TASK-017: Dong bo health-check/feedback contract

- [x] Xac dinh co su dung persistent health-check connection hay khong.
- [x] Khong bat persistent health-check tren server hien tai; giu `feedback=0`.
- [ ] Neu sau nay bat, hbbs phai tra `feedback` khac `0` khi muon client mo
  `HealthCheck`.
- [ ] Neu sau nay bat, them handler cho `RendezvousMessage.HealthCheck`, validate
  token va gan ket noi HC voi dung phien dang chay.

Hien trang:

- Client co the gui `HealthCheck` khi response co `feedback`.
- Server proto co `HealthCheck`, nhung khong thay handler va hien khong set
  `feedback`, nen luong nay dang duoc vo hieu hoa an toan.
- Khong bat `feedback` chi de tao heartbeat: client spawn HC tach roi, bo qua loi
  HC va khong dong phien remote dang chay khi HC mat ket noi/token het han.
- Chi nen mo lai contract nay khi server co session ID/state de revoke hoac dong
  dung phien remote. Khong can xoa code client vi client van can tuong thich voi
  server khac co ho tro contract nay.

### TASK-018: Them test matrix client-server

- [ ] Tao smoke test khoi dong `hbbs`, `hbbr`, API va hai client test.
- [ ] Test dang ky ID va public key.
- [ ] Test direct TCP, UDP, IPv6 va relay.
- [ ] Test WebSocket qua cong `21118` va `21119`.
- [ ] Test `must-login` voi JWT that/gia/het han.
- [ ] Test deploy-required neu che do nay duoc bat.
- [ ] Test client version hien tai va it nhat mot client cu.
- [ ] Ghi ro protocol/version matrix duoc ho tro.

## Kiem thu tong the

- [ ] Chay Go test cho `rustdesk-api`.
- [ ] Build `rustdesk-api`.
- [ ] Build `rustdesk-api-web`.
- [ ] Build `rustdesk-server`.
- [ ] Build client tu checkout sach voi submodule day du.
- [ ] Test login/logout tu RustDesk client.
- [ ] Test address book, users, peers va device-group.
- [ ] Test audit connection, active GUID, note va alarm.
- [ ] Test Web Admin command create/update/delete.
- [ ] Test deploy va CLI voi token dung/sai.
- [ ] Test rendezvous/relay protocol matrix giua server va client.
- [ ] Cap nhat Swagger va tai lieu API.

## Thu tu de xuat

1. `TASK-001` - Sua `cmdUpdate` vi controller da co san va Web Admin dang bi `404`.
2. `TASK-002` den `TASK-005` - Hoan thien audit contract client dang su dung.
3. `TASK-006` va `TASK-007` - Ho tro deploy va device CLI.
4. `TASK-012`, `TASK-013` va `TASK-016` - Khoi phuc build, dong bo protobuf va
   dong bo JWT truoc khi mo rong server-client.
5. `TASK-014`, `TASK-015`, `TASK-017` va `TASK-018` - Hoan thien va kiem thu
   protocol server-client.
6. `TASK-011` - Don dep contract Web Admin.
7. `TASK-008` den `TASK-010` - Chi trien khai khi xac nhan nhu cau van hanh.
