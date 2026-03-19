package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

const (
	_GOOGLE_PROVIDER_ID = -1
	_GITHUB_PROVIDER_ID = -2
	_MANUAL_PROVIDER_ID = -3
)

func (gu *googleUser) Get(ctx context.Context) (u UserInfo, ok bool, err error) {
	gu.log.Debug("googleUser.Get", zap.Bool("ctx_is_nil", ctx == nil))
	return gu.db.GetUser(ctx, gu.log, _GOOGLE_PROVIDER_ID, gu.ID)
}

func (gu *googleUser) Create(ctx context.Context) (u UserInfo, err error) {
	gu.log.Debug("googleUser.Create", zap.Bool("ctx_is_nil", ctx == nil))
	u = UserInfo{
		ID:             uuid.New().String(),
		Email:          gu.Email,
		Picture:        gu.Picture,
		Name:           gu.Name,
		UserProviderID: gu.ID,
		ProviderID:     _GOOGLE_PROVIDER_ID,
	}
	err = gu.db.AddNewProviderUser(ctx, gu.log, u)
	return u, err
}

func (gu *githubUser) Get(ctx context.Context) (u UserInfo, ok bool, err error) {
	gu.log.Debug("githubUser.Get", zap.Bool("ctx_is_nil", ctx == nil))
	return gu.db.GetUser(ctx, gu.log, _GITHUB_PROVIDER_ID, strconv.Itoa(gu.ID))
}

func (gu *githubUser) Create(ctx context.Context) (u UserInfo, err error) {
	gu.log.Debug("githubUser.Create", zap.Bool("ctx_is_nil", ctx == nil))
	u = UserInfo{
		ID:             uuid.New().String(),
		Email:          gu.Email,
		Picture:        gu.AvatarURL,
		Name:           gu.Name,
		UserProviderID: strconv.Itoa(gu.ID),
		ProviderID:     _GITHUB_PROVIDER_ID,
	}
	err = gu.db.AddNewProviderUser(ctx, gu.log, u)
	return u, err
}

func (g *googleWrap) Read(rd io.Reader) (u User, err error) {
	g.log.Debug("googleWrap.Read", zap.Bool("rd_is_nil", rd == nil))
	var data struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
	}
	body, err := io.ReadAll(rd)
	if err != nil {
		return u, err
	}

	err = sonic.Unmarshal(body, &data)
	u = &googleUser{
		ID:      data.ID,
		Email:   data.Email,
		Name:    data.GivenName + " " + data.FamilyName,
		Picture: data.Picture,
		db:      g.db,
		log:     g.log,
	}

	return u, err
}

func (g *githubWrap) Read(rd io.Reader) (u User, err error) {
	g.log.Debug("githubWrap.Read", zap.Bool("rd_is_nil", rd == nil))
	var (
		data struct {
			ID        int    `json:"id"`
			Email     string `json:"email"`
			Name      string `json:"name"`
			AvatarURL string `json:"avatar_url"`
		}
		body []byte
	)

	body, err = io.ReadAll(rd)
	if err != nil {
		return u, err
	}

	err = sonic.Unmarshal(body, &data)
	u = &githubUser{
		ID:        data.ID,
		Email:     data.Email,
		Name:      data.Name,
		AvatarURL: data.AvatarURL,
		db:        g.db,
		log:       g.log,
	}
	return u, err
}

func decodeAccessToken(log *zap.Logger, accessToken string, secret []byte, withTime bool) (c *claims, err error) {
	log.Debug("decodeAccessToken", zap.Int("accessToken_length", len(accessToken)), zap.Int("secret_length", len(secret)), zap.Bool("withTime", withTime))
	parser := jwt.NewParser(
		jwt.WithIssuer("vox_api"),
		jwt.WithAudience("admin"),
		jwt.WithLeeway(5*time.Second),
	)
	if !withTime {
		parser = jwt.NewParser(
			jwt.WithIssuer("vox_api"),
			jwt.WithAudience("admin"),
			jwt.WithLeeway(5*time.Second),
			jwt.WithoutClaimsValidation(),
		)
	}

	token, err := parser.ParseWithClaims(accessToken, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		log.Error("Failed to parse token", zap.Error(err))
		return c, err
	}

	c, ok := token.Claims.(*claims)
	if !ok {
		log.Error("failed to validate token", zap.Error(err))
		return nil, errors.New("failed to validate token")
	}
	log.Debug("Access token decoded", zap.Any("claims", c))

	return c, err
}

