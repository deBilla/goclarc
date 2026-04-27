---
sidebar_position: 9
---

# E2EE & Auth Architecture

`goclarc crypto` generates a complete end-to-end encryption and authentication layer in three files. This document explains the full cryptographic design, the token lifecycle, the security guarantees, and how to wire everything into your project.

## What Gets Generated

| File | Responsibility |
|---|---|
| `ecies.go` | Asymmetric encryption — X25519 ECDH + HKDF-SHA512 + AES-256-GCM |
| `jwt.go` | Token minting and validation — HS256, access + refresh |
| `redis_tokens.go` | Stateful token rotation and revocation — Redis-backed |

All three files land in one package (default `internal/core/crypto`). They have no dependency on each other except that `redis_tokens.go` calls `jwt.go` functions.

---

## Part 1 — End-to-End Encryption (ecies.go)

### The Model

The E2EE model is asymmetric and server-blind:

1. The **user generates an X25519 keypair on their device** — private key never leaves the client.
2. The **public key is uploaded and stored** in the user record on the server.
3. When the server stores sensitive data, it **encrypts it to the user's public key**.
4. The **server cannot decrypt it** — it holds no corresponding private key.
5. The **client decrypts** locally using the same construction.

```
Client                              Server
──────                              ──────
keygen() → (privKey, pubKey)
            pubKey ──────────────► store in users table
                                   ...
                                   plaintext
                                   + pubKey
                                   │
                                   ▼
                                   Encrypt(pubKey, plaintext)
                                   │
                                   ▼
                                   ciphertext ──────────────► store in DB
...
ciphertext ◄─────────────────────── fetch from DB
privKey + ciphertext
│
▼
Decrypt(privKey, ciphertext)
│
▼
plaintext
```

### Cryptographic Construction — ECIES

The scheme is ECIES (Elliptic Curve Integrated Encryption Scheme) using modern primitives:

| Step | Primitive | Detail |
|---|---|---|
| Key agreement | X25519 (Curve25519) | Ephemeral keypair per message |
| Key derivation | HKDF-SHA512 | Input: ECDH shared secret; Salt: ephemeral public key |
| Encryption | AES-256-GCM | 16-byte nonce, 16-byte auth tag |

### Encrypt

```go
func Encrypt(recipientPub, plaintext []byte) ([]byte, error)
```

1. Generate a random ephemeral X25519 private key.
2. Derive the corresponding ephemeral public key: `ephPub = X25519(ephPriv, Basepoint)`.
3. Compute ECDH shared secret: `shared = X25519(ephPriv, recipientPub)`.
4. Derive a 32-byte AES key via HKDF-SHA512: `key = HKDF(shared, salt=ephPub)`.
5. Generate a random 16-byte IV.
6. Encrypt with AES-256-GCM: produces `ciphertext ‖ tag`.
7. Write wire format.

### Decrypt

```go
func Decrypt(recipientPriv, payload []byte) ([]byte, error)
```

1. Parse wire format: extract `ephPub`, `iv`, `tag`, `ciphertext`.
2. Compute ECDH shared secret: `shared = X25519(recipientPriv, ephPub)`.
3. Derive the same AES key via HKDF-SHA512 (same inputs → same output).
4. Decrypt and verify with AES-256-GCM.

### Wire Format

```
 0       31 32      47 48      63 64 …
┌──────────┬──────────┬──────────┬───────────────┐
│ eph_pub  │    iv    │   tag    │  ciphertext   │
│  32 B    │  16 B    │  16 B    │  N bytes      │
└──────────┴──────────┴──────────┴───────────────┘
```

Fixed overhead per payload: **64 bytes** (`Overhead = 32 + 16 + 16`).

The 16-byte nonce (`cipher.NewGCMWithNonceSize(block, 16)`) is intentional — it matches the frontend wire format so the client can decrypt using the same layout without any byte-swapping.

### Why X25519 + HKDF + AES-GCM?

