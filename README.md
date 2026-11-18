# ⚡ DownForge

**DownForge** is a modern, modular, self-hosted download manager built with **Go** and **React**. Designed to run smoothly as a background system service on **Linux** (via `systemctl`) and **Windows**, it provides a responsive browser-based management UI at `http://localhost:8080`.

It seamlessly manages both **HTTP/HTTPS** and **BitTorrent** downloads with real-time speed, progress, and status updates delivered through WebSockets.

---

## 🌟 Key Features

- **Dual-Engine Downloader**:
  - **HTTP/HTTPS**: Supports chunked downloads, pause/resume via HTTP `Range` headers, auto-filename extraction from `Content-Disposition`, and atomic `.part` file handling.
  - **BitTorrent**: Powered by [`github.com/anacrolix/torrent`](https://github.com/anacrolix/torrent). Supports Magnet URIs, `.torrent` file uploads, DHT, PEX, peer tracking, piece verification, and multi-file tracking.
- **Real-Time Web UI**:
  - React (Vite + Tailwind CSS + Lucide Icons) with dark mode.
  - Live progress bars, ETA calculations, peer metrics, and instantaneous download/upload speed gauges.
  - Real-time updates pushed over **WebSockets** (`/ws`) with automatic reconnection.
- **Robust SQLite Database**:
  - Uses pure Go SQLite (`modernc.org/sqlite`) running in WAL mode for concurrent, crash-safe persistence.
- **Service-Ready**:
  - Runs as a headless daemon controlled through your browser.
  - Native Linux `systemd` unit file and built-in service installer (`./downforge -service install`).
  - Windows background service setup scripts.
- **Modular Architecture**:
  - Strictly separated domain logic, clean interfaces, and zero spaghetti code.

---

## 📁 Modular Directory Structure

```
DownForge/
├── cmd/
│   └── downforge/
│       └── main.go              # Application entrypoint, CLI flags, graceful shutdown
├── internal/
│   ├── api/                     # HTTP REST & WebSocket transport layer
│   │   ├── router.go            # Chi router, CORS, SPA static file serving
│   │   ├── handlers.go          # REST endpoints (add, pause, resume, delete, settings)
│   │   ├── websocket.go         # Gorilla WebSocket hub, client pumping, broadcasting
│   │   └── api_test.go          # REST API endpoint tests
│   ├── config/                  # Configuration management
│   │   └── config.go            # Config structs, paths, default settings loader
│   ├── db/                      # SQLite persistence layer
│   │   ├── db.go                # SQLite connection, WAL pragmas, schema migrations
│   │   ├── models.go            # Database models and DTO structs
│   │   ├── repository.go        # CRUD operations for downloads & settings
│   │   └── repository_test.go   # Database integration unit tests
│   ├── downloader/              # Unified download coordination
│   │   ├── manager.go           # Manager orchestrating HTTP & Torrent engines & ticker
│   │   ├── item.go              # Unified download snapshot, ETA and progress math
│   │   ├── http/                # HTTP/HTTPS downloader
│   │   │   └── client.go        # Resumable range requests, part files, speed tracker
│   │   └── torrent/             # BitTorrent downloader
│   │       └── client.go        # anacrolix/torrent engine, magnet/file handler
│   └── service/                 # System service integration
│       └── service.go           # Linux systemd unit file generator & installer
├── web/                         # React Frontend (Vite + Tailwind CSS)
│   ├── src/
│   │   ├── components/          # Navbar, DownloadList, DownloadItem, Modals
│   │   ├── hooks/               # useWebSocket hook for real-time reactivity
│   │   ├── services/            # REST API client & byte formatting helpers
│   │   ├── App.jsx              # Main Dashboard
│   │   └── main.jsx             # React DOM entrypoint
│   ├── package.json
│   ├── vite.config.js
│   └── tailwind.config.js
├── scripts/
│   ├── downforge.service        # Systemd unit file template
│   ├── install-service.sh       # Linux systemd installer script
│   └── install-service-windows.ps1 # Windows service installer script
├── Makefile                     # Build & run automation
├── go.mod
├── go.sum
└── README.md
```

---

## 🚀 Quick Start

### Prerequisites
- **Go**: 1.22+ installed
- **Node.js**: 18+ and `npm` installed

### 1. Build & Run Everything with Make
```bash
# Build React frontend and Go binary, then start the server
make run
```
Once started, open your browser and navigate to:
👉 **[http://localhost:8080](http://localhost:8080)**

---

## 🛠️ Development Mode

To run backend and frontend with Vite hot-reloading:

### Terminal 1: Backend
```bash
go run ./cmd/downforge -port 8080
```

### Terminal 2: React UI (Vite Dev Server)
```bash
cd web
npm install
npm run dev
```
The Vite development server will start at `http://localhost:5173` with proxy routing to `http://localhost:8080` for `/api` and `/ws`.

---

## 🐧 Running as a Linux Service (`systemctl`)

DownForge can run in the background as a `systemd` service and start automatically on boot.

### Method 1: Using DownForge CLI (Recommended)
```bash
# Compile the binary
go build -o downforge ./cmd/downforge

# Install and start systemd service
sudo ./downforge -service install
```

### Method 2: Using the Installation Script
```bash
sudo ./scripts/install-service.sh
```

### Managing the Service
```bash
# Check status
sudo systemctl status downforge.service

# Start / Stop / Restart
sudo systemctl start downforge.service
sudo systemctl stop downforge.service
sudo systemctl restart downforge.service

# View live background service logs
journalctl -u downforge.service -f
```

---

## 🪟 Running as a Windows Service

1. Open PowerShell as **Administrator**.
2. Run the provided installation script:
```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install-service-windows.ps1
```
This automatically compiles `downforge.exe`, registers it with Windows Service Manager (using `nssm` or `sc.exe`), and starts it.

---

## 🔌 REST API Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/downloads` | List all downloads with live progress snapshots |
| `GET` | `/api/downloads/{id}` | Get detail snapshot for a single download |
| `POST` | `/api/downloads` | Add new download (JSON or Multipart Form) |
| `POST` | `/api/downloads/{id}/pause` | Pause an active download |
| `POST` | `/api/downloads/{id}/resume` | Resume a paused download |
| `DELETE` | `/api/downloads/{id}` | Delete download (`?deleteFiles=true` to wipe disk) |
| `GET` | `/api/stats` | Global network speeds & download counts |
| `GET` | `/api/settings` | Read application settings |
| `PUT` | `/api/settings` | Update application settings |
| `POST` | `/api/open-folder` | Open local folder in OS file explorer |

### Example: Adding a Download via cURL

**HTTP/HTTPS URL:**
```bash
curl -X POST http://localhost:8080/api/downloads \
  -H "Content-Type: application/json" \
  -d '{"type": "http", "url": "https://example.com/file.zip"}'
```

**BitTorrent Magnet Link:**
```bash
curl -X POST http://localhost:8080/api/downloads \
  -H "Content-Type: application/json" \
  -d '{"type": "torrent", "url": "magnet:?xt=urn:btih:..."}'
```

**Uploading a `.torrent` File:**
```bash
curl -X POST http://localhost:8080/api/downloads \
  -F "torrent=@/path/to/download.torrent"
```

---

## 📡 WebSocket Real-Time API (`/ws`)

Connect to `ws://localhost:8080/ws` from any client. Messages are JSON objects:
```json
{
  "event": "progress",
  "payload": [
    {
      "id": "c1f7b889-40b4-4b57-a9a3-5c210ab2851a",
      "name": "ubuntu-24.04-desktop-amd64.iso",
      "type": "torrent",
      "source": "magnet:?xt=...",
      "save_path": "/home/user/Downloads/DownForge/ubuntu.iso",
      "total_bytes": 6053871616,
      "downloaded_bytes": 1204593664,
      "progress": 19.9,
      "download_speed": 12450000,
      "upload_speed": 450000,
      "eta": 389,
      "status": "downloading",
      "peers": 42
    }
  ]
}
```

Events broadcasted:
- `"progress"`: Array of active & completed download snapshots (every 1s)
- `"stats"`: Global counts & aggregate download/upload speeds (every 1s)
- `"added"`: Emitted immediately when a new task is created
- `"completed"`: Emitted immediately when a task finishes
- `"paused"`, `"resumed"`, `"deleted"`: Emitted on user actions

---

## 🧪 Testing

Run backend unit & integration tests:
```bash
make test
```

---

## 📄 License

MIT License. Designed and crafted for high-performance self-hosted downloading.
