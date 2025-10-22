# Runtime Dependencies for Sliver-Plus Features

**Date:** October 22, 2025  
**Last Updated:** October 22, 2025

---

## Overview

Sliver-Plus augments Sliver with additional reconnaissance capabilities. These features require external tools to be installed on the **operator's machine** (where the Sliver server runs).

**Important:** These are **runtime dependencies only** - they are NOT required to build/compile Sliver.

---

## Quick Install Guide

### All Tools (One-Liner)

**Linux/macOS:**
```bash
# Install Go tools
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest

# Install Python tools (requires Python 3.8+)
pip3 install theHarvester

# Install system tools
sudo apt-get update && sudo apt-get install -y nmap
# or on macOS: brew install nmap
```

**Windows:**
```powershell
# Install Go tools
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest

# Install Python tools (requires Python 3.8+)
pip3 install theHarvester

# Install nmap
choco install nmap
# or download from https://nmap.org/download.html
```

---

## Feature: hosts scan (Active Scanning)

### Required: nmap

**Purpose:** Network port scanning and service detection

**Installation:**

**Linux (Debian/Ubuntu):**
```bash
sudo apt-get update
sudo apt-get install -y nmap
```

**Linux (RHEL/CentOS):**
```bash
sudo yum install -y nmap
```

**macOS:**
```bash
brew install nmap
```

**Windows:**
- Download from: https://nmap.org/download.html
- Or via Chocolatey: `choco install nmap`

**Verification:**
```bash
nmap --version
# Should show: Nmap version 7.80 or higher
```

**Usage in Sliver:**
```bash
sliver> hosts scan 192.168.1.0/24 -s -o
```

**Graceful Failure:**
- If not installed: Clear error message
- Command fails with: "nmap not found in PATH - please install nmap"
- No crashes or undefined behavior

---

## Feature: loot recon (Passive OSINT)

### Required: subfinder

**Purpose:** Fast subdomain enumeration using passive sources

**Installation:**

```bash
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
```

**Post-Install:**
```bash
# Add to PATH (if not already)
export PATH=$PATH:~/go/bin

# Verify
subfinder -version
```

**Verification:**
```bash
subfinder -version
# Should show: subfinder version or usage info
```

**Usage in Sliver:**
```bash
sliver> loot recon domains example.com
```

**Graceful Failure:**
- If not installed: "subfinder not found in PATH - install: go install ..."
- Command fails cleanly

---

### Optional: theHarvester

**Purpose:** Email and additional OSINT data gathering

**Installation:**

**Via pip (Recommended):**
```bash
pip3 install theHarvester
```

**From source:**
```bash
git clone https://github.com/laramies/theHarvester.git
cd theHarvester
pip3 install -r requirements.txt
python3 setup.py install
```

**Verification:**
```bash
theHarvester --version
# or
theHarvester -h
```

**Usage in Sliver:**
```bash
sliver> loot recon emails example.com
sliver> loot recon all example.com
```

**Graceful Failure:**
- If not installed: "theHarvester not found in PATH - install: pip3 install theHarvester"
- Feature degrades gracefully (subdomains still work via subfinder)

---

## Dependency Matrix

| Feature | Command | Required Tools | Optional Tools | Failsafe |
|---------|---------|----------------|----------------|----------|
| Active Scanning | `hosts scan` | nmap | - | ✅ Yes |
| Subdomain Enum | `loot recon domains` | subfinder | theHarvester | ✅ Yes |
| Email Enum | `loot recon emails` | theHarvester | - | ✅ Yes |
| Full OSINT | `loot recon all` | subfinder | theHarvester | ✅ Partial |

**Failsafe:** All features check for tool availability before execution and provide clear installation instructions if missing.

---

## Tool Capabilities

### nmap
- **Speed:** Slow (minutes for large ranges)
- **Stealth:** Noisy (active scanning)
- **Accuracy:** Very high
- **Output:** Detailed service information
- **Use Case:** Internal network mapping, service enumeration

### subfinder
- **Speed:** Fast (seconds)
- **Stealth:** Silent (passive queries only)
- **Accuracy:** High
- **Output:** Subdomains with sources
- **Use Case:** Pre-engagement reconnaissance, domain mapping

### theHarvester
- **Speed:** Medium (30-60 seconds)
- **Stealth:** Silent (passive OSINT)
- **Accuracy:** Medium (depends on sources)
- **Output:** Emails, subdomains, IPs
- **Use Case:** Email harvesting, additional OSINT

---

## Environment Setup

### Recommended Setup Script

**Linux/macOS:** Create `setup-sliver-plus.sh`

```bash
#!/bin/bash
set -e

echo "[*] Installing Sliver-Plus runtime dependencies..."

# Check Go
if ! command -v go &> /dev/null; then
    echo "[!] Go is required but not installed. Please install Go first."
    exit 1
fi

# Install subfinder
echo "[*] Installing subfinder..."
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest

# Install theHarvester
echo "[*] Installing theHarvester..."
if command -v pip3 &> /dev/null; then
    pip3 install theHarvester
else
    echo "[!] pip3 not found, skipping theHarvester"
fi

# Install nmap (platform-specific)
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "[*] Installing nmap (requires sudo)..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y nmap
    elif command -v yum &> /dev/null; then
        sudo yum install -y nmap
    else
        echo "[!] Unknown package manager, please install nmap manually"
    fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "[*] Installing nmap via Homebrew..."
    if command -v brew &> /dev/null; then
        brew install nmap
    else
        echo "[!] Homebrew not found, please install nmap manually"
    fi
fi

echo ""
echo "[+] Installation complete!"
echo ""
echo "Verify installations:"
echo "  nmap --version"
echo "  subfinder -version"
echo "  theHarvester --version"
```

