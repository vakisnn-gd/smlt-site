package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = 12

var (
	jwtSecret     []byte
	passwordHash  []byte
	allowedOrigin string
	listenAddr    string
	dbURL         string
)

func loadConfig() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("[FATAL] JWT_SECRET env var is required")
	}
	jwtSecret = []byte(secret)

	loadPasswordHash()

	allowedOrigin = os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "https://smlt.lol"
	}

	listenAddr = os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	dbURL = os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("[FATAL] DATABASE_URL env var is required")
	}
}

// loadPasswordHash accepts either a pre-computed bcrypt hash (HOST_PASSWORD_HASH,
// recommended) or a plaintext password (HOST_PASSWORD) that gets hashed at startup.
func loadPasswordHash() {
	hashStr := strings.TrimSpace(os.Getenv("HOST_PASSWORD_HASH"))
	pw := os.Getenv("HOST_PASSWORD")

	switch {
	case hashStr != "":
		h := []byte(hashStr)
		if _, err := bcrypt.Cost(h); err != nil {
			log.Fatal("[FATAL] HOST_PASSWORD_HASH is not a valid bcrypt hash")
		}
		passwordHash = h
	case pw != "":
		cost := bcryptCost()
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), cost)
		if err != nil {
			log.Fatal("Failed to hash password:", err)
		}
		passwordHash = hash
	default:
		log.Fatal("[FATAL] HOST_PASSWORD or HOST_PASSWORD_HASH env var is required")
	}
}

func bcryptCost() int {
	raw := os.Getenv("BCRYPT_COST")
	if raw == "" {
		return defaultBcryptCost
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 4 || n > 31 {
		log.Printf("[WARN] Invalid BCRYPT_COST %q, falling back to %d", raw, defaultBcryptCost)
		return defaultBcryptCost
	}
	return n
}