- **X25519** is faster and more resilient to implementation errors than NIST curves (no point-at-infinity edge case, constant-time by design).
- **HKDF** ensures the raw ECDH output (which is a group element, not uniformly random) is properly stretched into a cryptographic key.
- **AES-256-GCM** provides authenticated encryption — tampering with the ciphertext is detected before any decryption output is returned.
- **Ephemeral keys per message** mean that compromising the recipient's long-term private key does not expose past messages (forward secrecy for stored data).

---

## Part 2 — JWT Authentication (jwt.go)

### Token Design

Two tokens are issued on login:

| Token | TTL | Stateful? | Purpose |
|---|---|---|---|
| Access token | 15 minutes | No | Authenticate API requests |
| Refresh token | 30 days | Yes (Redis) | Obtain new access tokens |

Both are HS256 JWTs with the same `Claims` struct:

```go
type Claims struct {
    jwt.RegisteredClaims
    UserID string `json:"user_id"`
}
```

The access token carries only `sub` (subject), `exp`, `iat`, and `user_id`.
The refresh token additionally carries `jti` — a random 32-character hex string used as the Redis key suffix.

### Functions

```go
// 15-minute access token — stateless, verified by HMAC only.
GenerateAccessToken(secret, userID string) (string, error)

// 30-day refresh token — jti is stored in Redis for revocation.
GenerateRefreshToken(secret, userID, jti string) (string, error)

// Validate signature and return claims. Does NOT check Redis.
ParseToken(secret, tokenStr string) (*Claims, error)
```

`ParseToken` verifies the HMAC signature and expiry but does not consult Redis — that check is done by `ValidateToken` in `redis_tokens.go` (refresh flow only). Access tokens are validated by `ParseToken` alone, which keeps the hot path stateless.

---

## Part 3 — Token Rotation and Revocation (redis_tokens.go)

### Redis Key Pattern

Every refresh token is backed by a Redis entry:

```
rt:{userID}:{jti}  →  "1"  (TTL: 30 days)
```

A user can have multiple active refresh tokens simultaneously (one per device/session). Revoking one entry does not affect others.

### Functions

```go
// Mint an access + refresh token pair. Stores the refresh token in Redis.
IssueTokens(ctx, rdb, jwtSecret, userID string) (accessToken, refreshToken string, err error)

// Parse the refresh token AND verify it exists in Redis. Returns an error if revoked.
ValidateToken(ctx, rdb, jwtSecret, tokenStr string) (*Claims, error)

// Delete the Redis entry, permanently invalidating the refresh token.
RevokeToken(ctx, rdb, userID, jti string) error
```

### Token Lifecycle

```
Register / Login
       │
       ▼
 IssueTokens(userID)
       │
       ├── accessToken  → httpOnly cookie, SameSite=Strict, Secure, MaxAge=15m
       └── refreshToken → httpOnly cookie, SameSite=Strict, Secure, MaxAge=30d
                          Redis: SET rt:{userID}:{jti} "1" EX 2592000
       
Authenticated Request
       │
       ▼
 ParseToken(accessToken)          ← stateless, no Redis
       │
       └── userID → c.Set("userID", userID)

Token Refresh
       │
       ▼
 ValidateToken(refreshToken)      ← checks Redis key exists
       │
       ▼
 RevokeToken(userID, old_jti)     ← DEL rt:{userID}:{old_jti}
       │
       ▼
 IssueTokens(userID)              ← new pair, new jti in Redis

Logout
       │
       ▼
 ParseToken(refreshToken)         ← extract claims
       │
       ▼
 RevokeToken(userID, jti)         ← DEL rt:{userID}:{jti}
       │
       ▼
 Clear cookies
```

### Why Stateless Access + Stateful Refresh?

- Validating access tokens requires no I/O — every authenticated request is fast.
- The short 15-minute window bounds the damage from a stolen access token without needing revocation lists.
- The refresh token is revocable instantly — logout, device wipe, or security incident can invalidate a session before the 30-day TTL expires.
- The `rt:{userID}:{jti}` key structure supports per-device session management: list all active sessions with `SCAN rt:{userID}:*`, revoke individual ones, or revoke all with a single `DEL rt:{userID}:*`.

