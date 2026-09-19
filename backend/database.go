package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func resolveHost(host string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	ips, err := r.LookupIPAddr(ctx, host)
	if err != nil || len(ips) == 0 {
		return "", fmt.Errorf("could not resolve %s: %v", host, err)
	}
	for _, ip := range ips {
		if ip.IP.To4() != nil {
			return ip.IP.String(), nil
		}
	}
	return ips[0].IP.String(), nil
}

func initDB() {
	connStr := dbURL

	if u, err := url.Parse(connStr); err == nil && u.Hostname() != "" {
		q := u.Query()
		if q.Get("connect_timeout") == "" {
			q.Set("connect_timeout", "5")
		}
		if q.Get("sslmode") == "" {
			q.Set("sslmode", "require")
		}
		u.RawQuery = q.Encode()

		if resolved, err := resolveHost(u.Hostname()); err == nil {
			log.Printf("[DB] Resolved %s -> %s", u.Hostname(), resolved)
			host := resolved
			if u.Port() != "" {
				host = net.JoinHostPort(resolved, u.Port())
			}
			u.Host = host
			connStr = u.String()
		} else {
			log.Printf("[DB] Resolve warning: %v", err)
			connStr = u.String()
		}
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS players (
		id SERIAL PRIMARY KEY,
		rank_order INTEGER NOT NULL DEFAULT 0,
		country VARCHAR(10) NOT NULL DEFAULT 'RU',
		name VARCHAR(30) NOT NULL UNIQUE,
		points DOUBLE PRECISION NOT NULL DEFAULT 0,
		demon VARCHAR(50) NOT NULL DEFAULT '—',
		global_rank INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_players_points ON players(points DESC);
	CREATE TABLE IF NOT EXISTS player_history (
		id BIGSERIAL PRIMARY KEY,
		player_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
		points DOUBLE PRECISION NOT NULL,
		global_rank INTEGER NOT NULL DEFAULT 0,
		recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_player_history_player_time ON player_history(player_id, recorded_at);
	ALTER TABLE players ADD COLUMN IF NOT EXISTS demonlist_id BIGINT;
	CREATE TABLE IF NOT EXISTS rank_changes (
		id BIGSERIAL PRIMARY KEY,
		player_name VARCHAR(30) NOT NULL,
		old_rank INTEGER NOT NULL,
		new_rank INTEGER NOT NULL,
		above_player VARCHAR(30) NOT NULL DEFAULT '',
		below_player VARCHAR(30) NOT NULL DEFAULT '',
		passed_players JSONB NOT NULL DEFAULT '[]'::jsonb,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_rank_changes_time ON rank_changes(created_at DESC);
	CREATE TABLE IF NOT EXISTS events (
		id SERIAL PRIMARY KEY,
		video_id VARCHAR(32) NOT NULL,
		title VARCHAR(120) NOT NULL,
		category VARCHAR(20) NOT NULL DEFAULT 'project',
		sort_order INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_events_order ON events(sort_order, id);
	`
	if _, err := db.Exec(schema); err != nil {
		log.Fatal("Failed to create schema:", err)
	}

	// Case-insensitive uniqueness so "Flik" and "flik" cannot coexist.
	// Best-effort: fails if the existing table already has such duplicates.
	if _, err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_players_name_lower ON players (lower(name))"); err != nil {
		log.Printf("[DB] WARN could not enforce case-insensitive unique names: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	if count == 0 {
		seedPlayers()
	}
	seedEventsIfEmpty()
	db.Exec(`INSERT INTO player_history (player_id, points, global_rank)
		SELECT p.id, p.points, p.global_rank FROM players p
		WHERE NOT EXISTS (SELECT 1 FROM player_history h WHERE h.player_id = p.id)`)
	log.Printf("[DB] Connected, %d players", count)
}

func dbGetEvents() ([]Event, error) {
	rows, err := db.Query("SELECT id, video_id, title, category, sort_order FROM events ORDER BY sort_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.VideoID, &event.Title, &event.Category, &event.SortOrder); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func dbAddEvent(event Event) error {
	var maxOrder int
	if err := db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) FROM events").Scan(&maxOrder); err != nil {
		return err
	}
	_, err := db.Exec("INSERT INTO events (video_id, title, category, sort_order) VALUES ($1,$2,$3,$4)", event.VideoID, event.Title, event.Category, maxOrder+1)
	return err
}

func dbDeleteEvent(id int) error {
	res, err := db.Exec("DELETE FROM events WHERE id=$1", id)
	if err != nil {
		return err
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		return fmt.Errorf("ивент не найден")
	}
	return nil
}

func dbReorderEvents(ids []int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for order, id := range ids {
		if _, err := tx.Exec("UPDATE events SET sort_order=$1 WHERE id=$2", order, id); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func dbGetPlayers() ([]Player, error) {
	rows, err := db.Query("SELECT id, rank_order, country, name, points, demon, global_rank, COALESCE(demonlist_id, 0) FROM players ORDER BY points DESC, name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.Rank, &p.Country, &p.Name, &p.Points, &p.Demon, &p.GlobalRank, &p.GDLID); err != nil {
			continue
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// refreshRanks recomputes rank_order from the points ordering so the exposed
// rank field stays meaningful after every mutation.
func refreshRanks() {
	db.Exec(`WITH ranked AS (
		SELECT id, ROW_NUMBER() OVER (ORDER BY points DESC, name ASC) AS rn FROM players
	) UPDATE players SET rank_order = ranked.rn FROM ranked WHERE players.id = ranked.id`)
}

func dbAddPlayer(p Player) error {
	err := db.QueryRow(
		"INSERT INTO players (rank_order, country, name, points, demon, global_rank) VALUES (0, $1, $2, $3, $4, $5) RETURNING id",
		p.Country, p.Name, p.Points, p.Demon, p.GlobalRank,
	).Scan(&p.ID)
	if err != nil {
		return err
	}
	refreshRanks()
	db.Exec("INSERT INTO player_history (player_id, points, global_rank) VALUES ($1,$2,$3)", p.ID, p.Points, p.GlobalRank)
	return nil
}

func dbUpdatePlayer(name string, p Player) error {
	before, _ := dbGetPlayers()
	var id int
	err := db.QueryRow(
		"UPDATE players SET country=$1, name=$2, points=$3, demon=$4, global_rank=$5, updated_at=NOW() WHERE lower(name) = lower($6) RETURNING id",
		p.Country, p.Name, p.Points, p.Demon, p.GlobalRank, name,
	).Scan(&id)
	if err != nil {
		return err
	}
	db.Exec("INSERT INTO player_history (player_id, points, global_rank) VALUES ($1,$2,$3)", id, p.Points, p.GlobalRank)
	refreshRanks()
	after, _ := dbGetPlayers()
	recordRankChange(db, id, before, after)
	return nil
}

func dbGetPlayer(name string) (Player, error) {
	var p Player
	err := db.QueryRow("SELECT id, rank_order, country, name, points, demon, global_rank, COALESCE(demonlist_id, 0) FROM players WHERE lower(name)=lower($1)", name).
		Scan(&p.ID, &p.Rank, &p.Country, &p.Name, &p.Points, &p.Demon, &p.GlobalRank, &p.GDLID)
	return p, err
}

func dbGetHistory(name string) ([]HistoryPoint, error) {
	rows, err := db.Query(`SELECT h.points, h.global_rank, h.recorded_at FROM player_history h JOIN players p ON p.id=h.player_id WHERE lower(p.name)=lower($1) ORDER BY h.recorded_at`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	points := []HistoryPoint{}
	for rows.Next() {
		var point HistoryPoint
		if rows.Scan(&point.Points, &point.GlobalRank, &point.RecordedAt) == nil {
			points = append(points, point)
		}
	}
	return points, rows.Err()
}

func dbDeletePlayer(name string) error {
	res, err := db.Exec("DELETE FROM players WHERE lower(name) = lower($1)", name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("игрок не найден")
	}
	refreshRanks()
	return nil
}

func dbPing() error {
	return db.Ping()
}
