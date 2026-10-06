package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const demonlistApiBase = "https://api.demonlist.org"

func demonlistCountryCode(country string) string {
	switch strings.ToLower(strings.TrimSpace(country)) {
	case "russia":
		return "RU"
	case "ukraine":
		return "UA"
	case "serbia":
		return "RS"
	case "belarus":
		return "BY"
	case "bulgaria":
		return "BG"
	case "germany":
		return "DE"
	case "armenia":
		return "AM"
	case "kazakhstan":
		return "KZ"
	default:
		return "OTHER"
	}
}

var demonlistClient = &http.Client{Timeout: 10 * time.Second}
var demonlistCacheMu sync.Mutex
var demonlistCache = map[string]cachedDemonlistResponse{}
var searchRateMu sync.Mutex
var searchRate = map[string]time.Time{}

type cachedDemonlistResponse struct {
	body    []byte
	expires time.Time
}

const demonlistCacheTTL = 5 * time.Minute

func cachedDemonlistGet(key string) ([]byte, bool) {
	demonlistCacheMu.Lock()
	defer demonlistCacheMu.Unlock()
	item, ok := demonlistCache[key]
	if !ok || time.Now().After(item.expires) {
		return nil, false
	}
	return append([]byte(nil), item.body...), true
}
func cachedDemonlistPut(key string, body []byte) {
	demonlistCacheMu.Lock()
	defer demonlistCacheMu.Unlock()
	demonlistCache[key] = cachedDemonlistResponse{body: append([]byte(nil), body...), expires: time.Now().Add(demonlistCacheTTL)}
}

type demonlistDetail struct {
	Data struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Banned      bool    `json:"banned"`
		Score       float64 `json:"score"`
		Rank        int     `json:"rank"`
		Nationality *struct {
			CountryCode string `json:"country_code"`
			Nation      string `json:"nation"`
		} `json:"nationality"`
		Records []struct {
			Demon struct {
				ID       int    `json:"id"`
				Name     string `json:"name"`
				Position int    `json:"position"`
			} `json:"demon"`
			Progress int    `json:"progress"`
			Status   string `json:"status"`
		} `json:"records"`
		Verified []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Position int    `json:"position"`
		} `json:"verified"`
		Created []struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Position int    `json:"position"`
		} `json:"created"`
	} `json:"data"`
}

func handleSearchDemonlist(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")
	if i := strings.IndexByte(ip, ','); i >= 0 {
		ip = ip[:i]
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	searchRateMu.Lock()
	last := searchRate[ip]
	if time.Since(last) < time.Second {
		searchRateMu.Unlock()
		writeDemonlistError(w, http.StatusTooManyRequests, "Слишком много запросов")
		return
	}
	searchRate[ip] = time.Now()
	searchRateMu.Unlock()
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Введите никнейм"})
		return
	}

	req, err := http.NewRequest("GET", demonlistUsersURL+"?limit=5&search="+url.QueryEscape(name), nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка запроса"})
		return
	}
	req.Header.Set("User-Agent", "SMLT-Leaderboard/1.0")
	body, ok := cachedDemonlistGet("search:" + strings.ToLower(name))
	if !ok {
		resp, err := demonlistClient.Do(req)
		if err != nil {
			writeDemonlistError(w, http.StatusBadGateway, "Не удалось подключиться к demonlist.org")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			writeDemonlistError(w, http.StatusBadGateway, "Demonlist API недоступен")
			return
		}
		body, err = io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			writeDemonlistError(w, http.StatusBadGateway, "Ошибка чтения ответа")
			return
		}
		cachedDemonlistPut("search:"+strings.ToLower(name), body)
	}
	if len(body) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Пустой ответ demonlist.org"})
		return
	}

	var search demonlistUserResponse
	if err := json.Unmarshal(body, &search); err != nil || len(search.Data.Users) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Игрок не найден на demonlist.org"})
		return
	}

	found := search.Data.Users[0]
	for _, user := range search.Data.Users {
		if strings.EqualFold(user.Username, name) {
			found = user
			break
		}
	}
	points, _ := strconv.ParseFloat(found.Points, 64)

	detailReq, err := http.NewRequest("GET", demonlistApiBase+"/user/get?id="+strconv.FormatInt(found.ID, 10), nil)
	if err != nil {
		writeDemonlistError(w, http.StatusInternalServerError, "Ошибка запроса")
		return
	}
	detailReq.Header.Set("User-Agent", "SMLT-Leaderboard/1.0")

	detailResp, err := demonlistClient.Do(detailReq)
	if err != nil {
		writeDemonlistError(w, http.StatusBadGateway, "Не удалось подключиться к demonlist.org")
		return
	}
	defer detailResp.Body.Close()

	if detailResp.StatusCode != 200 {
		writeDemonlistError(w, http.StatusBadGateway, "Некорректный ответ demonlist.org")
		return
	}

	detailBody, err := io.ReadAll(io.LimitReader(detailResp.Body, 4<<20))
	if err != nil {
		writeDemonlistError(w, http.StatusBadGateway, "Ошибка чтения ответа")
		return
	}

	var detail demonlistDetail
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		writeDemonlistError(w, http.StatusBadGateway, "Некорректный ответ demonlist.org")
		return
	}

	hardestDemon := ""
	completed := make([]struct {
		Name     string
		Position int
	}, 0)
	for _, rec := range detail.Data.Records {
		if rec.Progress == 100 && rec.Status == "approved" {
			completed = append(completed, struct {
				Name     string
				Position int
			}{Name: rec.Demon.Name, Position: rec.Demon.Position})
		}
	}
	for _, v := range detail.Data.Verified {
		completed = append(completed, struct {
			Name     string
			Position int
		}{Name: v.Name, Position: v.Position})
	}
	for _, c := range detail.Data.Created {
		completed = append(completed, struct {
			Name     string
			Position int
		}{Name: c.Name, Position: c.Position})
	}

	if len(completed) > 0 {
		sort.Slice(completed, func(i, j int) bool {
			return completed[i].Position < completed[j].Position
		})
		hardestDemon = completed[0].Name
	}

	country := demonlistCountryCode(found.Country)

	log.Printf("[DEMONLIST] Found: %s (rank #%d, score %.2f)", found.Username, found.Placement, points)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"name":       found.Username,
		"country":    country,
		"points":     points,
		"demon":      hardestDemon,
		"globalRank": found.Placement,
	})
}

