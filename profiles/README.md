# Sliver Malleable C2 Profiles

**Version:** 1.0  
**Status:** 🔧 **In Development**  
**Purpose:** Network signature evasion through traffic pattern customization

---

## 📖 Overview

Malleable C2 profiles allow you to customize how Sliver implants communicate with the C2 server, enabling them to blend in with legitimate traffic patterns.

**Inspired by:** Cobalt Strike Malleable C2 profiles  
**Adapted for:** Sliver's architecture and capabilities

---

## 🎯 Goals

1. **Evade Network Detection:** Make C2 traffic look like legitimate services
2. **Flexible Configuration:** Easy to create and customize profiles
3. **Backward Compatible:** Profiles are optional, defaults still work
4. **Reusable:** Share profiles across teams and operations

---

## 📁 Directory Structure

```
profiles/
├── README.md              # This file
├── schema.json            # JSON schema for validation
├── examples/
│   ├── minimal.yml        # Minimal working profile
│   ├── full.yml           # All features demonstrated
│   └── template.yml       # Template for custom profiles
├── services/
│   ├── amazon.yml         # AWS SDK traffic
│   ├── microsoft.yml      # Microsoft Graph API
│   ├── google.yml         # Google Cloud SDK
│   ├── cloudflare.yml     # Cloudflare API
│   ├── github.yml         # GitHub API
│   └── slack.yml          # Slack API
├── apt/
│   ├── apt1.yml           # Chinese APT patterns
│   ├── apt28.yml          # Russian APT patterns
│   └── apt29.yml          # Cozy Bear patterns
└── stealth/
    ├── low-and-slow.yml   # Minimal traffic profile
    └── high-frequency.yml # Aggressive polling
```

---

## 🔧 Profile Format

Profiles are written in **YAML** for ease of use and readability.

### **Basic Structure:**

```yaml
profile:
  name: "Profile Name"
  author: "Your Name"
  description: "What this profile mimics"
  version: "1.0"
  
  http:
    user_agents:
      - "User-Agent string 1"
      - "User-Agent string 2"
    
    uris:
      get: ["/path1", "/path2"]
      post: ["/api/endpoint"]
    
    headers:
      common:
        Accept: "*/*"
        Accept-Encoding: "gzip, deflate"
  
  tls:
    fingerprint: "chrome"  # chrome, firefox, ios, etc.
  
  timing:
    interval: 60  # seconds
    jitter: 30    # percent
```

---

## 📚 Profile Features

### **1. HTTP Customization**

#### **User-Agent Rotation**
- Define multiple User-Agent strings
- Random selection per request
- Mimics real browser/SDK behavior

#### **URI Patterns**
- Separate patterns for GET/POST
- Random selection from list
- Can include path parameters

#### **Headers**
- Custom headers per request type
- Ordered headers (important for fingerprinting)
- Dynamic values supported

### **2. TLS Fingerprinting**
- Specify browser to mimic (chrome, firefox, ios, etc.)
- Evades JA3/JA4/JARM detection
- Uses `utls` library

### **3. Timing**
- Configurable callback intervals
- Jitter (randomization)
- Long poll timeouts

### **4. Metadata Encoding**
- Location: header, cookie, uri-param, body
- Encoding: base64, hex, netbios
- Custom header names

---

## 🚀 Usage

### **CLI Usage:**

```bash
# List available profiles
sliver > profiles list

# Generate implant with profile
sliver > generate --http example.com --profile amazon

# Generate with custom profile directory
sliver > generate --http example.com --profile custom --profile-dir /path/to/profiles
```

### **Profile Selection:**

Profiles are loaded from:
1. Built-in profiles (`~/.sliver/profiles/`)
2. Custom directory (`--profile-dir`)
3. Current directory (`./profiles/`)

---

## 📝 Creating Custom Profiles

### **Step 1: Start with Template**

```bash
cp examples/template.yml my-custom-profile.yml
```

### **Step 2: Capture Real Traffic**

```bash
# Capture traffic from target service
tcpdump -i any -w capture.pcap host target-service.com

# Analyze with Wireshark or tshark
tshark -r capture.pcap -Y http -T fields -e http.user_agent
tshark -r capture.pcap -Y http -T fields -e http.request.uri
```

### **Step 3: Customize Profile**

Edit `my-custom-profile.yml` with captured patterns:
- User-Agent strings
- URI paths
- HTTP headers
- Request order

### **Step 4: Validate**

```bash
# Validate against schema
sliver > profile validate my-custom-profile.yml

# Test generation
sliver > generate --http example.com --profile my-custom-profile
```

