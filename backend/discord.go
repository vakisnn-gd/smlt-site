package main

// Discord notifications are deliberately implemented with the Discord HTTP API
// so the website does not need a second long-running gateway process.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func countryFlag(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 2 {
		return "🌐"
	}
	return string(rune(code[0]-'A')+0x1F1E6) + string(rune(code[1]-'A')+0x1F1E6)
}

type discordNotifier struct {
	token   string
	channel string
}

var discordOnce sync.Once
var discordBot *discordNotifier
var discordSent sync.Map
var discordStartedAt = time.Now()

func initDiscordNotifier() {
	discordOnce.Do(func() {
		discordBot = newDiscordNotifier()
		if discordBot == nil {
			log.Printf("[DISCORD] notifications disabled: missing DISCORD_BOT_TOKEN or DISCORD_SLAYERS_CHANNEL_ID")
			return
		}
		log.Printf("[DISCORD] notifications enabled for channel %s", discordBot.channel)
	})
}

func notifyPlayerEvent(eventKey, text string) {
	if discordBot == nil {
		return
	}
	key := eventKey + "\x00" + text
	if _, loaded := discordSent.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	if err := discordBot.send(text); err != nil {
		discordSent.Delete(key)
		log.Printf("[DISCORD] notification: %v", err)
	}
}

func newDiscordNotifier() *discordNotifier {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	channel := os.Getenv("DISCORD_SLAYERS_CHANNEL_ID")
	if token == "" || channel == "" {
		return nil
	}
	return &discordNotifier{token: token, channel: channel}
}

func (d *discordNotifier) send(content string) error {
	body, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "https://discord.com/api/v10/channels/"+d.channel+"/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bot "+d.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord status %d", resp.StatusCode)
	}
	return nil
}

func notifyRankChanges(changes []RankChange) {
	if discordBot == nil || len(changes) == 0 {
		return
	}
	for _, c := range changes {
		if c.CreatedAt.Before(discordStartedAt) {
			continue
		}
		if c.NewRank >= c.OldRank {
			continue
		}
		msg := fmt.Sprintf("%s **%s** поднялся с #%d на #%d", countryFlag(c.Country), c.PlayerName, c.OldRank, c.NewRank)
		if len(c.PassedPlayers) > 0 {
			passed := make([]string, 0, len(c.PassedPlayers))
			for _, name := range c.PassedPlayers {
				passed = append(passed, "**"+name+"**")
			}
			msg += ", обойдя " + strings.Join(passed, ", ")
		}
		key := "rank\x00" + msg
		if _, loaded := discordSent.LoadOrStore(key, struct{}{}); loaded {
			continue
		}
		if err := discordBot.send(msg); err != nil {
			discordSent.Delete(key)
			log.Printf("[DISCORD] rank update: %v", err)
		}
	}
}
