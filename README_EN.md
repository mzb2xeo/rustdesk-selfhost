# RustDesk Self-Hosted Stack

A self-hosted RustDesk deployment stack that packages the ID Server, Relay Server, API Server, and Web Admin into one Docker Compose setup. This repository is intended for private/internal RustDesk operations with user accounts, address book synchronization, automatic client configuration, and Windows device deployment.

> This is an integration/custom fork for private deployment. It is not an official RustDesk release.

Documentation: [Tiếng Việt](README.md) | **English**

## Main Components

- `hbbs`: ID/Rendezvous server for device registration and NAT traversal.
- `hbbr`: Relay server used when direct connections are unavailable.
- `rustdesk-api`: Admin API for authentication, server configuration, address book, and deploy tokens.
- `rustdesk-api-web`: Web Admin UI for users, devices, address book, client configuration, and deploy commands.
- `reverse-proxy`: Nginx entrypoint for API, Web Admin, and WebSocket traffic.

## Upstream Sources And Forks

| Component | Description | Fork source | Base version | Changes from upstream/base |
|---|---|---|---|---|
| `hbbs` | RustDesk ID/Rendezvous server | `https://github.com/rustdesk/rustdesk-server` | `1.1.15` | Integrates `MUST_LOGIN` checks, shares `JWT_SECRET` with the API, and supports the required login/deployment flow before device registration. |
| `hbbr` | RustDesk relay server | `https://github.com/rustdesk/rustdesk-server` | `1.1.15` | Packaged into the same Docker stack, shares the `hbbs` network namespace, and uses stack-level relay/domain configuration. |
| `rustdesk-api` | API server | `https://github.com/lejianwen/rustdesk-api` | `2.7` | Adds short-lived deploy tokens, a PowerShell deploy-script endpoint, deploy-token auth for `/api/devices/deploy` and `/api/devices/cli`, `hbbs` public-key discovery, and automatic client server configuration. |
| `rustdesk-api-web` | Web Admin | `https://github.com/lejianwen/rustdesk-api` | `2.7` | Adds `My -> Client Config`, client configuration display, self-downloading deploy command generation, script download, command copy actions, and token expiration metadata. |
| Docker/ops in this repo | Local integration/custom fork | Local working tree | Based on the sources above | Adds `docker-compose.yml`, `Dockerfile`, `Dockerfile.server`, `nginx.conf`, `deploy-host.ps1`, shared data volumes, and operational documentation for self-hosted deployment. |

When updating from upstream/forks, pay attention to these boundaries:

- Updating `rustdesk-server` can affect device registration, `MUST_LOGIN`, WebSocket behavior, and keypair handling.
- Updating `rustdesk-api` can affect DB schema, tokens, admin/API routes, and address book behavior.
- Updating `rustdesk-api-web` can affect menus, i18n, and Web Admin endpoint calls.

## Repository Layout

```text
.
├── docker-compose.yml          # Main runtime stack
├── nginx.conf                  # Reverse proxy for API/WebSocket
├── Dockerfile                  # Builds rustdesk-api + rustdesk-api-web
├── Dockerfile.server           # Builds hbbs/hbbr
├── deploy-host.ps1             # Windows wrapper that downloads and runs the API deploy script
├── rustdesk-server/            # RustDesk server / hbbs / hbbr source
├── rustdesk-api/               # Go API server
├── rustdesk-api-web/           # Vue Web Admin
└── data/                       # Runtime data: database, keys, logs
```

This repository currently does not use `.gitmodules`; component folders are separate working trees. Check their remotes with:

```bash
git -C rustdesk-server remote -v
git -C rustdesk-api remote -v
git -C rustdesk-api-web remote -v
```

## Architecture

```text
RustDesk Client
   |
   ├── 21116/tcp+udp --> hbbs: ID / rendezvous / NAT punch
   ├── 21117/tcp     --> hbbr: relay
   └── HTTPS/WSS     --> reverse-proxy --> rustdesk-api / Web Admin / WebSocket

hbbs/hbbr   --> ./data/server
rustdesk-api --> ./data/api
rustdesk-api --> read-only ./data/server/id_ed25519.pub
```

### Service Ports

| Port | Protocol | Component | Purpose |
|---|---|---|---|
| `8082` | TCP | Nginx | HTTP entrypoint in the current compose stack |
| `21114` | TCP | API through the `hbbs` namespace | Direct API/Web Admin access |
| `21115` | TCP | `hbbs` | Control port |
| `21116` | TCP/UDP | `hbbs` | ID/Rendezvous and NAT punch |
| `21117` | TCP | `hbbr` | Relay |
| `21118` | TCP | `hbbs` | WebSocket ID for Web Client |
| `21119` | TCP | `hbbr` | WebSocket Relay for Web Client |

For production, place a TLS reverse proxy or load balancer in front of `8082`, then point the public API/Web Admin domain to it.

## Environment Variables

Create a `.env` file in the repository root:

