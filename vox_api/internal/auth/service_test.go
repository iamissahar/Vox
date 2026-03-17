//go:build unit

package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/crypto/argon2"
)

func TestDecodeAccessToken(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Now()

	makeToken := func(c *claims, secret []byte) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
		signed, _ := token.SignedString(secret)
		return signed
	}

	baseClaims := &claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "vox_api",
			Audience:  jwt.ClaimStrings{"admin"},
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        "some-unique-id",
		},
	}

	expiredClaims := &claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "vox_api",
			Audience:  jwt.ClaimStrings{"admin"},
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        "some-unique-id",
		},
	}

	cases := []struct {
		name       string
		tokenF     func() string
		secretF    func() []byte
		withTime   bool
		wantErr    bool
		wantClaims *claims
	}{
		{
			name:       "valid token",
			tokenF:     func() string { return makeToken(baseClaims, secret) },
			secretF:    func() []byte { return secret },
			withTime:   true,
			wantErr:    false,
			wantClaims: baseClaims,
		},
		{
			name:       "wrong secret",
			tokenF:     func() string { return makeToken(baseClaims, []byte("original-secret")) },
			secretF:    func() []byte { return []byte("wrong-secret") },
			withTime:   true,
			wantErr:    true,
			wantClaims: nil,
		},
		{
			name:       "expired token with time check",
			tokenF:     func() string { return makeToken(expiredClaims, secret) },
			secretF:    func() []byte { return secret },
			withTime:   true,
			wantErr:    true,
			wantClaims: nil,
		},
		{
			name:       "expired token without time check",
			tokenF:     func() string { return makeToken(expiredClaims, secret) },
			secretF:    func() []byte { return secret },
			withTime:   false,
			wantErr:    false,
			wantClaims: expiredClaims,
		},
		{
			name: "wrong issuer",
			tokenF: func() string {
				return makeToken(&claims{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    "someone_else",
						Audience:  jwt.ClaimStrings{"admin"},
						ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
					},
				}, secret)
			},
			secretF:    func() []byte { return secret },
			withTime:   true,
			wantErr:    true,
			wantClaims: nil,
		},
		{
			name:       "garbage token",
			tokenF:     func() string { return "not.a.token" },
			secretF:    func() []byte { return secret },
			withTime:   true,
			wantErr:    true,
			wantClaims: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := zaptest.NewLogger(t)
			c, err := decodeAccessToken(log, tc.tokenF(), tc.secretF(), tc.withTime)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, c)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, c)
				assert.Equal(t, tc.wantClaims.Issuer, c.Issuer)
				assert.Equal(t, tc.wantClaims.Audience, c.Audience)
				assert.Equal(t, tc.wantClaims.Subject, c.Subject)
				assert.Equal(t, tc.wantClaims.RegisteredClaims.Issuer, c.RegisteredClaims.Issuer)
				assert.Equal(t, tc.wantClaims.RegisteredClaims.Subject, c.RegisteredClaims.Subject)
				assert.Equal(t, tc.wantClaims.RegisteredClaims.Audience, c.RegisteredClaims.Audience)
				assert.Equal(t, tc.wantClaims.ID, c.ID)
			}
		})
	}
}

func TestGeneratePair(t *testing.T) {
	secret := "test-secret"

	cases := []struct {
		name        string
		userID      string
		secret      string
		wantErr     bool
		assertExtra func(t *testing.T, access, refresh string)
	}{
		{
			name:    "valid pair",
			userID:  "user-123",
			secret:  secret,
			wantErr: false,
			assertExtra: func(t *testing.T, access, refresh string) {
				assert.NotEmpty(t, access)
				assert.NotEmpty(t, refresh)
				assert.Len(t, refresh, 64)
			},
		},
		{
			name:    "tokens are unique per call",
			userID:  "user-123",
			secret:  secret,
			wantErr: false,
			assertExtra: func(t *testing.T, access, refresh string) {
				log := zaptest.NewLogger(t)
				access2, refresh2, err := generatePair(log, "user-123", secret)
				assert.NoError(t, err)
				assert.NotEqual(t, access, access2)
				assert.NotEqual(t, refresh, refresh2)
			},
		},
		{
			name:    "access token decodes correctly",
			userID:  "user-123",
			secret:  secret,
			wantErr: false,
			assertExtra: func(t *testing.T, access, refresh string) {
				log := zaptest.NewLogger(t)
				c, err := decodeAccessToken(log, access, []byte(secret), true)
				assert.NoError(t, err)
				require.NotNil(t, c)
				assert.Equal(t, "user-123", c.RegisteredClaims.Subject)
				assert.Equal(t, "vox_api", c.RegisteredClaims.Issuer)
				assert.Equal(t, jwt.ClaimStrings{"admin"}, c.RegisteredClaims.Audience)
				assert.NotEmpty(t, c.ID)
				assert.NotEmpty(t, refresh)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := zaptest.NewLogger(t)
			access, refresh, err := generatePair(log, tc.userID, tc.secret)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, access)
				assert.Empty(t, refresh)
			} else {
				assert.NoError(t, err)
				tc.assertExtra(t, access, refresh)
			}
		})
	}
}

