package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func handleEventsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		events, err := dbGetEvents()
		if err != nil {
			http.Error(w, `{"success":false,"message":"Ошибка базы данных"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "events": events})
	case "POST":
		authMiddleware(handleAddEvent)(w, r)
	case "PUT":
		authMiddleware(handleReorderEvents)(w, r)
	default:
		w.Header().Set("Allow", "GET, POST, PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleEventByMethod(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.Header().Set("Allow", "DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var id int
	if _, err := fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/api/events/"), "%d", &id); err != nil || id < 1 {
		http.Error(w, `{"success":false,"message":"Неверный идентификатор"}`, http.StatusBadRequest)
		return
	}
	if err := dbDeleteEvent(id); err != nil {
		http.Error(w, `{"success":false,"message":"Ивент не найден"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleAddEvent(w http.ResponseWriter, r *http.Request) {
	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Неверный запрос"}`, http.StatusBadRequest)
		return
	}
	req.VideoID, req.Title, req.Category = sanitizeInput(req.VideoID), sanitizeInput(req.Title), sanitizeInput(req.Category)
	videoID, validVideo := normalizeYouTubeID(req.VideoID)
	if !validVideo || len(req.Title) < 1 || len(req.Title) > 120 || (req.Category != "beat" && req.Category != "project") {
		http.Error(w, `{"success":false,"message":"Вставьте корректную ссылку YouTube или 11-символьный ID, название и категорию"}`, http.StatusBadRequest)
		return
	}
	req.VideoID = videoID
	if err := dbAddEvent(Event{VideoID: req.VideoID, Title: req.Title, Category: req.Category}); err != nil {
		http.Error(w, `{"success":false,"message":"Ошибка базы данных"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleReorderEvents(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 || len(req.IDs) > 200 {
		http.Error(w, `{"success":false,"message":"Неверный порядок"}`, http.StatusBadRequest)
		return
	}
	if err := dbReorderEvents(req.IDs); err != nil {
		http.Error(w, `{"success":false,"message":"Ошибка сортировки"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handlePlayerByMethod(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		handleUpdatePlayer(w, r)
	case "DELETE":
		handleDeletePlayer(w, r)
	default:
		w.Header().Set("Allow", "PUT, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Метод не поддерживается"})
	}
}

func handlePlayersCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleGetPlayers(w, r)
	case "POST":
		authMiddleware(handleAddPlayer)(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Метод не поддерживается"})
	}
}

func handleGetPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := dbGetPlayers()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка базы данных"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "players": players})
}

func handleRecentChanges(w http.ResponseWriter, r *http.Request) {
	changes, err := dbGetRecentRankChanges(40)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка базы данных"})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "changes": changes})
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "authenticated": true})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "smlt_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleGetCaptcha(w http.ResponseWriter, r *http.Request) {
	captcha := createCaptcha()
	nonce := generateNonce()

	mu.Lock()
	nonces[nonce] = time.Now()
	mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "smlt_nonce",
		Value:    nonce,
		Path:     "/api",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   300,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"captchaId": captcha.ID,
		"imageUrl":  "/api/captcha/image/" + captcha.ID,
	})
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Неверный запрос"})
		return
	}

	req.Password = sanitizeInput(req.Password)
	req.CaptchaAnswer = sanitizeInput(req.CaptchaAnswer)
	req.CaptchaID = sanitizeInput(req.CaptchaID)

	nonceCookie, err := r.Cookie("smlt_nonce")
	if err != nil || nonceCookie.Value == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Капча истекла"})
		return
	}
	nonceVal := sanitizeInput(nonceCookie.Value)

	// Consume the captcha BEFORE touching the password so every password guess
	// costs the attacker a freshly solved captcha (anti-brute-force).
	captcha, exists := consumeCaptcha(req.CaptchaID, nonceVal)
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Капча истекла или nonce использован"})
		return
	}

	if subtle.ConstantTimeCompare([]byte(strings.ToLower(req.CaptchaAnswer)), []byte(strings.ToLower(captcha.Answer))) != 1 {
		log.Printf("[AUTH] FAIL captcha from %s", clientIP(r))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Неверный ответ на капчу"})
		return
	}

	if len(req.Password) < 1 || len(req.Password) > 256 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Неверный пароль"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(req.Password)); err != nil {
		log.Printf("[AUTH] FAIL login from %s", clientIP(r))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Неверный пароль"})
		return
	}

	token, err := generateToken()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка сервера"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "smlt_nonce",
		Value:    "",
		Path:     "/api",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "smlt_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	log.Printf("[AUTH] OK login from %s", clientIP(r))
	w.Header().Set("Content-Type", "application/json")
	// The bundled admin UI keeps this token only in memory. Authentication also
	// uses the HttpOnly session cookie, so the token is no longer persisted in
	// localStorage but the existing UI can still finish its login flow.
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "token": token})
}

func handleAddPlayer(w http.ResponseWriter, r *http.Request) {
	var req PlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Неверный запрос"})
		return
	}

	req.Name = sanitizeInput(req.Name)
	req.Demon = sanitizeInput(req.Demon)

	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Имя обязательно"})
		return
	}
	if len(req.Name) > 30 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Имя макс. 30 символов"})
		return
	}
	if len(req.Demon) > 50 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Название демона макс. 50 символов"})
		return
	}
	if !isValidCountry(req.Country) {
		req.Country = "OTHER"
	}
	if req.Points < 0 || req.Points > 999999 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Очки от 0 до 999999"})
		return
	}
	if req.GlobalRank < 0 || req.GlobalRank > 999999 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Мировой рейтинг от 0 до 999999"})
		return
	}

	if err := dbAddPlayer(Player{Country: req.Country, Name: req.Name, Points: req.Points, Demon: req.Demon, GlobalRank: req.GlobalRank}); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Игрок уже существует или ошибка БД"})
		return
	}

	log.Printf("[PLAYER] ADD %s by %s", req.Name, clientIP(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Игрок добавлен"})
}

func handleUpdatePlayer(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/players/")
	name, _ = url.PathUnescape(name)
	name = sanitizeInput(name)

	var req PlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Неверный запрос"})
		return
	}

	req.Name = sanitizeInput(req.Name)
	req.Demon = sanitizeInput(req.Demon)

	if len(req.Name) > 30 || len(req.Demon) > 50 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Превышена длина полей"})
		return
	}
	if !isValidCountry(req.Country) {
		req.Country = "OTHER"
	}
	if req.Points < 0 || req.Points > 999999 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Очки от 0 до 999999"})
		return
	}
	if req.GlobalRank < 0 || req.GlobalRank > 999999 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Мировой рейтинг от 0 до 999999"})
		return
	}

	if err := dbUpdatePlayer(name, Player{Country: req.Country, Name: req.Name, Points: req.Points, Demon: req.Demon, GlobalRank: req.GlobalRank}); err != nil {
		log.Printf("[PLAYER] UPDATE FAIL %s: %v", name, err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка обновления игрока"})
		return
	}

	log.Printf("[PLAYER] UPDATE %s -> %s by %s", name, req.Name, clientIP(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Игрок обновлён"})
}

func handleDeletePlayer(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/players/")
	name, _ = url.PathUnescape(name)
	name = sanitizeInput(name)

	if err := dbDeletePlayer(name); err != nil {
		log.Printf("[PLAYER] DELETE FAIL %s: %v", name, err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка удаления игрока"})
		return
	}

	log.Printf("[PLAYER] DELETE %s by %s", name, clientIP(r))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Игрок удалён"})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	err := dbPing()
	status := "ok"
	if err != nil {
		status = "db_error"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": status})
}
