# uTLS Integration Strategy - Based on ChatGPT-5 Deep Research

**Date:** 2025-10-25  
**Source:** ChatGPT-5 Agent + Deep Research  
**Status:** Strategic Planning Document

---

## Key Insights from Research

### ✅ What We're Doing RIGHT:

1. **✅ Separate Driver Approach** (Matches our plan)
   - Create NEW `UTLSHTTPDriver` alongside existing drivers
   - DO NOT modify existing Go/WinInet drivers
   - Use `driver=utls` parameter for opt-in selection

2. **✅ Clean Dependencies** (Already done in Phase 2)
   - Added utls v1.8.1 to go-mod ✅
   - Vendored successfully (93 files) ✅
   - Go 1.24.0 > required 1.21+ ✅

3. **✅ Protobuf Fields** (Done in Phase 3)
   - EnableTLSFingerprinting (bool, field 310) ✅
   - TLSFingerprint (string, field 311) ✅
   - Ready for implementation ✅

---

## Critical Implementation Patterns (From Research)

### Pattern 1: DialTLSContext Override
```go
transport := &http.Transport{
    DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
        // 1. TCP dial with timeout
        conn, err := dialer.DialContext(ctx, network, addr)
        if err != nil {
            return nil, err
        }
        
        // 2. Extract SNI from addr (host:port)
        host, _, _ := net.SplitHostPort(addr)
        cfg := &utls.Config{
            InsecureSkipVerify: true,  // Sliver does app-layer encryption
            ServerName: host,
        }
        
        // 3. Wrap with uTLS
        uconn := utls.UClient(conn, cfg, getBrowserID())
        
        // 4. Complete handshake
        if err := uconn.Handshake(); err != nil {
            conn.Close()
            return nil, err
        }
        
        return uconn, nil
    },
}
```