func generatePair(log *zap.Logger, userID, secret string) (access, refresh string, err error) {
	log.Debug("generatePair", zap.String("userID", userID), zap.Int("secret_length", len(secret)))
	if secret == "" {
		log.Error("JWT secret is empty")
		return "", "", errors.New("jwt secret is empty")
	}

	now := time.Now().Unix()
	key := []byte(secret)

	claims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "vox_api",
			Audience:  jwt.ClaimStrings{"admin"},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Unix(now, 0).Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Unix(now, 0)),
			NotBefore: jwt.NewNumericDate(time.Unix(now, 0)),
			ID:        uuid.New().String(),
		},
	}

	access, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		log.Error("Failed to sign access token", zap.Error(err))
		return access, refresh, err
	}

	randomBytes := make([]byte, 32)
	if _, err = rand.Read(randomBytes); err != nil {
		log.Error("Failed to generate refresh token", zap.Error(err))
		return access, refresh, err
	}
	refresh = hex.EncodeToString(randomBytes)

	log.Debug("Pair generated", zap.Bool("access_is_empty", access == ""), zap.Bool("refresh_is_empty", refresh == ""), zap.String("user_id", userID))
	return access, refresh, err
}

func isRefreshValid(log *zap.Logger, providedToken, storedHash string) bool {
	log.Debug("isRefreshValid", zap.Int("providedToken_length", len(providedToken)), zap.Int("storedHash_length", len(storedHash)))
	h := sha256.Sum256([]byte(providedToken))
	currentHash := hex.EncodeToString(h[:])
	return subtle.ConstantTimeCompare([]byte(currentHash), []byte(storedHash)) == 1
}

func parseArgon2Hash(log *zap.Logger, stored string) (hash, salt []byte, err error) {
	log.Debug("parseArgon2Hash", zap.Int("stored_length", len(stored)))

	parts := strings.SplitN(stored, "$", 2)
	if len(parts) != 2 {
		err = errors.New("invalid stored hash format: missing separator")
		log.Warn("Failed to parse argon2 hash", zap.Error(err))
		return hash, salt, err
	}

	hashHex, saltHex := parts[0], parts[1]

	if len(hashHex) != 64 {
		err = fmt.Errorf("invalid hash hex length: expected 64, got %d", len(hashHex))
		log.Warn("Failed to parse argon2 hash", zap.Error(err))
		return hash, salt, err
	}
	if len(saltHex) != 32 {
		err = fmt.Errorf("invalid salt hex length: expected 32, got %d", len(saltHex))
		log.Warn("Failed to parse argon2 hash", zap.Error(err))
		return hash, salt, err
	}

	hash, err = hex.DecodeString(hashHex)
	if err != nil {
		hash = nil
		log.Warn("Failed to decode argon2 hash", zap.Error(err))
		return hash, salt, err
	}

	salt, err = hex.DecodeString(saltHex)
	if err != nil {
		hash = nil
		salt = nil
		log.Warn("Failed to decode argon2 salt", zap.Error(err))
		return hash, salt, err
	}

	log.Debug("Stored hash and salt parsed successfully",
		zap.Bool("hash_is_nil", hash == nil),
		zap.Bool("salt_is_nil", salt == nil))
	return hash, salt, err
}

func verifyPassword(log *zap.Logger, storedHash, providedPassword string) (err error) {
	log.Debug("verifyPassword", zap.Int("storedHash_length", len(storedHash)), zap.Int("providedPassword_length", len(providedPassword)))
	decoded, salt, err := parseArgon2Hash(log, storedHash)
	if err != nil {
		return err
	}
	computed := argon2.IDKey([]byte(providedPassword), salt, 1, 64*1024, 4, 32)
	if subtle.ConstantTimeCompare(computed, decoded) != 1 {
		err = errors.New("not verified")
		log.Warn("Not verified")
		return err
	}
	log.Debug("Password verified")
	return err
}

func generateArgon2Hash(password string) ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	result := hex.EncodeToString(hash) + "$" + hex.EncodeToString(salt)
	return []byte(result), nil
}
