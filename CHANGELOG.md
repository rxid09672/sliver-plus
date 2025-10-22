# Changelog - Sliver-Plus

All notable additions and enhancements to the original [Sliver by Bishop Fox](https://github.com/BishopFox/sliver) are documented in this file.

---

## [Unreleased] - 2025-10-22

### About This Fork

**Sliver-Plus** is an enhanced fork of [Sliver by Bishop Fox](https://github.com/BishopFox/sliver). All credit for the core Sliver framework goes to Bishop Fox and the Sliver contributors. This fork adds reconnaissance and adversary simulation capabilities while maintaining the original Sliver architecture and code quality standards.

**Repository:** https://github.com/rxid09672/sliver-plus  
**Upstream:** https://github.com/BishopFox/sliver  
**License:** GPL v3 (same as Sliver)

---

## Added Features

### 🔍 Active Network Scanning

#### `hosts scan` Command

**Purpose:** Network reconnaissance from the operator machine with automatic host discovery.

**Usage:**
```bash
# Scan single IP
sliver> hosts scan 192.168.1.1

# Scan CIDR range with service detection
sliver> hosts scan 10.0.0.0/24 -s

# Full scan with OS detection
sliver> hosts scan 192.168.1.0/24 -s -o -p 1-65535
```

**Features:**
- ✅ **Nmap Integration** - Uses nmap for battle-tested network scanning
- ✅ **Service Detection** - Identify running services with `-s` flag
- ✅ **OS Fingerprinting** - Detect operating systems with `-o` flag
- ✅ **Custom Port Ranges** - Specify ports with `-p` flag (e.g., `22,80,443` or `1-1000`)
- ✅ **Auto-Population** - Discovered hosts automatically added to Sliver's hosts database
- ✅ **Pretty Output** - Results displayed in formatted tables
- ✅ **Graceful Errors** - Clear messages if nmap not installed
- ✅ **Timeout Control** - Configure scan timeout with `--scan-timeout`

**CLI Flags:**
- `-p, --ports <spec>` - Port specification (default: common ports)
- `-s, --service-detection` - Enable service version detection
- `-o, --os-detection` - Enable OS fingerprinting
- `--scan-timeout <seconds>` - Scan timeout (default: 300)

**Storage:** Results stored in Sliver's `hosts` database with services in `ExtensionData`.

**Runtime Dependency:** `nmap` (installation guide in [RUNTIME_DEPENDENCIES.md](docs-sliver-plus/RUNTIME_DEPENDENCIES.md))

**Files Added:**
- `server/core/hosts_scan.go` - Core scanning logic (~445 lines)
- `client/command/hosts/hosts-scan.go` - CLI interface (~145 lines)

**Files Modified:**
- `server/rpc/rpc-hosts.go` - Added `HostScan` RPC handler
- `client/command/hosts/commands.go` - Registered `hosts scan` subcommand
- `protobuf/clientpb/client.proto` - Added `HostScanReq`, `HostScanResult`, `HostScanResults` messages
- `protobuf/rpcpb/services.proto` - Added `HostScan` RPC method

---

### 🕵️ Passive OSINT Reconnaissance

#### `loot recon` Command

**Purpose:** Passive intelligence gathering from public sources with automatic loot storage.

**Usage:**
```bash
# Subdomain enumeration
sliver> loot recon domains example.com

# Email harvesting
sliver> loot recon emails example.com

# Combined reconnaissance
sliver> loot recon all example.com

# With custom timeout
sliver> loot recon domains example.com --timeout 600
```

**Features:**
- ✅ **Subdomain Enumeration** - Discover subdomains using subfinder
- ✅ **Email Harvesting** - Find email addresses using theHarvester
- ✅ **Combined Mode** - Run all reconnaissance types at once
- ✅ **Auto-Storage** - Results automatically stored in Sliver's loot system
- ✅ **Tool Detection** - Graceful handling if tools not installed
- ✅ **JSON Parsing** - Native parsing with text fallback
- ✅ **Pretty Output** - Clear result summaries
- ✅ **Timeout Control** - Configure reconnaissance timeout

**CLI Flags:**
- `--timeout <seconds>` - Reconnaissance timeout (default: 300)

**Reconnaissance Types:**
- `domains` - Subdomain enumeration (subfinder)
- `emails` - Email address harvesting (theHarvester)
- `all` - Run all reconnaissance types

**Storage:** Results stored in Sliver's `loot` system as type `recon/domains`, `recon/emails`, or `recon/all` with full JSON data.

**Loot Viewing:**
```bash
# List all reconnaissance loot
sliver> loot --filter recon

# Fetch specific reconnaissance data
sliver> loot fetch <loot-id>
```

**Runtime Dependencies:**
- `subfinder` - For subdomain enumeration
- `theHarvester` - For email/OSINT gathering

**Files Added:**
- `server/core/recon.go` - OSINT reconnaissance logic (~350 lines)
- `client/command/loot/loot-recon.go` - CLI interface (~140 lines)

**Files Modified:**
- `server/rpc/rpc-loot.go` - Added `Recon` RPC handler with loot event publishing
- `client/command/loot/commands.go` - Registered `loot recon` subcommand
- `protobuf/clientpb/client.proto` - Added `ReconReq`, `ReconResult` messages
- `protobuf/rpcpb/services.proto` - Added `Recon` RPC method

---

### 🧬 Build Diversity CLI Integration

#### Enhanced `generate` Command

**Purpose:** Expose metamorphic code generation capabilities through intuitive CLI flags.

**Research Background:**

The metamorphic code generation techniques implemented in Sliver-Plus are inspired by extensive analysis of historical malware from the **[VX-Underground](https://www.vx-underground.org/) malware source code archive** (1,497 ASM files studied). Special credit to:

**Primary Inspiration:**
- **Virus.Win32.Morpher** by x0man (VirusTech, 2008) - Two-pass metamorphic engine with instruction expansion and dead code injection
- Techniques reverse-engineered from assembly source and reimplemented in Go with modern safety guarantees

**Additional Study:**
- **Win32.Metaphor** (2000) - "Accordion model" metamorphic permutation
- **Win32.CJD** - Entry-point obscuring and permutation techniques
- **Win32.Filly** - Flag-dependent runtime code generation
- **Win32.Flying (FinE)** - Structured morphing with formal parameters
- **Win32.badf00d** - Polymorphic engine architecture
- Various other polymorphic engines from the VX-Underground archive

**Novel Contributions:**

1. **Lito Length Disassembler** (~1,030 lines)
   - Inspired by malware engines' need for instruction length parsing
   - Modern Go implementation with x64 REX prefix support
   - O(1) table-based lookups, type-safe enums
   - Used by both Morpher and Wig engines

2. **Morpher Metamorphic Engine** (~1,120 lines)
   - Go reimplementation of Win32.Morpher techniques
   - Two-pass algorithm (structure mutation + relocation)
   - 30+ NOP-equivalent dead code patterns
   - Xorshift-128 crypto-quality RNG
   - Instruction expansion (SHORT jumps → NEAR jumps)
   - Safe relocation with validation

3. **Wig Vector Space Metamorphism** (~1,640 lines)
   - **Novel approach** not found in historical malware
   - 100-dimensional code space exploration
   - Chain-of-thought reasoning (5 thought types)
   - Executable manifold constraint satisfaction
   - Alien pattern generation (maximally foreign to detectors)
   - 4 distance metrics for novelty scoring

**Acknowledgments:**
- **VX-Underground** for preserving historical malware source code for security research
- **x0man (VirusTech)** for the original Morpher engine design
- Historical virus writers whose techniques advanced the field of code obfuscation
- All work reimplemented with safety guarantees and modern error handling

**Usage:**
```bash
# Basic diversity (Morpher metamorphism)
sliver> generate --os windows --arch amd64 --mtls 10.0.0.1:8888 --diversity

# Vector Space Metamorphism (Wig)
sliver> generate --os windows --arch amd64 --mtls 10.0.0.1:8888 --diversity --novel-techniques

# Reproducible builds
sliver> generate --os windows --arch amd64 --mtls 10.0.0.1:8888 --diversity --seed "operation-alpha"

# Custom novelty threshold
sliver> generate --os windows --arch amd64 --mtls 10.0.0.1:8888 --diversity --novel-techniques --min-novelty 0.8
```

**New CLI Flags:**
- `--diversity` - Enable metamorphic code generation (default: false)
- `--novel-techniques` - Use Wig Vector Space Metamorphism for maximum evasion
- `--seed <string>` - Seed for reproducible builds
- `--min-novelty <float>` - Minimum novelty score for Wig (0.0-1.0, default: 0.7)

**Features:**
- ✅ **Two-Tier Diversity** - Choose between Morpher (traditional) or Wig (revolutionary)
- ✅ **Reproducible Builds** - Use `--seed` for identical builds across operators
- ✅ **Novelty Control** - Adjust `--min-novelty` for desired evasion level
- ✅ **Validation** - User-friendly error messages for invalid combinations
- ✅ **Server Logging** - All diversity configurations logged server-side

**Validation Rules:**
- `--novel-techniques` requires `--diversity` to be enabled
- `--min-novelty` requires both `--diversity` and `--novel-techniques`

**Files Modified:**
- `client/command/generate/commands.go` - Added diversity flags
- `client/command/generate/generate.go` - Added diversity config parsing and validation
- `server/rpc/rpc-generate.go` - Added diversity configuration logging
- `protobuf/clientpb/client.proto` - Added `MinNoveltyScore` to `BuildDiversityConfig`

---

## 📊 Statistics

### Code Metrics

**Total Lines Added:** ~1,680 lines (excluding documentation)
- Server Code: ~845 lines
- Client Code: ~385 lines
- Protobuf: ~60 lines

**New Files:** 4 Go source files
- `server/core/hosts_scan.go`
- `server/core/recon.go`
- `client/command/hosts/hosts-scan.go`
- `client/command/loot/loot-recon.go`

**Modified Files:** 8 files
- 2 Protobuf definitions (client.proto, services.proto)
- 2 Server RPC handlers (rpc-hosts.go, rpc-loot.go)
- 4 Client command files (generate/, hosts/, loot/)

**Generated Files:** 2 protobuf bindings (regenerated from .proto)

---

## 🛠️ Runtime Dependencies

### Required Tools (for runtime, not compilation)

| Tool | Purpose | Command | Installation |
|------|---------|---------|--------------|
| **nmap** | Active scanning | `hosts scan` | `apt install nmap` / `brew install nmap` |
| **subfinder** | Subdomain enum | `loot recon domains` | `go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest` |
| **theHarvester** | Email/OSINT | `loot recon emails` | `pip3 install theHarvester` |

**Note:** These tools are NOT required for building Sliver-Plus, only for executing the corresponding commands.

**Graceful Handling:** If tools are not found, commands display clear installation instructions and exit gracefully without crashing.

---

## 🎯 Design Principles

All enhancements follow these principles to maintain Sliver's code quality:

### ✅ Augmentation Over Porting
- Extended existing commands (`hosts`, `loot`, `generate`)
- No parallel tracking systems
- No new database tables
- Reused existing Sliver infrastructure

### ✅ Native Go Implementation
- Direct tool execution via `exec.Command`
- Native JSON parsing
- Minimal external dependencies
- No Python wrappers for these features

### ✅ Sliver-Native Feel
- Followed existing command patterns
- Used existing database schemas (hosts, loot)
- Consistent error handling conventions
- Standard table formatting

### ✅ Code Quality First
- Small, focused functions
- Comprehensive error handling
- Clear separation of concerns
- Graceful degradation
- Well-documented code

### ✅ Backwards Compatibility
- Zero breaking changes to original Sliver
- All existing commands work unchanged
- Original Sliver functionality preserved
- Can sync with upstream Sliver

---

## 📚 Documentation

### Comprehensive Guides

All documentation is available in `dshc2/docs/`:

- **PROJECT_STATUS.md** - Overall project status and metrics
- **RUNTIME_DEPENDENCIES.md** - Tool installation guide (comprehensive, ~350 lines)
- **MILESTONE_D_FINAL_SUMMARY.md** - CLI integration details
- **MILESTONE_E_STATUS.md** - Reconnaissance features status
- **CHUNK_3_COMPLETE.md** - `hosts scan` implementation details
- **CHUNK_4_COMPLETE.md** - `loot recon` implementation details
- **GIT_COMMIT_GUIDE.md** - Repository setup and commit guide

---

## 🚧 Known Limitations

### Current Implementation

1. **Host Deduplication**
   - `hosts scan` always creates new host entries
   - No IP-based deduplication yet
   - Workaround: Manual reconciliation via `hosts` command
   - Future: Add IP-based matching

2. **No Real-Time Progress**
   - Long scans/recon appear to hang
   - No streaming progress updates
   - Workaround: Informational messages ("This may take a while...")
   - Future: Add streaming progress via gRPC

3. **theHarvester Output Parsing**
   - Different versions produce different JSON formats
   - Fallback text parsing implemented
   - May miss some results on newer versions
   - Future: Better version detection

4. **Tool Dependencies**
   - External tools required (nmap, subfinder, theHarvester)
   - Not bundled with Sliver
   - Must be installed separately
   - See RUNTIME_DEPENDENCIES.md for guides

---

## 🔮 Future Enhancements

### Planned Features (Not Yet Implemented)

#### Additional OSINT Tools
- Amass integration (comprehensive subdomain enum)
- WHOIS lookup capability
- Certificate transparency log queries
- Shodan API integration

#### BloodHound Integration
- `bloodhound ingest <file>` command
- Parse BloodHound JSON collectors (SharpHound, AzureHound)
- Create host entries from AD computers
- Store AD relationships
- Enhance `pivots graph` with AD edges
- Query attack paths

#### Implant-Side Scanning
- Run `hosts scan` from compromised implants
- Pivot network scanning
- Local subnet discovery
- Service enumeration from inside networks

#### Streaming Progress
- Real-time progress updates for long operations
- Percentage completion indicators
- Cancel/abort support
- Live tool output streaming

#### Result Enrichment
- Cross-reference with threat intelligence
- CVE lookups for discovered services
- Shodan/Censys enrichment
- GeoIP location data

---

## 🎊 Milestones Completed

### Milestone D: CLI Integration ✅
- **Status:** 100% Complete
- **Time:** ~2 hours
- **Lines:** ~61 lines (integration)
- **Features:** 4 CLI flags for build diversity

### Milestone E: Reconnaissance ✅ (67% Complete)
- **Status:** 67% Complete (4 of 6 chunks)
- **Time:** ~1.5 hours
- **Lines:** ~1,080 lines
- **Features:** 2 major capabilities
  - ✅ Active scanning (`hosts scan`)
  - ✅ Passive OSINT (`loot recon`)
  - 📝 BloodHound integration (planned)

---

## 🤝 Contributing

This is a fork/enhancement of [Bishop Fox's Sliver](https://github.com/BishopFox/sliver). 

**Upstream:** https://github.com/BishopFox/sliver  
**This Fork:** https://github.com/rxid09672/sliver-plus

To contribute:
1. Fork this repository
2. Create a feature branch
3. Make your changes following the design principles above
4. Test thoroughly
5. Submit a pull request

---

## 📜 License

Same as Sliver - **GPL v3**

This fork maintains the same license as the original Sliver project by Bishop Fox.

---

## 🙏 Acknowledgments

### Core Framework
**Bishop Fox and Sliver Contributors:**
- Original Sliver framework and architecture
- Excellent C2 capabilities and implant design
- Clean codebase and documentation
- Open source community leadership

### Metamorphic Code Generation Research
**VX-Underground:**
- Malware source code preservation (https://www.vx-underground.org/)
- 1,497 ASM file dataset used for metamorphic engine research
- Critical resource for understanding historical evasion techniques

**Malware Researchers & Virus Writers:**
- **x0man (VirusTech)** - Virus.Win32.Morpher metamorphic engine (2008)
- **Metaphor authors** - "Accordion model" permutation techniques (2000)
- **CJD, Filly, Flying, badf00d authors** - Various polymorphic engine innovations
- All historical virus writers whose published techniques advanced code obfuscation research

**Note:** All malware techniques were studied from publicly available source code archives for defensive security research purposes. All implementations are original Go code with modern safety guarantees and error handling. No malicious code was copied verbatim.

### Tool Integration
**ProjectDiscovery:**
- subfinder for subdomain enumeration (https://github.com/projectdiscovery/subfinder)

**theHarvester Team:**
- theHarvester for OSINT gathering (https://github.com/laramies/theHarvester)

**Nmap Project:**
- nmap for network scanning (https://nmap.org/)

### This Fork Contributions
- Enhanced reconnaissance capabilities (hosts scan, loot recon)
- OSINT integration with existing Sliver loot system
- Active network scanning with database auto-population
- Metamorphic engine Go implementations (Lito, Morpher, Wig)
- Build diversity CLI exposure and integration

---

## 📞 Support

**For Sliver-Plus Issues:**
- Create an issue on this repository: https://github.com/rxid09672/sliver-plus/issues

**For Original Sliver Issues:**
- Check the official Sliver documentation: https://sliver.sh/docs
- Visit the Sliver repository: https://github.com/BishopFox/sliver
- Join the Bloodhound Gang Slack: https://bloodhoundgang.herokuapp.com/

---

## 📅 Version History

### v0.1.0 - 2025-10-22 (Unreleased)

**Added:**
- `hosts scan` command for active network scanning
- `loot recon` command for passive OSINT
- Build diversity CLI flags for `generate` command
- Comprehensive documentation

**Technical:**
- 4 new Go source files (~1,080 lines)
- 8 modified files (protobuf, RPC, CLI)
- Full protobuf integration
- Native nmap, subfinder, theHarvester integration

---

**Last Updated:** October 22, 2025  
**Maintainer:** Sliver-Plus Team  
**Based On:** Sliver by Bishop Fox (https://github.com/BishopFox/sliver)