---

## Security Properties

| Property | Mechanism |
|---|---|
| Server-blind encryption | Server holds only public keys; private keys never leave the client |
| Forward secrecy (stored data) | Ephemeral ECDH keypair per encryption — past ciphertexts not exposed by key compromise |
| Authenticated encryption | AES-GCM auth tag detects tampering before returning any output |
| Stateless fast path | Access token validation is a single HMAC verify — no Redis, no DB |
| Revocable sessions | Redis-backed refresh tokens — logout and device revocation take effect immediately |
| Per-device sessions | `rt:{userID}:{jti}` — revoking one device does not affect others |
| httpOnly cookies | Tokens are not accessible to JavaScript — XSS cannot steal them |

### What This System Does Not Cover

- **Key storage on the client** — the private key must be protected by the client (OS keychain, hardware key, user passphrase). goclarc does not generate key management code.
- **Key rotation** — there is no built-in mechanism to re-encrypt stored data when a user rotates their keypair.
- **Access token revocation** — a stolen access token is valid until its 15-minute expiry. If immediate revocation is needed, shorten the TTL or move to a stateful access token.
- **HTTPS** — the httpOnly cookie protection is meaningless without TLS. Always deploy behind HTTPS.

---

## Integration Guide

### 1. Generate the crypto package

```bash
goclarc crypto
# → internal/core/crypto/{ecies,jwt,redis_tokens}.go
```

### 2. Add required dependencies

```bash
go get golang.org/x/crypto
go get github.com/golang-jwt/jwt/v5
go get github.com/redis/go-redis/v9
```

### 3. Store the user's public key

Add a `public_key` field to your user schema:

```yaml
module: user
fields:
  - name: public_key
    type: string     # base64-encoded 32-byte X25519 public key
    nullable: true
```

### 4. Encrypt data before storing

```go
import "encoding/base64"
import "yourproject/internal/core/crypto"

pubKeyBytes, _ := base64.StdEncoding.DecodeString(user.PublicKey)
ciphertext, err := crypto.Encrypt(pubKeyBytes, []byte(sensitiveData))
if err != nil { ... }
// store base64.StdEncoding.EncodeToString(ciphertext)
```

### 5. Wire tokens in the auth handler

```go
// Login handler (pseudo-code)
func (h *Handler) Login(c *gin.Context) {
    // ... validate credentials ...

    access, refresh, err := crypto.IssueTokens(c.Request.Context(), h.rdb, h.jwtSecret, user.ID)
    if err != nil { _ = c.Error(err); return }

    c.SetCookie("access_token", access, 15*60, "/", "", true, true)
    c.SetCookie("refresh_token", refresh, 30*24*60*60, "/auth/refresh", "", true, true)
    c.JSON(http.StatusOK, gin.H{"success": true})
}
```

### 6. JWT middleware

```go
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token, err := c.Cookie("access_token")
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            return
        }
        claims, err := crypto.ParseToken(jwtSecret, token)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

### 7. Refresh endpoint

```go
func (h *Handler) Refresh(c *gin.Context) {
    token, err := c.Cookie("refresh_token")
    if err != nil { c.AbortWithStatus(http.StatusUnauthorized); return }

    claims, err := crypto.ValidateToken(c.Request.Context(), h.rdb, h.jwtSecret, token)
    if err != nil { c.AbortWithStatus(http.StatusUnauthorized); return }

    // Revoke the old refresh token before issuing a new one.
    _ = crypto.RevokeToken(c.Request.Context(), h.rdb, claims.UserID, claims.ID)

    access, refresh, err := crypto.IssueTokens(c.Request.Context(), h.rdb, h.jwtSecret, claims.UserID)
    if err != nil { _ = c.Error(err); return }

    c.SetCookie("access_token", access, 15*60, "/", "", true, true)
    c.SetCookie("refresh_token", refresh, 30*24*60*60, "/auth/refresh", "", true, true)
    c.JSON(http.StatusOK, gin.H{"success": true})
}
```
