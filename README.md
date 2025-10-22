# Sliver-Plus

> **This is a fork of [Sliver by Bishop Fox](https://github.com/BishopFox/sliver) with enhanced reconnaissance capabilities.**
>
> All credit for the core Sliver framework goes to Bishop Fox and the Sliver contributors. This fork adds additional features for adversary simulation and red team operations.

## About Sliver

Sliver is an open source cross-platform adversary emulation/red team framework, it can be used by organizations of all sizes to perform security testing. Sliver's implants support C2 over Mutual TLS (mTLS), WireGuard, HTTP(S), and DNS and are dynamically compiled with per-binary asymmetric encryption keys.

The server and client support MacOS, Windows, and Linux. Implants are supported on MacOS, Windows, and Linux (and possibly every Golang compiler target but we've not tested them all).

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

## Sliver-Plus Enhancements

This fork extends Sliver with additional reconnaissance and intelligence gathering capabilities:

### 🔍 Active Network Scanning
- **`hosts scan`** - Nmap-powered network reconnaissance
- Service version detection
- OS fingerprinting
- CIDR range scanning
- Auto-populates hosts database

### 🕵️ Passive OSINT Reconnaissance
- **`loot recon`** - Passive intelligence gathering
- Subdomain enumeration (subfinder)
- Email harvesting (theHarvester)
- Multiple reconnaissance modes
- Auto-stores results in loot system

### 🧬 Build Diversity CLI
- **`--diversity`** - Enable metamorphic code generation
- **`--novel-techniques`** - Vector Space Metamorphism
- **`--seed`** - Reproducible builds
- **`--min-novelty`** - Novelty threshold control

See [CHANGELOG.md](CHANGELOG.md) for detailed feature documentation and [RUNTIME_DEPENDENCIES.md](docs-sliver-plus/RUNTIME_DEPENDENCIES.md) for tool setup.

---

## 🚀 Installation

### Prerequisites

- **Go 1.22+** (for building from source)
- **Make** (for building)
- **Git** (for cloning)

### Quick Install from Source

```bash
# Clone the repository
git clone https://github.com/rxid09672/sliver-plus.git
cd sliver-plus

# Build server and client
make

# The binaries will be in the root directory:
# ./sliver-server (server)
# ./sliver-client (client)
```

### Linux/macOS One-Liner Install

```bash
git clone https://github.com/rxid09672/sliver-plus.git && cd sliver-plus && make
```

### Runtime Dependencies (Optional)

Sliver-Plus works out of the box, but to use the enhanced reconnaissance features, install these tools:

#### **For Active Network Scanning (`hosts scan`)**
```bash
# Linux (Debian/Ubuntu)
sudo apt install nmap

# macOS
brew install nmap

# Windows
choco install nmap
```

#### **For Passive OSINT (`loot recon`)**
```bash
# Subfinder (subdomain enumeration)
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest

# theHarvester (email/OSINT gathering)
pip3 install theHarvester
```

**Note:** These tools are only needed for running the respective commands. Sliver-Plus will display clear installation instructions if a tool is missing.

See [RUNTIME_DEPENDENCIES.md](docs-sliver-plus/RUNTIME_DEPENDENCIES.md) for detailed installation guides including troubleshooting.

---

## 📚 Quick Start

### 1. Start the Server

```bash
# First time setup (generates configs)
./sliver-server

# Or run as daemon
./sliver-server daemon
```

### 2. Connect with Client

```bash
./sliver-client
```

### 3. Generate an Implant

```bash
# Basic implant
sliver> generate --os windows --arch amd64 --mtls your-server-ip:8888

# With metamorphic diversity (unique signatures)
sliver> generate --os windows --mtls your-server-ip:8888 --diversity

# With Vector Space Metamorphism (maximum evasion)
sliver> generate --os windows --mtls your-server-ip:8888 --diversity --novel-techniques
```

### 4. Use Enhanced Recon Features

```bash
# Active network scanning
sliver> hosts scan 192.168.1.0/24 -s -o

# Passive OSINT
sliver> loot recon domains example.com
sliver> loot recon emails example.com
```

---

## 📖 Documentation

- **Sliver-Plus Features:** [CHANGELOG.md](CHANGELOG.md)
- **Runtime Dependencies:** [docs-sliver-plus/RUNTIME_DEPENDENCIES.md](docs-sliver-plus/RUNTIME_DEPENDENCIES.md)
- **Original Sliver Docs:** [sliver.sh/docs](https://sliver.sh/docs)

---

## Original Sliver Information

# v1.6.0 / `master`

**NOTE:** You are looking at the latest master branch of Sliver v1.6.0; new PRs should target this branch. However, this branch is NOT RECOMMENDED for production use yet. Please use release tagged versions for the best experience.

For PRs containing bug fixes specific to Sliver v1.5, please target the [`v1.5.x/master` branch](https://github.com/BishopFox/sliver/tree/v1.5.x/master).

### Features

- Dynamic code generation
- Compile-time obfuscation
- Multiplayer-mode
- Staged and Stageless payloads
- [Procedurally generated C2](https://sliver.sh/docs?name=HTTPS+C2) over HTTP(S)
- [DNS canary](https://sliver.sh/docs?name=DNS+C2) blue team detection
- [Secure C2](https://sliver.sh/docs?name=Transport+Encryption) over mTLS, WireGuard, HTTP(S), and DNS
- Fully scriptable using [JavaScript/TypeScript](https://github.com/moloch--/sliver-script) or [Python](https://github.com/moloch--/sliver-py)
- Windows process migration, process injection, user token manipulation, etc.
- Let's Encrypt integration
- In-memory .NET assembly execution
- COFF/BOF in-memory loader
- TCP and named pipe pivots
- Much more!

### Getting Started

Download the latest [release](https://github.com/BishopFox/sliver/releases) and see the Sliver [wiki](https://sliver.sh/docs?name=Getting+Started) for a quick tutorial on basic setup and usage. To get the very latest and greatest compile from source.

#### Linux One Liner

`curl https://sliver.sh/install|sudo bash` and then run `sliver`

### Help!

Please checkout the [wiki](https://sliver.sh/), or start a [GitHub discussion](https://github.com/BishopFox/sliver/discussions). We also tend to hang out in the #golang Slack channel on the [Bloodhound Gang](https://bloodhoundgang.herokuapp.com/) server.

### Compile From Source

See the [wiki](https://sliver.sh/docs?name=Compile+from+Source).

### Feedback

Please take a moment and fill out [our survey](https://forms.gle/SwVsHFNh24ChG58C6).

### License - GPLv3

Sliver is licensed under [GPLv3](https://www.gnu.org/licenses/gpl-3.0.en.html), some sub-components may have separate licenses. See their respective subdirectories in this project for details.
