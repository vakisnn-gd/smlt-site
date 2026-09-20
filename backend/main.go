package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	loadConfig()
	initCaptchaWords()
	initDB()
	defer db.Close()

	log.Printf("SMLT Leaderboard starting on %s", listenAddr)
	log.Printf("CORS: %s", allowedOrigin)

	go cleanupExpired()
	go startDemonlistSync()

	frontendDir := os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		frontendDir = "../frontend-dist"
	}
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		frontendDir = "../frontend"
	}
	fs := http.FileServer(http.Dir(frontendDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sitemap.xml" {
			serveSitemap(w, r)
			return
		}
		if r.URL.Path == "/players" || strings.HasPrefix(r.URL.Path, "/player/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path != "/" {
			full := filepath.Join(frontendDir, filepath.FromSlash(path.Clean("/"+r.URL.Path)))
			if info, err := os.Stat(full); err == nil && !info.IsDir() {
				if strings.HasPrefix(r.URL.Path, "/assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "public, max-age=86400")
				}
				fs.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
	})

	sec := func(h http.HandlerFunc) http.HandlerFunc {
		return corsMiddleware(h)
	}

	http.HandleFunc("/api/players", sec(bodyLimitMiddleware(handlePlayersCollection)))
	http.HandleFunc("/api/players/", sec(bodyLimitMiddleware(authMiddleware(handlePlayerByMethod))))
	http.HandleFunc("/api/recent-changes", sec(methodCheck("GET", handleRecentChanges)))
	http.HandleFunc("/api/events", sec(bodyLimitMiddleware(handleEventsCollection)))
	http.HandleFunc("/api/events/", sec(authMiddleware(handleEventByMethod)))
	http.HandleFunc("/api/session", sec(methodCheck("GET", authMiddleware(handleSession))))
	http.HandleFunc("/api/logout", sec(methodCheck("POST", handleLogout)))
	http.HandleFunc("/api/captcha", sec(methodCheck("GET", rateLimitMiddleware("captcha", 60, handleGetCaptcha))))
	http.HandleFunc("/api/captcha/image/", sec(methodCheck("GET", handleCaptchaImage)))
	http.HandleFunc("/api/auth", sec(rateLimitMiddleware("auth", 15, bodyLimitMiddleware(methodCheck("POST", handleAuth)))))
	http.HandleFunc("/api/health", corsMiddleware(methodCheck("GET", handleHealth)))
	http.HandleFunc("/api/search-demonlist", sec(methodCheck("GET", rateLimitMiddleware("demonlist", 30, handleSearchDemonlist))))
	http.HandleFunc("/api/player-levels", sec(methodCheck("GET", rateLimitMiddleware("demonlist", 30, handlePlayerLevels))))

	fmt.Println("========================================")
	fmt.Println("  SMLT Leaderboard - smlt.lol")
	fmt.Println("========================================")
	fmt.Printf("  Listen:   %s\n", listenAddr)
	fmt.Printf("  Origin:   %s\n", allowedOrigin)
	fmt.Println("  DB:       PostgreSQL")
	fmt.Println("========================================")

	log.Fatal(http.ListenAndServe(listenAddr, securityHeadersMiddleware(http.DefaultServeMux)))
}
