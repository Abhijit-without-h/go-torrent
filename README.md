# 🌀 GoTorrent

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go\&logoColor=white)](https://go.dev/)
[![Build Status](https://img.shields.io/github/actions/workflow/status/yourname/gotorrent/ci.yml?branch=main)](https://github.com/yourname/gotorrent/actions)
[![License](https://img.shields.io/github/license/yourname/gotorrent?color=brightgreen)](LICENSE)

A minimal, educational peer‑to‑peer (P2P) file‑sharing app written in Go.
Create `.torrent` metadata, seed files, and download chunks from peers – all with a single binary.

---

## ✨ Key Features

* **Torrent Creation** – Generate metadata files with SHA‑256 chunk hashes.
* **Seeding** – Serve file chunks to peers over TCP.
* **Downloading** – Fetch chunks in parallel and reconstruct the original file.
* **Integrity‑First** – Every chunk is verified before it hits disk.
* **Progress Tracking** – Simple, human‑readable stats for both seeders and downloaders.

---

## 🚀 Built With

| Framework / Library | Purpose                                            |
| ------------------- | -------------------------------------------------- |
| **Go stdlib**       | Core language + `net`, `os`, `crypto/sha256`, etc. |

> *Add‑ons & plugins such as Echo, Swagger, or DDNS clients are planned but are listed later in **Acknowledgements**.*

## 📑 Table of Contents

1. [Getting Started](#-getting-started)
2. [Usage](#-usage)
3. [Roadmap](#-roadmap)
4. [Contributing](#-contributing)
5. [License](#-license)
6. [Acknowledgements](#-acknowledgements)

---

## 🏗️ Getting Started

### Prerequisites

* Go **1.22 or later** installed (`go version`).

### Installation

```bash
git clone https://github.com/yourname/gotorrent.git   # replace with actual URL
cd gotorrent
go build -o gotorrent .
```

---

## 🔧 Usage

GoTorrent has three modes: **create**, **seed**, **download**.

### 1. Create a `.torrent` file

```bash
./gotorrent create <file_path>
# Example
./gotorrent create ./movies/Interstellar.mkv
```

The generated `Interstellar.mkv.torrent` will appear in `./torrents`.

### 2. Seed the file

```bash
./gotorrent seed <original_file_path>
# Example
./gotorrent seed ./movies/Interstellar.mkv
```

By default the seeder listens on `127.0.0.1:8080`.

### 3. Download using the torrent file

```bash
./gotorrent download ./torrents/Interstellar.mkv.torrent
```

The file is reconstructed in `./downloads`.

> **Tip:** Use different terminal tabs for seeder and downloader when testing locally.

---

## 🌱 Roadmap

| Stage    | Focus                                 | Status |
| -------- | ------------------------------------- | ------ |
| **v0.1** | Local P2P over TCP (current)          | ✅      |
| **v0.2** | Switch to HTTP chunk endpoints (Echo) | 🚧     |
| **v0.3** | Public tracker + cloud deployment     | 🕒     |
| **v1.0** | NAT traversal & full decentralisation | 🔭     |

See [FUTURE.md](FUTURE.md) for a detailed plan on scalability, load‑balancing, and object storage integration.

---

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

> **Good first issues** are tagged `help-wanted`. Feel free to ask questions!

---

## 📝 License

Distributed under the MIT License. See **LICENSE** for more information.

---

## 💐 Acknowledgements

* [Echo](https://echo.labstack.com/) – planned HTTP framework
* [Swag](https://github.com/swaggo/swag) – API docs generator
* [Railway](https://railway.app/) – free cloud deployment platform
* [DDNS‑Go](https://github.com/jeessy2/ddns-go) – dynamic DNS client
* Inspired by the original [BitTorrent protocol](https://www.bittorrent.org/).

---

Made with ❤️ and Go.
