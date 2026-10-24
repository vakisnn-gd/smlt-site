package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type trackedDemon struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Placement int    `json:"placement"`
}
type trackedLevels struct {
	Hardest   trackedDemon   `json:"hardest"`
	Main      []trackedDemon `json:"main"`
	Extended  []trackedDemon `json:"extended"`
	Advanced  []trackedDemon `json:"advanced"`
	Unbounded []trackedDemon `json:"unbounded"`
}

func fetchPlayerDemons(gdlID int64) ([]trackedDemon, trackedDemon, error) {
	resp, err := demonlistHTTPClient.Get(demonlistApiBase + "/user/get?id=" + url.QueryEscape(strconv.FormatInt(gdlID, 10)))
	if err != nil {
		return nil, trackedDemon{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, trackedDemon{}, fmt.Errorf("demonlist status %d", resp.StatusCode)
	}
	var payload struct {
		Data struct {
			Levels trackedLevels `json:"levels"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, trackedDemon{}, err
	}
	all := append([]trackedDemon{}, payload.Data.Levels.Main...)
	all = append(all, payload.Data.Levels.Extended...)
	all = append(all, payload.Data.Levels.Advanced...)
	all = append(all, payload.Data.Levels.Unbounded...)
	return all, payload.Data.Levels.Hardest, nil
}

func syncPlayerDemons(players []Player) {
	for _, player := range players {
		if player.GDLID == 0 {
			continue
		}
		demons, hardest, err := fetchPlayerDemons(player.GDLID)
		if err != nil {
			log.Printf("[DEMONLIST SYNC] hardest for %q: %v", player.Name, err)
			continue
		}
		if _, err := db.Exec("UPDATE players SET demon=$1, demon_placement=$2 WHERE id=$3", hardest.Name, hardest.Placement, player.ID); err != nil {
			log.Printf("[DEMONLIST SYNC] hardest update for %q: %v", player.Name, err)
		}
		var initialized bool
		if err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM player_demon_sync_state WHERE player_id=$1)", player.ID).Scan(&initialized); err != nil {
			log.Printf("[DEMONLIST SYNC] baseline state for %q: %v", player.Name, err)
			continue
		}
		baseline := !initialized
		syncComplete := true
		for _, demon := range demons {
			var exists bool
			if err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM player_demons WHERE player_id=$1 AND demon_id=$2)", player.ID, demon.ID).Scan(&exists); err != nil {
				syncComplete = false
				continue
			}
			if exists {
				continue
			}
			if _, err := db.Exec("INSERT INTO player_demons (player_id,demon_id,demon_name) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING", player.ID, demon.ID, demon.Name); err != nil {
				syncComplete = false
				continue
			}
			if !baseline {
				_, _ = db.Exec("INSERT INTO player_events (event_type, player_name, country, detail) VALUES ('demon',$1,$2,$3)", player.Name, player.Country, demon.Name)
				if discordBot != nil {
					notifyPlayerEvent("demon", fmt.Sprintf("%s **%s** прошёл новый демон: **%s**", countryFlag(player.Country), player.Name, demon.Name))
				}
			}
		}
		if syncComplete {
			if _, err := db.Exec("INSERT INTO player_demon_sync_state (player_id) VALUES ($1) ON CONFLICT (player_id) DO NOTHING", player.ID); err != nil {
				log.Printf("[DEMONLIST SYNC] baseline marker for %q: %v", player.Name, err)
			}
		}
	}
}
