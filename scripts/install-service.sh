#!/usr/bin/env bash
#
# DownForge Service Installer for Linux (systemd)
#
set -e

if [ "$EUID" -ne 0 ]; then
  echo "❌ Please run this script with sudo or as root: sudo ./scripts/install-service.sh"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARY_SRC="$PROJECT_ROOT/downforge"
TARGET_BIN="/usr/local/bin/downforge"
SERVICE_SRC="$SCRIPT_DIR/downforge.service"
SERVICE_DEST="/etc/systemd/system/downforge.service"
ACTUAL_USER="${SUDO_USER:-$USER}"

echo "🔨 Building DownForge..."
cd "$PROJECT_ROOT"
go build -o "$BINARY_SRC" ./cmd/downforge

echo "📦 Installing binary to $TARGET_BIN..."
cp -f "$BINARY_SRC" "$TARGET_BIN"
chmod +x "$TARGET_BIN"

echo "⚙️ Configuring systemd unit for user: $ACTUAL_USER..."
sed -e "s/%i/$ACTUAL_USER/g" "$SERVICE_SRC" > "$SERVICE_DEST"
chmod 644 "$SERVICE_DEST"

echo "🔄 Reloading systemd daemon..."
systemctl daemon-reload

echo "🚀 Enabling and starting downforge.service..."
systemctl enable downforge.service
systemctl restart downforge.service

echo "✅ DownForge is now installed and running as a systemd service!"
echo "   Status: sudo systemctl status downforge.service"
echo "   Web UI: http://localhost:8080"
