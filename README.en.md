# 🌐 MultiVPN

![License](https://img.shields.io/badge/license-GPLv3-green)
![Status](https://img.shields.io/badge/status-Active_Development-orange)
![Stack](https://img.shields.io/badge/stack-Go_%7C_HTMX_%7C_Alpine.js_%7C_Tailwind-informational)
![Platforms](https://img.shields.io/badge/platforms-Linux_%7C_Windows_(WIP)_%7C_BSD_(WIP)-blue)

Intelligent VPN connection manager with **embedded local WebUI**, dual persistence with **SQLite**, and **automatic Default Gateway shielding** (Strict Split Tunneling).

Built with **Go** as a standalone single static binary without heavy external dependencies or Node.js toolchains.

---

## 📦 Installation & Download

### Option 1: Quick one-liner installer (Recommended for Linux)
Run the following in your terminal to download, verify, and install:
```bash
curl -fsSL https://raw.githubusercontent.com/terracenter/multivpn/main/install.sh | bash
```

---

### Option 2: Manual installation (Linux)
```bash
curl -fsSL -o multivpn https://github.com/terracenter/multivpn/releases/latest/download/multivpn
mkdir -p ~/.local/bin
install -m 755 multivpn ~/.local/bin/multivpn
rm -f multivpn
```

**On Windows:**
Download `multivpn.exe` from [Releases](https://github.com/terracenter/multivpn/releases) and run it via Terminal or PowerShell.

---

### Option 2: Build from source (Requires Go 1.22+)
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
