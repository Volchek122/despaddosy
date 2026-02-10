# Flixy - TV Hub Client

A Netflix-like client for `hdrezka.ag` (or similar providers) built with Python and PyQt6.

## Features
- **Beautiful Design:** Dark theme, poster grid, seamless navigation.
- **TV Optimized:** Full keyboard/remote control support.
- **Provider:** Uses `HdRezkaApi` for content (requires valid installation/proxy) or falls back to a Mock provider for demo.
- **Player:** Integrated video player.

## Installation

1. Install system dependencies (Ubuntu/Debian):
   ```bash
   sudo apt install python3-pyqt6 python3-requests python3-bs4 python3-lxml vlc
   ```
   *Note: VLC is recommended if using external player, but the app has an internal player.*

2. Install Python dependencies:
   ```bash
   pip install -r requirements.txt
   ```

## Running

Run the start script:
```bash
./start.sh
```

## Configuration

The application tries to use `HdRezkaApi` if installed. If it fails (e.g. due to Cloudflare blocks or missing library), it automatically falls back to a `MockProvider` with demo content.

To use the real API, ensure you have a working `HdRezkaApi` library installed and network access to `hdrezka.ag`.