**Usage:**
```bash
chmod +x setup-sliver-plus.sh
./setup-sliver-plus.sh
```

---

## Troubleshooting

### Issue: "command not found"

**Symptom:**
```
sliver> hosts scan 192.168.1.1
[!] Error: nmap not found in PATH
```

**Solution:**
1. Install the tool (see sections above)
2. Verify it's in PATH: `which nmap` (Linux/macOS) or `where nmap` (Windows)
3. If installed but not in PATH, add to PATH:
   ```bash
   export PATH=$PATH:/path/to/tool
   ```

### Issue: Permission Denied (nmap)

**Symptom:**
```
[!] nmap failed: permission denied
```

**Solution:**
- Some nmap scans require root/admin privileges
- Use `sudo` when starting Sliver server: `sudo ./sliver-server`
- Or grant capabilities: `sudo setcap cap_net_raw,cap_net_admin=eip /usr/bin/nmap`

### Issue: subfinder not finding subdomains

**Symptom:**
```
[*] Reconnaissance complete
[*] Found 0 subdomains
```

**Solutions:**
1. Check internet connectivity
2. Configure API keys for better results:
   ```bash
   # Create config at ~/.config/subfinder/config.yaml
   shodan: [your-api-key]
   censys: [your-api-key]
   virustotal: [your-api-key]
   ```
3. Try with a known domain: `sliver> loot recon domains google.com`

### Issue: theHarvester timeout

**Symptom:**
```
[!] theHarvester timed out after 300 seconds
```

**Solutions:**
1. Increase timeout: Add `--timeout 600` flag (if supported)
2. Use fewer sources: theHarvester `-b` option
3. Check network connectivity

---

## Security Considerations

### Legal

**⚠️ WARNING:** These tools perform reconnaissance activities that may be:
- **Illegal** without proper authorization
- **Against Terms of Service** of target systems
- **Logged and monitored** by security teams

**Only use on:**
- Systems you own
- Systems you have explicit written permission to test
- Within the scope of authorized penetration tests

### Operational Security

**Active Scanning (nmap):**
- ❌ **Noisy** - Generates logs, triggers IDS/IPS
- ❌ **Attribution** - Scans originate from your IP
- ✅ **Use:** VPN/proxy, authorized networks only

**Passive OSINT (subfinder, theHarvester):**
- ✅ **Quiet** - No direct target contact
- ⚠️ **Some attribution** - API keys may be logged
- ✅ **Generally safer** for pre-engagement recon

### Recommendations

1. **Use in lab environments first**
2. **Understand what each tool does before running**
3. **Check authorization before any scanning**
4. **Use VPN/proxy for attribution control**
5. **Review and follow your organization's security policies**

---

## Version Requirements

| Tool | Minimum Version | Recommended | Notes |
|------|----------------|-------------|-------|
| nmap | 7.80 | Latest | Older versions may lack features |
| subfinder | 2.5.0 | Latest | API integrations improve over time |
| theHarvester | 4.0.0 | Latest | Earlier versions have different output formats |
| Go | 1.21 | 1.25+ | Required for subfinder installation |
| Python | 3.8 | 3.11+ | Required for theHarvester |

---

## API Keys (Optional but Recommended)

### subfinder API Keys

**Improves subdomain discovery significantly**

**Configuration:** `~/.config/subfinder/config.yaml`

```yaml
shodan: [your-shodan-api-key]
censys:
  - [censys-api-id]
  - [censys-api-secret]
virustotal: [your-virustotal-api-key]
passivetotal:
  username: [your-username]
  key: [your-api-key]
securitytrails: [your-securitytrails-api-key]
```

**Free API Keys:**
- **Shodan:** https://account.shodan.io/
- **VirusTotal:** https://www.virustotal.com/gui/join-us
- **SecurityTrails:** https://securitytrails.com/app/signup

---

## Build-Time vs Runtime

### ✅ Build-Time (NOT Required)

**These tools are NOT needed to compile Sliver:**
- Sliver compiles with standard Go toolchain
- External tool references are just strings
- `exec.Command("nmap", ...)` compiles without nmap installed

### ⚠️ Runtime (Required for Features)

**These tools ARE needed to USE the features:**
- Checked at runtime via `exec.LookPath()`
- Clear error messages if missing
- Graceful feature degradation

---

## Summary

### Quick Reference

**For Active Scanning:**
```bash
sudo apt-get install nmap  # or brew install nmap
```

**For Passive OSINT:**
```bash
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
pip3 install theHarvester
```

**Verify All:**
```bash
nmap --version && subfinder -version && theHarvester --version
```

### Feature Support

| If You Have... | You Can Use... |
|----------------|----------------|
| Nothing | ❌ No recon features |
| nmap only | ✅ `hosts scan` |
| subfinder only | ✅ `loot recon domains` |
| theHarvester only | ✅ `loot recon emails` |
| nmap + subfinder | ✅ Active + passive recon |
| All tools | ✅ Full Sliver-Plus capabilities |

---

## Support

**Tool Issues:**
- nmap: https://github.com/nmap/nmap/issues
- subfinder: https://github.com/projectdiscovery/subfinder/issues
- theHarvester: https://github.com/laramies/theHarvester/issues

**Sliver-Plus Integration Issues:**
- Check tool is in PATH
- Verify tool works standalone first
- Check Sliver server logs

---

**Last Updated:** October 22, 2025  
**Maintained by:** Sliver-Plus Development Team