**Key Points:**
- ✅ Respect HTTPOptions.NetTimeout and TlsTimeout
- ✅ Extract SNI from address
- ✅ InsecureSkipVerify: true (Sliver's design)
- ✅ Call Handshake() explicitly
- ✅ Close conn on error

### Pattern 2: ClientHelloID Selection
```go
func getBrowserID(fingerprint string) utls.ClientHelloID {
    switch fingerprint {
    case "chrome":
        return utls.HelloChrome_Auto
    case "firefox":
        return utls.HelloFirefox_Auto
    case "ios":
        return utls.HelloIOS_Auto
    case "randomized":
        return utls.HelloRandomizedALPN
    default:
        return utls.HelloRandomizedALPN  // Safe default
    }
}
```

**Available Fingerprints:**
- `HelloRandomizedALPN` - Random with ALPN (recommended default)
- `HelloRandomizedNoALPN` - Random without ALPN (force HTTP/1.1)
- `HelloChrome_Auto` - Latest Chrome
- `HelloFirefox_Auto` - Latest Firefox  
- `HelloIOS_Auto` - Latest iOS Safari
- Many more browser-specific variants

### Pattern 3: Driver Registration
```go
// In drivers_generic.go (Linux/macOS)
const (
    goHTTPDriver  = "go"
    utlsDriver    = "utls"  // NEW
)

func GetHTTPDriver(origin string, secure bool, opts *HTTPOptions) (HTTPDriver, error) {
    switch opts.Driver {
    case utlsDriver:
        return UTLSHTTPDriver(origin, secure, opts)  // NEW
    case goHTTPDriver:
        return GoHTTPDriver(origin, secure, opts)
    default:
        return GoHTTPDriver(origin, secure, opts)  // Default unchanged
    }
}
```

---

## Common Pitfalls to AVOID (From Research)

### ❌ Issue 1: Import Alias Conflicts
**Problem:** Standard `crypto/tls` conflicts with utls alias
**Solution:**
```go
import (
    utls "github.com/refraction-networking/utls"  // Use 'utls' not 'tls'
    // NO: tls "github.com/refraction-networking/utls"
)
```

### ❌ Issue 2: Not Calling Handshake()
**Problem:** Must explicitly call `uconn.Handshake()`
**Why:** Unlike stdlib, utls doesn't auto-handshake

### ❌ Issue 3: ALPN HTTP/2 Issues
**Problem:** Some proxies don't like HTTP/2
**Solution:** Use `HelloRandomizedNoALPN` or `.NoALPN` variant

### ❌ Issue 4: Breaking Existing Functionality
**Problem:** Modifying existing drivers breaks compatibility
**Solution:** ✅ NEW driver only, existing drivers untouched

---

## Testing Strategy (From Research)

### Phase 3: Protobuf Regeneration
**Status:** Ready to execute
**Command:** `make pb` or Docker build auto-regenerates
**Validation:** Server compiles, no errors

### Phase 4: Minimal Implementation
**Files to Create:**
1. `implant/sliver/transports/httpclient/utls.go` (~150 lines)
   - `getBrowserID()` function
   - `dialTLS()` function  
   - `UTLSHTTPDriver()` function

2. Modify `implant/sliver/transports/httpclient/drivers_generic.go`
   - Add `utlsDriver` constant
   - Add case in `GetHTTPDriver()`

**Validation Steps:**
1. ✅ Server builds
2. ✅ Client builds
3. ✅ **Implant builds** (critical test!)
4. ✅ Implant runs (connects to server)
5. ✅ JA3 hash different from baseline

### Phase 5: CLI Integration
**Files to Modify:**
- `client/command/generate/commands.go` - Add `--tls-fingerprint` flag
- `client/command/generate/generate.go` - Parse flag, set config

**Usage:**
```bash
generate --http example.com:443 --tls-fingerprint chrome
```

---

## Architecture Decision: Template vs Runtime

**Research Finding:** Sliver uses Go templates for implant generation

**Option A: Template-Time Configuration** (CHOSEN)
- Profile applied during `generate` command
- Fingerprint embedded in compiled implant
- ✅ No runtime overhead
- ✅ Single purpose per implant
- ✅ Matches Sliver's design

**Option B: Runtime Configuration**
- Read fingerprint from config/env var
- Switch fingerprints at runtime
- ❌ Adds complexity
- ❌ Not how Sliver works

**Decision:** Use Template-Time (Option A)

---

## Implementation Phases (Revised Based on Research)

### Phase 3: ✅ Protobuf Fields (Complete)
- Fields added ✅
- **Next:** Regenerate protobufs

### Phase 4: Minimal TLS Implementation (2-3h)
**Research-Informed Approach:**

1. **Step 4.1:** Create `utls.go` with:
   - `getBrowserID()` - 20 lines
   - `dialTLS()` - 40 lines
   - `UTLSHTTPDriver()` - 60 lines
   - Total: ~120 lines

2. **Step 4.2:** Modify `drivers_generic.go`:
   - Add constant - 1 line
   - Add case - 2 lines

3. **Step 4.3:** TEST BUILD:
   - ✅ Server compiles
   - ✅ Client compiles
   - ✅ **Implant compiles** (CRITICAL)

4. **Step 4.4:** If successful, commit and proceed

**Estimated Time:** 2-3 hours (with testing)

### Phase 5: CLI Integration (1h)
**Simple Addition:**
- Add `--tls-fingerprint` flag
- Parse and set `EnableTLSFingerprinting=true`
- Set `TLSFingerprint` to user's choice
- Test: `generate --tls-fingerprint chrome`

### Phase 6: End-to-End Validation (2h)
**Full Test:**
1. Generate baseline implant (no TLS fingerprinting)
2. Generate chrome implant (`--tls-fingerprint chrome`)
3. Run both, capture traffic
4. Compare JA3 hashes
5. Verify: chrome ≠ baseline
6. Verify: callbacks work

---

## Key Differences from Archive Branch

| Aspect | Archive (Failed) | V2 (Current) | Research Aligns |
|--------|------------------|--------------|-----------------|
| Testing | No build tests | Test each phase | ✅ Yes |
| Architecture | Modified existing | New driver | ✅ Yes |
| Code Quality | Dead code, bugs | Minimal, focused | ✅ Yes |
| Dependencies | Not validated | Tested first | ✅ Yes |
| Approach | Implement → debug | Research → implement | ✅ Yes |

---

## Success Criteria (From Research)

### Phase 4 Success:
- ✅ Implant builds without errors
- ✅ Code compiles on Linux AND Windows
- ✅ No runtime panics
- ✅ Server still works normally

### Phase 6 Success:
- ✅ Implant connects to C2
- ✅ Commands execute
- ✅ JA3 hash matches chosen browser
- ✅ Different from baseline Go fingerprint

---

## Next Actions

**Immediate (Complete Phase 3):**
1. Regenerate protobufs (Docker or manual)
2. Commit regenerated files
3. Mark Phase 3 complete

**Then (Phase 4):**
1. Review this strategy document ✅
2. Implement `utls.go` based on patterns above
3. Modify `drivers_generic.go`
4. **TEST IMPLANT BUILD** ← Critical validation point
5. Only proceed if build succeeds

---

**Status:** Ready to complete Phase 3 and begin Phase 4 with high confidence based on research. 🚀

