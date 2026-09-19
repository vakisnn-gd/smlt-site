package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func generateNonce() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Registry of tokens issued by this process. A JWT is only accepted if its jti
// is present here, so restarting the server revokes every session and stolen
// tokens die with the process. Expired entries are pruned periodically.
var (
	tokenMu      sync.Mutex
	issuedTokens = make(map[string]int64) // jti -> unix exp
)

func rememberToken(jti string, exp int64) {
	tokenMu.Lock()
	issuedTokens[jti] = exp
	tokenMu.Unlock()
}

func tokenIsKnown(jti string) bool {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	exp, ok := issuedTokens[jti]
	return ok && exp > time.Now().Unix()
}

func pruneExpiredTokens() {
	now := time.Now().Unix()
	tokenMu.Lock()
	for jti, exp := range issuedTokens {
		if exp <= now {
			delete(issuedTokens, jti)
		}
	}
	tokenMu.Unlock()
}

func generateToken() (string, error) {
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	jti := uuid.New().String()

	claims := jwt.MapClaims{
		"host": true,
		"exp":  exp.Unix(),
		"iat":  now.Unix(),
		"jti":  jti,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	rememberToken(jti, exp.Unix())
	return signed, nil
}

func validateToken(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	jti, _ := claims["jti"].(string)
	if jti == "" || !tokenIsKnown(jti) {
		return false
	}
	return true
}
