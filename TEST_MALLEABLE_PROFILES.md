# Testing Malleable C2 Profile Integration

## Quick Test Commands

### 1. Drop into Docker Container
```powershell
docker exec -it sliver-server bash
```

### 2. Inside Container - Start Sliver Server
```bash
cd /sliver
./sliver-server daemon &
sleep 2
./sliver-client
```

### 3. Test Profile Loading (in Sliver console)

#### Test 1: Load Amazon Profile
```
generate --http 10.0.0.1:443 --os linux --arch amd64 --malleable-profile amazon --save /tmp/test-amazon
```

**Expected Output:**
```
Applied Malleable C2 profile: Amazon AWS SDK v1.0
  Description: Mimics Amazon AWS SDK HTTP traffic
TLS fingerprinting enabled: mimicking chrome browser
```

#### Test 2: Load by Full Path
```
generate --http 10.0.0.1:443 --os linux --arch amd64 --malleable-profile profiles/services/github.yml --save /tmp/test-github
```

#### Test 3: Test Low-and-Slow Profile
```
generate --http 10.0.0.1:443 --os linux --arch amd64 --malleable-profile low-and-slow --save /tmp/test-stealth
```

#### Test 4: Profile + Explicit TLS Override
```
generate --http 10.0.0.1:443 --os linux --arch amd64 --malleable-profile amazon --tls-fingerprint --tls-browser firefox --save /tmp/test-override
```

**Expected:** Firefox TLS should override Amazon profile's Chrome TLS

### 4. Verify Profile Applied

Check the generated implant config:
```bash
strings /tmp/test-amazon | grep -i "amazon\|aws"
```

## Available Profiles

- **examples/minimal** - Bare minimum profile
- **examples/template** - Full template with all options
- **services/amazon** - AWS SDK traffic (Chrome TLS, 30s interval)
- **services/github** - GitHub API traffic (Chrome TLS, 60s interval)
- **services/microsoft** - Microsoft Graph traffic (Edge TLS, 45s interval)
- **services/slack** - Slack API traffic (Chrome TLS, 20s interval)
- **stealth/low-and-slow** - Operational security (Safari TLS, 300s interval)

## Profile Search Paths

The system searches for profiles in this order:
1. Current directory: `{name}.yml`
2. Profiles directory: `profiles/{name}.yml`
3. Recursive: `profiles/examples/{name}.yml`
4. Recursive: `profiles/services/{name}.yml`
5. Recursive: `profiles/stealth/{name}.yml`
6. User home: `~/.sliver/profiles/{name}.yml`

## What Gets Applied

From the profile:
- ✅ `tls.fingerprint` → `--tls-fingerprint` and `--tls-browser`
- ✅ `timing.interval` → `--reconnect` (if not explicitly set)
- ✅ `timing.poll_timeout` → `--poll-timeout` (if not explicitly set)
- ✅ Profile name stored in `config.MalleableC2Profile`
- ⏳ HTTP settings (user-agents, URIs, headers) - future enhancement

## Parallel Implementation Notes

- `--malleable-profile` is **opt-in** - doesn't affect existing workflows
- Profile settings **don't override** explicit CLI flags
- Works alongside existing `--c2profile` (different system)
- Safe to commit - adds features, doesn't break anything

## Troubleshooting

### "Profile not found"
- Check profile name spelling
- Try full path: `--malleable-profile profiles/services/amazon.yml`
- Verify file exists: `ls profiles/services/`

### "Failed to apply profile"
- Check YAML syntax: `cat profiles/services/amazon.yml`
- Verify required fields (metadata.name)

### No effect visible
- Profile settings won't override explicit CLI flags
- Check if you specified `--reconnect` or `--tls-fingerprint` manually

