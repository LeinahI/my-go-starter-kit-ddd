# `/scripts`

Shell scripts for **production operations** on a VPS or bare-metal host — pull, build, migrate, restart. They keep the root `Makefile` focused on **local development**.

This follows the same pattern used by large Go and infra projects:

- [HashiCorp Terraform `scripts/`](https://github.com/hashicorp/terraform/tree/main/scripts)
- [Kubernetes Helm `scripts/`](https://github.com/kubernetes/helm/tree/master/scripts)
- [CockroachDB `scripts/`](https://github.com/cockroachdb/cockroach/tree/master/scripts)

[golang-standards/project-layout](https://github.com/golang-standards/project-layout#scripts) describes `/scripts` as the place for build, install, and analysis operations that would clutter the Makefile.

## Makefile vs scripts vs deployments

| Tool | Audience | Examples |
|---|---|---|
| **`Makefile`** | Developers on laptop | `make run-api`, `make test`, `make migrate-up` |
| **`scripts/`** | Production server (SSH) | `./scripts/build.sh`, `./scripts/deploy.sh` |
| **`deployments/`** | Container / compose definitions | `Dockerfile`, `docker-compose.yml` |

**Rule:** If you run it over SSH on a VPS after merging to `main`, it belongs in `scripts/`. If you run it while coding locally, use `make`.

## Scripts in this starter

| Script | Purpose |
|---|---|
| [`build.sh`](build.sh) | `git pull` from `main` (optional) + compile `bin/api` and `bin/migrate` |
| [`deploy.sh`](deploy.sh) | Full deploy: `build.sh` → goose migrate → restart systemd service |
| [`lib/common.sh`](lib/common.sh) | Shared logging and env validation |

Make executable once on the server:

```bash
chmod +x scripts/*.sh
```

## VPS workflow (your main use case)

After pushing to `main`:

```bash
ssh deploy@your-vps
cd /opt/ws          # clone path on the server
./scripts/deploy.sh
```

Or build only (no migrate/restart):

```bash
./scripts/build.sh
```

`build.sh` alone is enough when you only need a fresh binary; `deploy.sh` is the full release path.

## First-time server setup

1. **Install Go** (same major version as `go.mod`) and **Git** on the VPS.

2. **Clone the repo** (once):

```bash
sudo mkdir -p /opt/ws
sudo chown "$USER" /opt/ws
git clone git@github.com:yourorg/ws.git /opt/ws
cd /opt/ws
chmod +x scripts/*.sh
```

3. **Create production env** (never commit this):

```bash
sudo mkdir -p /etc/ws
sudo nano /etc/ws/ws.env
```

```bash
APP_ENV=production
HTTP_HOST=0.0.0.0
HTTP_PORT=8080
DATABASE_URL=postgres://user:pass@127.0.0.1:5432/ws?sslmode=disable
CORS_ALLOWED_ORIGINS=https://admin.example.com,https://www.example.com
```

```bash
sudo chmod 600 /etc/ws/ws.env
```

4. **Run migrations** on first deploy:

```bash
export ENV_FILE=/etc/ws/ws.env
source /etc/ws/ws.env
./scripts/build.sh
./bin/migrate up
```

5. **systemd service** (example — adjust paths):

```ini
# /etc/systemd/system/ws-api.service
[Unit]
Description=ws API
After=network.target postgresql.service

[Service]
Type=simple
User=ws
WorkingDirectory=/opt/ws
EnvironmentFile=/etc/ws/ws.env
ExecStart=/opt/ws/bin/api
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable ws-api
export SYSTEMD_SERVICE=ws-api   # or add to /etc/ws/ws.env
```

6. **Deploy** on every release:

```bash
./scripts/deploy.sh
```

## Environment variables

| Variable | Default | Used by |
|---|---|---|
| `GIT_BRANCH` | `main` | `build.sh` |
| `SKIP_GIT_PULL` | `0` | `build.sh` — set `1` to skip pull (CI compile) |
| `BIN_DIR` | `bin` | `build.sh`, `deploy.sh` |
| `GOOS` / `GOARCH` | `linux` / `amd64` | `build.sh` |
| `ENV_FILE` | `/etc/ws/ws.env` | `deploy.sh` |
| `DATABASE_URL` | — | `deploy.sh` (required for migrations) |
| `SYSTEMD_SERVICE` | — | `deploy.sh` — e.g. `ws-api` |

## CI vs VPS

| Context | Command |
|---|---|
| GitHub Actions build artifact | `SKIP_GIT_PULL=1 ./scripts/build.sh` |
| VPS after merge to main | `./scripts/deploy.sh` |
| Local dev | `make build` / `make run-api` |

## What not to put here

- Application business logic (belongs in `internal/`)
- Docker compose files (belongs in `deployments/`)
- One-off dev helpers (use `make` targets instead)
- Secrets (use `/etc/ws/ws.env` or your secret manager — never commit)

## Laravel mapping

| Laravel / Forge | This starter |
|---|---|
| Forge “Deploy Now” | `./scripts/deploy.sh` |
| `git pull` on server | `build.sh` (built in) |
| `php artisan migrate --force` | `bin/migrate up` in `deploy.sh` |
| `php artisan config:cache` | env file at `/etc/ws/ws.env` |
| Supervisor / systemd restart | `SYSTEMD_SERVICE=ws-api` |

## Trade-off (stated honestly)

**Scripts on the server are simple but not zero-downtime.** For blue/green or rolling deploys, move to Docker + orchestrator (use `deployments/Dockerfile`) or a proper CD pipeline. This starter optimizes for **small teams on a single VPS** — the common first production step before Kubernetes.