```env
DOMAIN=rd.example.com
DOMAIN_API=rustdesk.example.com
TZ=Asia/Ho_Chi_Minh
JWT_SECRET=change_me_to_a_long_random_secret
MUST_LOGIN=N
```

Key variables:

- `DOMAIN`: domain/IP used by RustDesk clients for `hbbs` and `hbbr`.
- `DOMAIN_API`: public API/Web Admin domain used to generate client configuration and deploy-script URLs.
- `JWT_SECRET`: JWT signing key shared by `hbbs` and `rustdesk-api`.
- `MUST_LOGIN=Y`: requires clients to log in/deploy before device registration is allowed.

## Startup

```bash
docker compose up -d --build
```

Check runtime status:

```bash
docker compose ps
docker compose logs -f rustdesk-api
```

Web Admin:

```text
http://<server-ip>:8082/_admin/
```

The initial admin password is printed in the `rustdesk-api` container logs.

## Windows Client Deployment

Recommended end-user flow:

1. Log in to Web Admin.
2. Open `My` -> `Client Config`.
3. Click `Generate Deploy Command`.
4. Copy the `Download script and run deploy` command.
5. Open PowerShell as Administrator on the Windows machine.
6. Paste and run the command.

The script will:

- Download the RustDesk client if it is not installed.
- Write `ID Server`, `Relay Server`, `API Server`, and public key configuration.
- Verify `RustDesk2.toml` (host/relay/api/key) and run network checks to ID/Relay/API.
- Read the device ID with `rustdesk.exe --get-id` (on Windows, pipe output: `| Out-String`).
- Call `/api/devices/deploy` directly with the short-lived deploy token.
- Generate a random static password for unattended access.
- Sync the device into the `My Devices` address book.
- Auto sign in the RustDesk account on the client (`access_token` in `RustDesk_local.toml`).
- Revoke the deploy token after successful setup.

The API first loads the deployment template from `data/templates/deploy-host.ps1`. In Docker this maps to `./data/api/templates/deploy-host.ps1` on the host, so the script template can be edited without rebuilding the Go binary. If that file does not exist, the API uses the default template at `resources/templates/deploy-host.ps1`, then falls back to the embedded template in code.

You can also run the wrapper manually if you already have a token:

```powershell
.\deploy-host.ps1 -DeployToken "<DEPLOY_TOKEN>" -ApiUrl "https://rustdesk.example.com"
```

## Manual Client Deployment

Windows:

```cmd
"C:\Program Files\RustDesk\rustdesk.exe" --deploy --token <USER_API_TOKEN>
"C:\Program Files\RustDesk\rustdesk.exe" --get-id | more
```

PowerShell (read ID — the Windows GUI build often does not print to the console directly):

```powershell
& "C:\Program Files\RustDesk\rustdesk.exe" --get-id | Out-String
```

Linux/macOS:

```bash
sudo rustdesk --deploy --token <USER_API_TOKEN>
rustdesk --get-id
```

After a successful deployment, the device appears in Web Admin/My Devices or is synced to the address book when using the automated script.

## Operations

Rebuild after code changes:

```bash
docker compose up -d --build
```

View logs:

```bash
docker compose logs -f hbbs
docker compose logs -f hbbr
docker compose logs -f rustdesk-api
docker compose logs -f reverse-proxy
```

Minimum backup set:

```text
data/server/   # hbbs/hbbr keypair
data/api/      # API database
.env           # domain, JWT secret, flags
```

Do not lose `data/server/id_ed25519` unless you are prepared to reconfigure all clients with a new server key.

## Updating Forks / Upstream

Before updating:

```bash
git status --short
git -C rustdesk-server status --short
git -C rustdesk-api status --short
git -C rustdesk-api-web status --short
```

Update each component:

```bash
git -C rustdesk-server pull --ff-only
git -C rustdesk-api pull --ff-only
git -C rustdesk-api-web pull --ff-only
docker compose up -d --build
```

Minimum checks after updating:

- Web Admin login works.
- Clients receive the correct `ID Server`, `Relay Server`, `API Server`, and `key`.
- Deploy tokens can be generated in `My -> Client Config`.
- The Windows deploy script reaches address book sync.
- `MUST_LOGIN=Y` prevents unknown devices from self-registering when required deployment is enabled.

## Local Tests

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

Full stack:

```bash
docker compose up -d --build
docker compose ps
```

## Security Notes

- Do not commit `.env`, `data/`, databases, private keys, or tokens.
- Use a long random `JWT_SECRET`, and do not reuse it across test and production.
- Generate deploy tokens only when needed because they can assign devices to accounts.
- Put TLS in front of the public reverse proxy before distributing deploy commands outside the internal network.
- After changing domain/API URL settings, generate a new deploy command so the script downloads from the correct endpoint.

## License And Responsibility

RustDesk components and related forks keep their respective upstream licenses. This integration repository documents packaging, configuration, and internal operational changes. Review each upstream/fork license before redistributing binaries or container images.
