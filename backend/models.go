package main

import "time"

type Player struct {
	ID         int     `json:"id"`
	Rank       int     `json:"rank"`
	Country    string  `json:"country"`
	Name       string  `json:"name"`
	Points     float64 `json:"points"`
	Demon      string  `json:"demon"`
	GlobalRank int     `json:"globalRank"`
	GDLID      int64   `json:"gdlId,omitempty"`
}

type RankChange struct {
	ID            int64     `json:"id"`
	PlayerName    string    `json:"playerName"`
	OldRank       int       `json:"oldRank"`
	NewRank       int       `json:"newRank"`
	AbovePlayer   string    `json:"abovePlayer,omitempty"`
	BelowPlayer   string    `json:"belowPlayer,omitempty"`
	PassedPlayers []string  `json:"passedPlayers"`
	CreatedAt     time.Time `json:"createdAt"`
}

type HistoryPoint struct {
	Points     float64   `json:"points"`
	GlobalRank int       `json:"globalRank"`
	RecordedAt time.Time `json:"recordedAt"`
}

type Event struct {
	ID        int    `json:"id"`
	VideoID   string `json:"videoId"`
	Title     string `json:"title"`
	Category  string `json:"category"`
	SortOrder int    `json:"sortOrder"`
}

type EventRequest struct {
	VideoID  string `json:"videoId"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

type Captcha struct {
	ID        string
	Answer    string
	CreatedAt time.Time
}

type AuthRequest struct {
	Password      string `json:"password"`
	CaptchaID     string `json:"captchaId"`
	CaptchaAnswer string `json:"captchaAnswer"`
}

type PlayerRequest struct {
	Name       string  `json:"name"`
	Country    string  `json:"country"`
	Points     float64 `json:"points"`
	Demon      string  `json:"demon"`
	GlobalRank int     `json:"globalRank"`
}

type RateLimit struct {
	Attempts    int
	LastTry     time.Time
	LockedUntil time.Time
}
