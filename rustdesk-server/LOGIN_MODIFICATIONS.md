# RustDesk Server - Tài liệu chỉnh sửa tính năng Xác thực đăng nhập (JWT Auth)

Tài liệu này ghi nhận các thay đổi được thực hiện trên mã nguồn máy chủ `rustdesk-server` (gốc) để hỗ trợ xác thực mã JWT Token từ API Go. Khi tiến hành cập nhật mã nguồn máy chủ từ thượng nguồn (upstream), bạn có thể tham chiếu tài liệu này để ghép (merge) hoặc áp dụng lại phần login.

---

## 1. File thêm mới: `src/jwt.rs`
Tạo file mới tại đường dẫn `src/jwt.rs` với nội dung xử lý giải mã và xác thực token:

```rust
use jsonwebtoken::{decode, encode, Algorithm, DecodingKey, EncodingKey, Header, Validation};
use once_cell::sync::Lazy;
use serde::{Deserialize, Serialize};
use std::env;

pub static SECRET: Lazy<String> =
    Lazy::new(|| env::var("RUSTDESK_API_JWT_KEY").unwrap_or_else(|_| "".to_string()));

#[derive(Debug, Serialize, Deserialize)]
pub struct Claims {
    user_id: u32,
    exp: usize,
}

pub fn generate_token(user_id: u32, exp: i64) -> Result<String, String> {
    println!("secret: {:}", SECRET.to_string());
    let claims = Claims {
        user_id,
        exp: (chrono::Utc::now() + chrono::Duration::seconds(exp)).timestamp() as usize,
    };

    let token = encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(SECRET.as_ref()),
    );

    match token {
        Ok(t) => Ok(t),
        Err(e) => Err(e.to_string()),
    }
}

pub fn verify_token(token: &str) -> Result<Claims, String> {
    let validation = Validation::new(Algorithm::HS256);

    let decoded = decode::<Claims>(
        token,
        &DecodingKey::from_secret(SECRET.as_ref()),
        &validation,
    );
    match decoded {
        Ok(token_data) => {
            let now = chrono::Utc::now().timestamp() as usize;
            if token_data.claims.exp > now {
                Ok(token_data.claims)
            } else {
                Err("Token status invalid or expired".to_string())
            }
        }
        Err(_) => Err("Invalid token".to_string()),
    }
}
```

---

## 2. File chỉnh sửa: `src/lib.rs`
Khai báo module `jwt` vừa tạo:

```diff
 mod database;
 mod peer;
 mod version;
+pub mod jwt;
```

---

## 3. File chỉnh sửa: `src/main.rs`
Thêm tùy chọn dòng lệnh `--must-login` vào danh sách tham số của `hbbs`:

```diff
         , --mask=[MASK] 'Determine if the connection comes from LAN, e.g. 192.168.0.0/16'
-        -k, --key=[KEY] 'Only allow the client with the same key'",
+        -k, --key=[KEY] 'Only allow the client with the same key'
+        , --must-login=[Y|N] 'Only allow the client with login'",
```

---

## 4. File chỉnh sửa: `src/rendezvous_server.rs`
Thực hiện các chỉnh sửa sau:

### 4.1. Import module và định nghĩa biến cấu hình
```diff
 use crate::common::*;
 use crate::peer::*;
+use crate::jwt;
...
 type RelayServers = Vec<String>;
 const CHECK_RELAY_TIMEOUT: u64 = 3_000;
 static ALWAYS_USE_RELAY: AtomicBool = AtomicBool::new(false);
+static MUST_LOGIN: AtomicBool = AtomicBool::new(false);
```

### 4.2. Quét tham số cấu hình bật/tắt tính năng đăng nhập tại hàm `start()`
```diff
         log::info!(
             "ALWAYS_USE_RELAY={}",
             if ALWAYS_USE_RELAY.load(Ordering::SeqCst) {
                 "Y"
             } else {
                 "N"
             }
         );
+        let must_login = get_arg("must-login");
+        if must_login.to_uppercase() == "Y"
+            || (must_login == ""
+                && std::env::var("MUST_LOGIN")
+                    .unwrap_or_default()
+                    .to_uppercase()
+                    == "Y")
+        {
+            MUST_LOGIN.store(true, Ordering::SeqCst);
+        }
+        log::info!(
+            "MUST_LOGIN={}",
+            if MUST_LOGIN.load(Ordering::SeqCst) {
+                "Y"
+            } else {
+                "N"
+            }
+        );
```

### 4.3. Xác thực Token khi xử lý yêu cầu kết nối tại `handle_punch_hole_request()`
```diff
         if !key.is_empty() && ph.licence_key != key {
             log::warn!("Authentication failed from {} for peer {} - invalid key", addr, ph.id);
             let mut msg_out = RendezvousMessage::new();
             msg_out.set_punch_hole_response(PunchHoleResponse {
                 failure: punch_hole_response::Failure::LICENSE_MISMATCH.into(),
                 ..Default::default()
             });
             return Ok((msg_out, None));
         }
+        if MUST_LOGIN.load(Ordering::SeqCst) {
+            if ph.token.is_empty() {
+                let mut msg_out = RendezvousMessage::new();
+                msg_out.set_punch_hole_response(PunchHoleResponse {
+                    other_failure: String::from("Connection failed, please login!"),
+                    ..Default::default()
+                });
+                return Ok((msg_out, None));
+            } else if !jwt::SECRET.is_empty() {
+                let token = ph.token;
+                let token = jwt::verify_token(token.as_str());
+                if token.is_err() {
+                    let mut msg_out = RendezvousMessage::new();
+                    msg_out.set_punch_hole_response(PunchHoleResponse {
+                        other_failure: String::from("Token error, please log out and log back in!"),
+                        ..Default::default()
+                    });
+                    return Ok((msg_out, None));
+                }
+            }
+        }
         let id = ph.id;
```

### 4.4. Thêm hiển thị trợ giúp lệnh ml / must-login trong terminal quản trị hbbs
```diff
                 res = format!(
-                    "{}\n{}\n{}\n{}\n{}\n{}\n{}\n",
+                    "{}\n{}\n{}\n{}\n{}\n{}\n{}\n{}\n",
                     "relay-servers(rs) <separated by ,>",
                     "reload-geo(rg)",
                     "ip-blocker(ib) [<ip>|<number>] [-]",
                     "ip-changes(ic) [<id>|<number>] [-]",
                     "punch-requests(pr) [<number>] [-]",
                     "always-use-relay(aur)",
-                    "test-geo(tg) <ip1> <ip2>"
+                    "test-geo(tg) <ip1> <ip2>",
+                    "must-login(ml) [Y|N]"
                 )
```

### 4.5. Thêm lệnh quản trị động `must-login` / `ml` tại CLI runtime
```diff
             Some("always-use-relay" | "aur") => {
...
             }
+            Some("must-login" | "ml") => {
+                if let Some(rs) = fds.next() {
+                    if rs.to_uppercase() == "Y" {
+                        MUST_LOGIN.store(true, Ordering::SeqCst);
+                    } else {
+                        MUST_LOGIN.store(false, Ordering::SeqCst);
+                    }
+                } else {
+                    let _ = writeln!(res, "MUST_LOGIN: {:?}", MUST_LOGIN.load(Ordering::SeqCst));
+                }
+            }
             Some("test-geo" | "tg") => {
```