func writeDemonlistError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": message})
}

type demonlistLevel struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Placement int    `json:"placement"`
	VideoURL  string `json:"video_url,omitempty"`
}

type demonlistLevels struct {
	Hardest   *demonlistLevel  `json:"hardest"`
	Main      []demonlistLevel `json:"main"`
	Extended  []demonlistLevel `json:"extended"`
	Advanced  []demonlistLevel `json:"advanced"`
	Unbounded []demonlistLevel `json:"unbounded"`
	Progress  []demonlistLevel `json:"progress"`
	Verified  []demonlistLevel `json:"verified"`
}

func handlePlayerLevels(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	name := strings.TrimSpace(r.URL.Query().Get("name"))

	if id == "" && name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Не указан игрок"})
		return
	}

	if id == "" {
		searchReq, err := http.NewRequest("GET", demonlistUsersURL+"?limit=5&search="+url.QueryEscape(name), nil)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка запроса"})
			return
		}
		searchReq.Header.Set("User-Agent", "SMLT-Leaderboard/1.0")
		searchResp, err := demonlistClient.Do(searchReq)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Не удалось подключиться к demonlist.org"})
			return
		}
		searchBody, err := io.ReadAll(io.LimitReader(searchResp.Body, 2<<20))
		searchResp.Body.Close()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка чтения ответа"})
			return
		}
		var search demonlistUserResponse
		if err := json.Unmarshal(searchBody, &search); err != nil || len(search.Data.Users) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Игрок не найден на demonlist.org"})
			return
		}
		found := search.Data.Users[0]
		for _, user := range search.Data.Users {
			if strings.EqualFold(user.Username, name) {
				found = user
				break
			}
		}
		id = strconv.FormatInt(found.ID, 10)
	}
	cacheKey := "levels:" + id
	if cached, ok := cachedDemonlistGet(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, max-age=300")
		_, _ = w.Write(cached)
		return
	}

	req, err := http.NewRequest("GET", demonlistApiBase+"/user/get?id="+url.QueryEscape(id), nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка запроса"})
		return
	}
	req.Header.Set("User-Agent", "SMLT-Leaderboard/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := demonlistClient.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Не удалось подключиться к demonlist.org"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Игрок не найден на demonlist.org"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка чтения ответа"})
		return
	}

	var parsed struct {
		Data struct {
			Username  string          `json:"username"`
			Placement int             `json:"placement"`
			Levels    demonlistLevels `json:"levels"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Некорректный ответ demonlist.org"})
		return
	}

	log.Printf("[DEMONLIST] Levels for %s (place #%d)", parsed.Data.Username, parsed.Data.Placement)

	result := map[string]interface{}{
		"success":   true,
		"id":        id,
		"username":  parsed.Data.Username,
		"placement": parsed.Data.Placement,
		"levels":    parsed.Data.Levels,
	}
	encoded, _ := json.Marshal(result)
	cachedDemonlistPut(cacheKey, encoded)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = w.Write(encoded)
}
