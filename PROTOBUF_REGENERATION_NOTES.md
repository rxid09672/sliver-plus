# Protobuf Regeneration Notes

**Date:** 2025-10-23  
**Branch:** feature/utls-integration  
**Changes:** Added TLS fingerprinting fields to ImplantConfig

---

## ⚠️ IMPORTANT: Protobuf Changes Made

The following protobuf fields were added to `protobuf/clientpb/client.proto`:

```protobuf
message ImplantConfig {
  // ... existing fields ...
  
  // TLS Fingerprinting (Milestone D - Network Evasion, fields 310+)
  bool EnableTLSFingerprinting = 310;
  string TLSFingerprint = 311;  // "chrome", "firefox", "ios", "android", "edge", "safari"
  string MalleableC2Profile = 312;  // Path to Malleable C2 profile file (future)
}
```

---

## 🐧 Regeneration Required on Linux

**These changes MUST be regenerated on Linux/WSL/macOS before they work!**

### Why?
- Windows doesn't have `make` or `protoc` toolchain
- Sliver's `Makefile` requires Unix environment
- Protocol buffer compiler needs to be installed

### How to Regenerate:

```bash
# On Linux/WSL/macOS

# 1. Navigate to sliver-master directory
cd /path/to/sliver-master

# 2. Ensure protoc is installed
which protoc  # Should show path to protoc binary
# If not installed: apt-get install -y protobuf-compiler (Debian/Ubuntu)

# 3. Regenerate protocol buffers
make pb

# 4. Verify generated files updated
git status
# Should show changes to:
#   - protobuf/clientpb/client.pb.go
#   - Other .pb.go files if dependencies changed

# 5. Commit the generated files
git add protobuf/
git commit -m "chore: Regenerate protobufs for TLS fingerprinting"
```

---

## 🔍 What Gets Generated

Running `make pb` will regenerate:

1. **`protobuf/clientpb/client.pb.go`** - Go structs for ImplantConfig
2. **Other `.pb.go` files** - If cross-references changed

These generated files contain:
- Go struct definitions matching the proto schema
- Marshal/Unmarshal methods
- Field getters/setters
- Protobuf encoding/decoding logic

---

## ✅ Verification

After regeneration, verify the fields are accessible in Go:

```go
config := &clientpb.ImplantConfig{
    EnableTLSFingerprinting: true,
    TLSFingerprint:          "chrome",
    MalleableC2Profile:      "",
}

// Should compile without errors
```

---

## 📝 Development Workflow

**On Windows (development):**
1. Edit `.proto` files
2. Document changes
3. Write Go code that uses new fields (will error until regenerated)
4. Commit proto changes

**On Linux/Docker (regeneration):**
1. Pull changes
2. Run `make pb`
3. Verify generated code
4. Commit `.pb.go` files

**Back on Windows (continue development):**
1. Pull generated code
2. Go code now compiles
3. Continue implementation

---

## 🐳 Docker Alternative

If you don't have Linux, use Docker:

```bash
# From Windows
cd C:\Users\maste\Documents\Projects\dshc22\sliver-master

# Run in Docker
docker run --rm -v "${PWD}:/sliver" -w /sliver golang:1.25-alpine sh -c "
  apk add --no-cache make protobuf-dev git &&
  make pb
"

# Verify changes
git status
```

---

## ⚠️ Current Status

**Proto Changes:** ✅ Complete  
**Regeneration:** ⏳ **PENDING - MUST BE DONE BEFORE USE**  
**Go Code:** Can be written, but won't compile until regeneration

---

## 🔗 References

- Sliver Makefile: `sliver-master/Makefile` (see `pb` target)
- Protobuf docs: https://protobuf.dev/
- Sliver AGENTS.md: Repository guidelines

---

**Last Updated:** 2025-10-23 23:50 UTC  
**Next Action:** Regenerate protobufs on Linux before continuing with Go implementation

