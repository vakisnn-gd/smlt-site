package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const demonlistUsersURL = "https://api.demonlist.org/leaderboard/user/list"

var (
	demonlistHTTPClient = &http.Client{Timeout: 12 * time.Second}
	demonlistSyncMu     sync.Mutex
)

type demonlistUser struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Placement int    `json:"placement"`
	Points    string `json:"points"`
}

type demonlistUserResponse struct {
	Data struct {
		Users []demonlistUser `json:"users"`
	} `json:"data"`
}

type standingUpdate struct {
	PlayerID   int
	GDLID      int64
	Points     float64
	GlobalRank int
}

type sqlExecutor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func startDemonlistSync() {
	time.Sleep(2 * time.Second)
	syncDemonlistPlayers()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		syncDemonlistPlayers()
	}
}

func syncDemonlistPlayers() {
	if !demonlistSyncMu.TryLock() {
		return
	}
	defer demonlistSyncMu.Unlock()

	players, err := dbGetPlayers()
	if err != nil {
		log.Printf("[DEMONLIST SYNC] players: %v", err)
		return
	}

	type result struct {
		update standingUpdate
		err    error
	}
	jobs := make(chan Player)
	results := make(chan result, len(players))
	var workers sync.WaitGroup
	for i := 0; i < 4; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for player := range jobs {
				user, err := fetchDemonlistUser(player.Name)
				if err != nil {
					results <- result{err: err}
					continue
				}
				points, err := strconv.ParseFloat(user.Points, 64)
				if err != nil {
					results <- result{err: err}
					continue
				}
				results <- result{update: standingUpdate{PlayerID: player.ID, GDLID: user.ID, Points: points, GlobalRank: user.Placement}}
			}
		}()
	}
	go func() {
		for _, player := range players {
			jobs <- player
		}
		close(jobs)
		workers.Wait()
		close(results)
	}()

	updates := make([]standingUpdate, 0, len(players))
	failed := 0
	for item := range results {
		if item.err != nil {
			failed++
			continue
		}
		updates = append(updates, item.update)
	}

	changed, err := dbApplyStandingUpdates(updates)
	if err != nil {
		log.Printf("[DEMONLIST SYNC] update: %v", err)
		return
	}
	log.Printf("[DEMONLIST SYNC] matched=%d changed=%d failed=%d", len(updates), changed, failed)
}

func fetchDemonlistUser(name string) (demonlistUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	endpoint := demonlistUsersURL + "?limit=5&search=" + url.QueryEscape(name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return demonlistUser{}, err
	}
	req.Header.Set("User-Agent", "SMLT-Leaderboard/2.0")
	resp, err := demonlistHTTPClient.Do(req)
	if err != nil {
		return demonlistUser{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return demonlistUser{}, fmt.Errorf("demonlist status %d", resp.StatusCode)
	}
	var payload demonlistUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return demonlistUser{}, err
	}
	for _, user := range payload.Data.Users {
		if strings.EqualFold(user.Username, name) {
			return user, nil
		}
	}
	return demonlistUser{}, fmt.Errorf("player %q not found", name)
}