### **Step 5: Test & Refine**

```bash
# Generate implant
sliver > generate --http example.com --profile my-custom-profile

# Capture traffic
tcpdump -i any -w test.pcap port 8888

# Compare with real traffic
# Iterate and refine
```

---

## 🔍 Profile Examples

### **Amazon AWS SDK**

Mimics AWS SDK traffic for blending with cloud environments.

**Use cases:**
- Cloud-hosted targets
- AWS infrastructure
- Environments with AWS traffic

### **Microsoft Graph API**

Mimics Microsoft 365 API traffic.

**Use cases:**
- Corporate networks
- Office 365 environments
- Microsoft-heavy infrastructure

### **GitHub API**

Mimics GitHub API requests.

**Use cases:**
- Development environments
- CI/CD pipelines
- Git-heavy workflows

---

## ⚠️ Important Notes

### **Limitations:**

1. **Server-Side:** Server responses are not yet customizable
2. **Binary Prepend:** Traffic masking (JS/HTML prepend) not yet implemented
3. **Encoding Chaining:** Multiple encoding layers not supported
4. **Sleep Masking:** Memory obfuscation separate from profiles

### **Best Practices:**

1. **Test First:** Always test profiles before operational use
2. **Update Regularly:** Services change their traffic patterns
3. **Validate Traffic:** Capture and compare with real traffic
4. **Document Changes:** Note why each customization was made
5. **Share Carefully:** Profiles may contain operational details

---

## 🔐 Security Considerations

### **Operational Security:**

- **Profile metadata:** May reveal target or operation details
- **Custom headers:** Don't include real credentials/tokens
- **Comments:** Be careful with operational notes
- **Distribution:** Treat profiles as sensitive

### **Detection Risks:**

- **Inconsistencies:** Mismatched patterns can increase detection
- **Static Patterns:** Rotate profiles regularly
- **Timing Patterns:** Even "random" timing can be fingerprinted
- **TLS + HTTP Mismatch:** Ensure TLS fingerprint matches HTTP patterns

---

## 📊 Profile Effectiveness

### **Measuring Success:**

1. **Network Detection:** Test against EDR/NDR
2. **Traffic Analysis:** Compare pcaps with legitimate traffic
3. **Anomaly Detection:** Check for pattern deviations
4. **Long-term:** Monitor detection rates over time

### **Improvement Cycle:**

```
Create Profile → Test → Capture Traffic → Analyze → Refine → Repeat
```

---

## 🛠️ Development Status

| Feature | Status | Notes |
|---------|--------|-------|
| YAML Parser | 🔄 In Progress | Core functionality |
| Schema Validation | ⏳ Planned | JSON Schema |
| User-Agent Rotation | ⏳ Planned | Random selection |
| URI Patterns | ⏳ Planned | GET/POST patterns |
| Header Customization | ⏳ Planned | Ordered headers |
| TLS Fingerprinting | ✅ Complete | utls integration |
| Timing/Jitter | ⏳ Planned | Interval randomization |
| CLI Integration | ⏳ Planned | --profile flag |
| Profile Validation | ⏳ Planned | Pre-generation check |
| Built-in Profiles | 🔄 In Progress | Service mimicry |

---

## 📚 Further Reading

### **Cobalt Strike Resources:**
- [Malleable C2 Documentation](https://www.cobaltstrike.com/help-malleable-c2)
- Profile examples: `../../Networking/Malleable-C2-Profiles-master/`

### **TLS Fingerprinting:**
- [JA3 Fingerprinting](https://github.com/salesforce/ja3)
- [utls Library](https://github.com/refraction-networking/utls)

### **Traffic Analysis:**
- Wireshark
- tshark
- Zeek
- Suricata

---

## 🤝 Contributing

### **Adding New Profiles:**

1. Create profile in appropriate directory
2. Validate against schema
3. Test generation
4. Document use case
5. Submit PR with examples

### **Profile Naming:**

- **Service profiles:** `service-name.yml` (e.g., `slack.yml`)
- **APT profiles:** `apt-number.yml` (e.g., `apt28.yml`)
- **Custom profiles:** `descriptive-name.yml`

---

## 📞 Support

For questions or issues:
- Check `AUTONOMOUS_WORK_LOG.md` for implementation details
- Review `TESTING_STATUS.md` for testing procedures
- See `MILESTONE_D_COMPLETE.md` for TLS fingerprinting

---

**Status:** 🚧 **Active Development**  
**Last Updated:** 2025-10-24  
**Next:** Profile schema definition & parser implementation

