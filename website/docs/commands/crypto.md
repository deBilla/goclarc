---
sidebar_position: 3
---

# goclarc crypto

Generate E2EE cryptography and authentication boilerplate into an existing project.

## Usage

```bash
goclarc crypto [flags]
```

## Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--out` | `-o` | `internal/core/crypto` | Output directory for generated files |
| `--force` | `-f` | `false` | Overwrite existing files |

## Generated Files

```
internal/core/crypto/
  ecies.go          ← X25519 ECDH + HKDF-SHA512 + AES-256-GCM encrypt/decrypt
  jwt.go            ← GenerateAccessToken, GenerateRefreshToken, ParseToken (HS256)
  redis_tokens.go   ← IssueTokens, ValidateToken, RevokeToken (rt:{userID}:{jti})
```

The package name is derived from the output directory (`filepath.Base(--out)`).

## Examples

```bash
# Default — writes to internal/core/crypto/
goclarc crypto

# Custom output directory
goclarc crypto --out pkg/auth/crypto

# Overwrite after template changes
goclarc crypto --force
```

## Architecture

See [E2EE & Auth Architecture](../e2ee-architecture) for a full explanation of the cryptographic design, token lifecycle, security properties, and integration guide.