func dbApplyStandingUpdates(updates []standingUpdate) (int, error) {
	if len(updates) == 0 {
		return 0, nil
	}
	before, err := dbGetPlayers()
	if err != nil {
		return 0, err
	}
	byID := make(map[int]Player, len(before))
	for _, player := range before {
		byID[player.ID] = player
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	changedIDs := make([]int, 0)
	changed := 0
	for _, update := range updates {
		old, ok := byID[update.PlayerID]
		if !ok {
			continue
		}
		pointsChanged := math.Abs(old.Points-update.Points) > 0.000001
		standingChanged := pointsChanged || old.GlobalRank != update.GlobalRank
		identityChanged := old.GDLID != update.GDLID
		if !standingChanged && !identityChanged {
			continue
		}
		if _, err := tx.Exec(`UPDATE players SET points=$1, global_rank=$2, demonlist_id=$3, updated_at=NOW() WHERE id=$4`, update.Points, update.GlobalRank, update.GDLID, update.PlayerID); err != nil {
			return 0, err
		}
		if standingChanged {
			if _, err := tx.Exec("INSERT INTO player_history (player_id, points, global_rank) VALUES ($1,$2,$3)", update.PlayerID, update.Points, update.GlobalRank); err != nil {
				return 0, err
			}
			changed++
		}
		if pointsChanged {
			changedIDs = append(changedIDs, update.PlayerID)
		}
	}

	if _, err := tx.Exec(`WITH ranked AS (
		SELECT id, ROW_NUMBER() OVER (ORDER BY points DESC, name ASC) AS rn FROM players
	) UPDATE players SET rank_order = ranked.rn FROM ranked WHERE players.id = ranked.id`); err != nil {
		return 0, err
	}

	after, err := getPlayersTx(tx)
	if err != nil {
		return 0, err
	}
	for _, id := range changedIDs {
		if err := recordRankChange(tx, id, before, after); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return changed, nil
}

func getPlayersTx(tx *sql.Tx) ([]Player, error) {
	rows, err := tx.Query("SELECT id, rank_order, country, name, points, demon, global_rank, COALESCE(demonlist_id, 0) FROM players ORDER BY points DESC, name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	players := []Player{}
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.Rank, &player.Country, &player.Name, &player.Points, &player.Demon, &player.GlobalRank, &player.GDLID); err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	return players, rows.Err()
}

func recordRankChange(executor sqlExecutor, playerID int, before, after []Player) error {
	var oldPlayer, newPlayer Player
	for _, player := range before {
		if player.ID == playerID {
			oldPlayer = player
			break
		}
	}
	for _, player := range after {
		if player.ID == playerID {
			newPlayer = player
			break
		}
	}
	// The feed describes achievements: one row for the player who moved up.
	// Players passively displaced by that move must not create duplicate rows.
	if oldPlayer.Rank == 0 || newPlayer.Rank == 0 || newPlayer.Rank >= oldPlayer.Rank {
		return nil
	}
	above, below := "", ""
	if newPlayer.Rank > 1 {
		above = after[newPlayer.Rank-2].Name
	}
	if newPlayer.Rank < len(after) {
		below = after[newPlayer.Rank].Name
	}
	passed := []string{}
	if newPlayer.Rank < oldPlayer.Rank {
		for rank := newPlayer.Rank + 1; rank <= oldPlayer.Rank && rank <= len(after); rank++ {
			passed = append(passed, after[rank-1].Name)
		}
	} else {
		for rank := oldPlayer.Rank; rank < newPlayer.Rank && rank <= len(after); rank++ {
			passed = append(passed, after[rank-1].Name)
		}
	}
	passedJSON, _ := json.Marshal(passed)
	_, err := executor.Exec(`INSERT INTO rank_changes (player_name, old_rank, new_rank, above_player, below_player, passed_players)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb)`, newPlayer.Name, oldPlayer.Rank, newPlayer.Rank, above, below, string(passedJSON))
	return err
}

func dbGetRecentRankChanges(limit int) ([]RankChange, error) {
	rows, err := db.Query(`SELECT id, player_name, old_rank, new_rank, above_player, below_player, passed_players, created_at
		FROM rank_changes ORDER BY created_at DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	changes := []RankChange{}
	for rows.Next() {
		var change RankChange
		var passedJSON []byte
		if err := rows.Scan(&change.ID, &change.PlayerName, &change.OldRank, &change.NewRank, &change.AbovePlayer, &change.BelowPlayer, &passedJSON, &change.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(passedJSON, &change.PassedPlayers)
		if change.PassedPlayers == nil {
			change.PassedPlayers = []string{}
		}
		changes = append(changes, change)
	}
	return changes, rows.Err()
}
