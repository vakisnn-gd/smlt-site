package main

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type pageMeta struct {
	title       string
	description string
	canonical   string
}

func serveAppPage(w http.ResponseWriter, r *http.Request, frontendDir string) bool {
	requestPath := r.URL.Path
	pagePath := requestPath
	isEnglish := pagePath == "/en" || strings.HasPrefix(pagePath, "/en/")
	if isEnglish {
		pagePath = strings.TrimPrefix(pagePath, "/en")
		if pagePath == "" {
			pagePath = "/"
		}
	}
	meta := pageMeta{
		title:       "SMLT Leaderboard",
		description: "Рейтинг и достижения игроков сообщества SMLT.",
		canonical:   "https://smlt.lol" + requestPath,
	}
	content := renderLeaderboardFallback(isEnglish)
	if isEnglish {
		meta.description = "SMLT community leaderboard and player achievements."
	}

	switch pagePath {
	case "/":
	case "/events":
		meta = pageMeta{"SMLT — ивенты", "Ивенты и коллабы сообщества SMLT.", "https://smlt.lol" + requestPath}
		content = renderEventsFallback(false)
		if isEnglish {
			meta.title, meta.description, content = "SMLT Events", "SMLT events and collaborations.", renderEventsFallback(true)
		}
	case "/about":
		meta = pageMeta{"О SMLT", "Информация о сообществе SMLT, контакты и проекты.", "https://smlt.lol" + requestPath}
		content = `<main class="page community-page"><h1>Что такое SMLT?</h1><p>SMLT — это дискорд-сервер, где происходят разные ивенты: коллабы, турниры, прохождения уровней по частям и даже игры в Майнкрафте.</p><p>Все участники сервера — участники SMLT, независимо от их скилла.</p></main>`
		if isEnglish {
			meta.title, meta.description = "About SMLT", "Information about the SMLT community, contacts and projects."
			content = `<main class="page community-page"><h1>What is SMLT?</h1><p>SMLT is a Discord server where we hold events: collaborations, tournaments, level part runs, and even Minecraft games.</p><p>Every member of the server is a member of SMLT, regardless of their skill.</p></main>`
		}
	default:
		if !strings.HasPrefix(pagePath, "/player/") {
			return false
		}
		publicID, err := strconv.ParseInt(strings.TrimPrefix(pagePath, "/player/"), 10, 64)
		if err != nil || publicID < 1 {
			return false
		}
		player, ok := findPublicPlayer(publicID)
		if !ok {
			return false
		}
		meta = pageMeta{player.Name + " — SMLT", "Рейтинг и достижения игроков сообщества SMLT.", "https://smlt.lol" + requestPath}
		if isEnglish {
			meta.description = "SMLT community leaderboard and player achievements."
		}
		content = fmt.Sprintf(`<main class="page"><h1>%s</h1><p>#%d SMLT · %.2f · #%d</p><p>%s</p></main>`, html.EscapeString(player.Name), player.Rank, player.Points, player.GlobalRank, html.EscapeString(player.Demon))
	}

	index, err := os.ReadFile(filepath.Join(frontendDir, "index.html"))
	if err != nil {
		http.Error(w, "Site is temporarily unavailable", http.StatusServiceUnavailable)
		return true
	}
	page := string(index)
	if isEnglish {
		page = strings.Replace(page, `<html lang="ru">`, `<html lang="en">`, 1)
	}
	page = replaceMeta(page, `<title>`, `</title>`, html.EscapeString(meta.title))
	page = replaceAttribute(page, `meta name="description" content="`, html.EscapeString(meta.description))
	page = replaceAttribute(page, `meta property="og:title" content="`, html.EscapeString(meta.title))
	page = replaceAttribute(page, `meta property="og:description" content="`, html.EscapeString(meta.description))
	page = replaceAttribute(page, `meta property="og:url" content="`, html.EscapeString(meta.canonical))
	page = replaceAttribute(page, `link rel="canonical" href="`, html.EscapeString(meta.canonical))
	ruPath := pagePath
	enPath := "/en" + pagePath
	if pagePath == "/" {
		enPath = "/en"
	}
	page = replaceAttribute(page, `link rel="alternate" hreflang="ru" href="`, "https://smlt.lol"+ruPath)
	page = replaceAttribute(page, `link rel="alternate" hreflang="en" href="`, "https://smlt.lol"+enPath)
	page = replaceAttribute(page, `link rel="alternate" hreflang="x-default" href="`, "https://smlt.lol"+ruPath)
	page = strings.Replace(page, `<div id="app"></div>`, `<div id="app">`+content+`</div>`, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(page))
	return true
}

func replaceMeta(page, start, end, value string) string {
	left := strings.Index(page, start)
	if left < 0 {
		return page
	}
	right := strings.Index(page[left+len(start):], end)
	if right < 0 {
		return page
	}
	right += left + len(start)
	return page[:left+len(start)] + value + page[right:]
}

func replaceAttribute(page, marker, value string) string {
	start := strings.Index(page, marker)
	if start < 0 {
		return page
	}
	start += len(marker)
	end := strings.IndexByte(page[start:], '"')
	if end < 0 {
		return page
	}
	return page[:start] + value + page[start+end:]
}

func findPublicPlayer(publicID int64) (Player, bool) {
	players, err := dbGetPlayers()
	if err != nil {
		return Player{}, false
	}
	for _, player := range players {
		id := player.GDLID
		if id == 0 {
			id = int64(player.ID)
		}
		if id == publicID {
			return player, true
		}
	}
	return Player{}, false
}

func renderLeaderboardFallback(english bool) string {
	players, err := dbGetPlayers()
	if err != nil {
		return `<main class="page"><h1>SMLT Leaderboard</h1></main>`
	}
	var out strings.Builder
	out.WriteString(`<main class="page leaderboard-page"><h1>SMLT Leaderboard</h1><ol>`)
	prefix := ""
	if english {
		prefix = "/en"
	}
	for _, player := range players {
		id := player.GDLID
		if id == 0 {
			id = int64(player.ID)
		}
		fmt.Fprintf(&out, `<li><a href="%s/player/%d">%s</a> — %.2f</li>`, prefix, id, html.EscapeString(player.Name), player.Points)
	}
	out.WriteString(`</ol></main>`)
	return out.String()
}

func renderEventsFallback(english bool) string {
	events, err := dbGetEvents()
	title := "Ивенты и коллабы"
	if english {
		title = "Events and collabs"
	}
	if err != nil {
		return `<main class="page"><h1>` + title + `</h1></main>`
	}
	var out strings.Builder
	out.WriteString(`<main class="page events-page"><h1>` + title + `</h1><ul>`)
	for _, event := range events {
		fmt.Fprintf(&out, `<li>%s</li>`, html.EscapeString(event.Title))
	}
	out.WriteString(`</ul></main>`)
	return out.String()
}

func isAppPath(requestPath string) bool {
	pagePath := requestPath
	if pagePath == "/en" || strings.HasPrefix(pagePath, "/en/") {
		pagePath = strings.TrimPrefix(pagePath, "/en")
		if pagePath == "" {
			pagePath = "/"
		}
	}
	return pagePath == "/" || pagePath == "/events" || pagePath == "/about" || strings.HasPrefix(pagePath, "/player/")
}
