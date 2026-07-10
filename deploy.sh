#!/usr/bin/env bash
# Deploys the laverna yomitan server + Traefik stack to a remote VM over SSH.
# No Docker registry involved: the source is tarred, copied over SSH, and built on the VM.
#
# Usage: ./deploy.sh user@host [remote-dir]
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 user@host [remote-dir]" >&2
  exit 1
fi

TARGET="$1"
REMOTE_DIR="${2:-/opt/laverna}"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "==> Ensuring Docker is installed on ${TARGET} (Debian)"
ssh "$TARGET" bash -s <<'REMOTE_SCRIPT'
set -euo pipefail
if command -v docker >/dev/null 2>&1; then
  exit 0
fi

apt update
apt install -y ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/debian
Suites: $(. /etc/os-release && echo "$VERSION_CODENAME")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF

apt update
apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
systemctl enable --now docker
REMOTE_SCRIPT

echo "==> Copying project to ${TARGET}:${REMOTE_DIR}"
ssh "$TARGET" "mkdir -p '$REMOTE_DIR'"
tar --exclude='.git' --exclude='main' --exclude='.claude' -C "$PROJECT_ROOT" -czf - . \
  | ssh "$TARGET" "tar -xzf - -C '$REMOTE_DIR'"

echo "==> Preparing Traefik prerequisites and starting the stack"
ssh "$TARGET" bash -s -- "$REMOTE_DIR" <<'REMOTE_SCRIPT'
set -euo pipefail
REMOTE_DIR="$1"
cd "$REMOTE_DIR"

docker network inspect traefik-network >/dev/null 2>&1 || docker network create traefik-network

if [[ ! -f acme.json ]]; then
  touch acme.json
  chmod 600 acme.json
fi

docker compose up -d --build
docker compose ps
REMOTE_SCRIPT

echo "==> Done"
