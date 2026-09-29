package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type trackedDemon struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type trackedLevels struct {
	Main      []trackedDemon `json:"main"`
	Extended  []trackedDemon `json:"extended"`
	Advanced  []trackedDemon `json:"advanced"`
	Unbounded []trackedDemon `json:"unbounded"`
}

func fetchPlayerDemons(gdlID int64) ([]trackedDemon, error) {
	resp, err := demonlistHTTPClient.Get(demonlistApiBase + "/user/get?id=" + url.QueryEscape(strconv.FormatInt(gdlID, 10)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("demonlist status %d", resp.StatusCode)
	}
	var payload struct {
		Data struct {
			Levels trackedLevels `json:"levels"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	all := append([]trackedDemon{}, payload.Data.Levels.Main...)
	all = append(all, payload.Data.Levels.Extended...)
	all = append(all, payload.Data.Levels.Advanced...)
	all = append(all, payload.Data.Levels.Unbounded...)
	return all, nil
}

func syncPlayerDemons(players []Player) {
	for _, player := range players {
		if player.GDLID == 0 {
			continue
		}
		demons, err := fetchPlayerDemons(player.GDLID)
		if err != nil {
			continue
		}
		var known int
		_ = db.QueryRow("SELECT COUNT(*) FROM player_demons WHERE player_id=$1", player.ID).Scan(&known)
		baseline := known == 0
		for _, demon := range demons {
			var exists bool
			if err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM player_demons WHERE player_id=$1 AND demon_id=$2)", player.ID, demon.ID).Scan(&exists); err != nil || exists {
				continue
			}
			if _, err := db.Exec("INSERT INTO player_demons (player_id,demon_id,demon_name) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING", player.ID, demon.ID, demon.Name); err != nil {
				continue
			}
			_, _ = db.Exec("INSERT INTO player_events (event_type, player_name, country, detail) VALUES ('demon',$1,$2,$3)", player.Name, player.Country, demon.Name)
			if !baseline && discordBot != nil {
				notifyPlayerEvent("demon", fmt.Sprintf("%s **%s** прошёл новый демон: **%s**", countryFlag(player.Country), player.Name, demon.Name))
			}
		}
	}
}
