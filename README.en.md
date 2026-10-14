# 🌐 MultiVPN

![License](https://img.shields.io/badge/license-GPLv3-green)
![Status](https://img.shields.io/badge/status-Active_Development-orange)
![Stack](https://img.shields.io/badge/stack-Go_%7C_HTMX_%7C_Alpine.js_%7C_Tailwind-informational)
![Platforms](https://img.shields.io/badge/platforms-Linux_%7C_Windows_(WIP)_%7C_BSD_(WIP)-blue)

Intelligent VPN connection manager with **embedded local WebUI**, dual persistence with **SQLite**, and **automatic Default Gateway shielding** (Strict Split Tunneling).

Built with **Go** as a standalone single static binary without heavy external dependencies or Node.js toolchains.

---

## 📦 Installation & Download

### On Linux (Official Quick Install)
Run the following in your terminal to download, verify, and install the latest release:
```bash
curl -fsSL https://raw.githubusercontent.com/terracenter/multivpn/main/install.sh | bash
```
> The installer detects your system architecture, verifies binary integrity, resolves any previous path collisions, and installs cleanly to `~/.local/bin/multivpn` (or `/usr/local/bin/` if run with `sudo`).

---

### On Windows
Download `multivpn.exe` directly from [GitHub Releases](https://github.com/terracenter/multivpn/releases) and run it via Terminal or PowerShell.

---

### Build from source (For Developers - Go 1.22+)
```bash
git clone https://github.com/terracenter/multivpn.git
cd multivpn
make build
```

---

## 🚀 Quick Start

Once you have the `multivpn` binary:

```bash
./multivpn web
# Or simply:
./multivpn
```

The system will start on `http://127.0.0.1:8765` and automatically launch your default web browser.

---

## 🤝 Contributing

This project is **100% Free Software** under **GPLv3**.
Contributions, bug reports, and enhancements are welcome! See [`CONTRIBUTING.md`](CONTRIBUTING.md).