func TestParseArgon2Hash(t *testing.T) {
	validHash := strings.Repeat("a", 64) // 64 hex chars = 32 bytes
	validSalt := strings.Repeat("b", 32) // 32 hex chars = 16 bytes
	validStored := fmt.Sprintf("%s$%s", validHash, validSalt)

	cases := []struct {
		name        string
		stored      string
		wantErr     bool
		assertExtra func(t *testing.T, hash, salt []byte)
	}{
		{
			name:    "valid stored hash",
			stored:  validStored,
			wantErr: false,
			assertExtra: func(t *testing.T, hash, salt []byte) {
				expectedHash, _ := hex.DecodeString(validHash)
				expectedSalt, _ := hex.DecodeString(validSalt)
				assert.Equal(t, expectedHash, hash)
				assert.Equal(t, expectedSalt, salt)
				assert.Len(t, hash, 32)
				assert.Len(t, salt, 16)
			},
		},
		{
			name:        "wrong format no separator",
			stored:      strings.Repeat("a", 96), // no $ separator
			wantErr:     true,
			assertExtra: nil,
		},
		{
			name:        "wrong hash length",
			stored:      fmt.Sprintf("%s$%s", strings.Repeat("a", 32), validSalt), // 32 instead of 64
			wantErr:     true,
			assertExtra: nil,
		},
		{
			name:        "wrong salt length",
			stored:      fmt.Sprintf("%s$%s", validHash, strings.Repeat("b", 16)), // 16 instead of 32
			wantErr:     true,
			assertExtra: nil,
		},
		{
			name:        "invalid hex in hash",
			stored:      fmt.Sprintf("%s$%s", strings.Repeat("z", 64), validSalt), // z is not hex
			wantErr:     true,
			assertExtra: nil,
		},
		{
			name:        "invalid hex in salt",
			stored:      fmt.Sprintf("%s$%s", validHash, strings.Repeat("z", 32)), // z is not hex
			wantErr:     true,
			assertExtra: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := zaptest.NewLogger(t)
			hash, salt, err := parseArgon2Hash(log, tc.stored)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, hash)
				assert.Nil(t, salt)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, hash)
				require.NotNil(t, salt)
				tc.assertExtra(t, hash, salt)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	// mirrors exactly how your real code would store a password
	makeStoredHash := func(password string) string {
		salt := make([]byte, 16)
		_, _ = rand.Read(salt)
		hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
		return fmt.Sprintf("%s$%s", hex.EncodeToString(hash), hex.EncodeToString(salt))
	}

	correctPassword := "supersecret"
	validStored := makeStoredHash(correctPassword)

	cases := []struct {
		name             string
		storedHash       string
		providedPassword string
		wantErr          bool
	}{
		{
			name:             "correct password",
			storedHash:       validStored,
			providedPassword: correctPassword,
			wantErr:          false,
		},
		{
			name:             "wrong password",
			storedHash:       validStored,
			providedPassword: "wrongpassword",
			wantErr:          true,
		},
		{
			name:             "empty password",
			storedHash:       validStored,
			providedPassword: "",
			wantErr:          true,
		},
		{
			name:             "invalid stored hash",
			storedHash:       "not-a-valid-hash",
			providedPassword: correctPassword,
			wantErr:          true,
		},
		{
			name:             "empty stored hash",
			storedHash:       "",
			providedPassword: correctPassword,
			wantErr:          true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := zaptest.NewLogger(t)
			err := verifyPassword(log, tc.storedHash, tc.providedPassword)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
