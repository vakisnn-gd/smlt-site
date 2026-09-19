package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const demonlistBase = "https://pointercrate.com/api/v1"

var demonlistClient = &http.Client{Timeout: 10 * time.Second}

type demonlistRanking struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Banned      bool    `json:"banned"`
	Score       float64 `json:"score"`
	Rank        int     `json:"rank"`
	Nationality *struct {
		CountryCode string `json:"country_code"`
		Nation      string `json:"nation"`
	} `json:"nationality"`
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
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Введите никнейм"})
		return
	}

	req, err := http.NewRequest("GET", demonlistBase+"/players/ranking/?name_contains="+url.QueryEscape(name)+"&limit=5", nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка запроса"})
		return
	}
	req.Header.Set("User-Agent", "SMLT-Leaderboard/1.0")

	resp, err := demonlistClient.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Не удалось подключиться к demonlist.org"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Игрок не найден на demonlist.org"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Ошибка чтения ответа"})
		return
	}

	var rankings []demonlistRanking
	if err := json.Unmarshal(body, &rankings); err != nil || len(rankings) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Игрок не найден на demonlist.org"})
		return
	}

	best := rankings[0]
	for _, p := range rankings {
		if p.Banned {
			continue
		}
		if strings.EqualFold(p.Name, name) {
			best = p
			break
		}
	}
	if best.Banned {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Игрок забанен на demonlist.org"})
		return
	}

	detailReq, err := http.NewRequest("GET", fmt.Sprintf("%s/players/%d/", demonlistBase, best.ID), nil)
	if err != nil {
		sendDemonlistResult(w, best, "")
		return
	}
	detailReq.Header.Set("User-Agent", "SMLT-Leaderboard/1.0")

	detailResp, err := demonlistClient.Do(detailReq)
	if err != nil {
		sendDemonlistResult(w, best, "")
		return
	}
	defer detailResp.Body.Close()

	if detailResp.StatusCode != 200 {
		sendDemonlistResult(w, best, "")
		return
	}

	detailBody, err := io.ReadAll(io.LimitReader(detailResp.Body, 4<<20))
	if err != nil {
		sendDemonlistResult(w, best, "")
		return
	}

	var detail demonlistDetail
	if err := json.Unmarshal(detailBody, &detail); err != nil {
		sendDemonlistResult(w, best, "")
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

	country := ""
	if best.Nationality != nil {
		country = best.Nationality.CountryCode
	}

	log.Printf("[DEMONLIST] Found: %s (rank #%d, score %.0f)", best.Name, best.Rank, best.Score)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"name":       best.Name,
		"country":    country,
		"points":     best.Score,
		"demon":      hardestDemon,
		"globalRank": best.Rank,
	})
}

func sendDemonlistResult(w http.ResponseWriter, p demonlistRanking, demon string) {
	country := ""
	if p.Nationality != nil {
		country = p.Nationality.CountryCode
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"name":       p.Name,
		"country":    country,
		"points":     p.Score,
		"demon":      demon,
		"globalRank": p.Rank,
	})
}
