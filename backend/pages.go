package main

import (
	"encoding/json"
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
		content = fmt.Sprintf(`<main class="page"><h1>%s</h1><p>%s <strong>%s</strong> · #%d SMLT · %.2f points · global rank #%d</p><p>Hardest demon: <strong>%s</strong></p><p><a href="https://demonlist.org/profile/%d">Demonlist profile</a></p></main>`, html.EscapeString(serverCountryFlag(player.Country)), html.EscapeString(player.Country), html.EscapeString(player.Name), player.Rank, player.Points, player.GlobalRank, html.EscapeString(player.Demon), player.GDLID)
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
	page = addJSONLD(page, meta, pagePath, isEnglish)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(page))
	return true
}

func addJSONLD(page string, meta pageMeta, pagePath string, english bool) string {
	type structuredPage struct {
		Context     string `json:"@context"`
		Type        string `json:"@type"`
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description,omitempty"`
	}
	type website struct {
		Context     string `json:"@context"`
		Type        string `json:"@type"`
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
	}
	var raw []byte
	if strings.HasPrefix(pagePath, "/player/") {
		id, _ := strconv.ParseInt(strings.TrimPrefix(pagePath, "/player/"), 10, 64)
		player, ok := findPublicPlayer(id)
		if ok {
			type person struct {
				Context     string `json:"@context"`
				Type        string `json:"@type"`
				Name        string `json:"name"`
				URL         string `json:"url"`
				Nationality string `json:"nationality"`
				Description string `json:"description"`
			}
			raw, _ = json.Marshal(person{"https://schema.org", "Person", player.Name, meta.canonical, player.Country, fmt.Sprintf("SMLT player with %.2f points and global rank #%d", player.Points, player.GlobalRank)})
		}
	}
	if raw == nil && pagePath == "/" {
		players, _ := dbGetPlayers()
		items := make([]map[string]interface{}, 0, len(players))
		for i, player := range players {
			id := player.GDLID
			if id == 0 {
				id = int64(player.ID)
			}
			items = append(items, map[string]interface{}{"@type": "ListItem", "position": i + 1, "name": player.Name, "url": fmt.Sprintf("https://smlt.lol/player/%d", id)})
		}
		raw, _ = json.Marshal(map[string]interface{}{"@context": "https://schema.org", "@type": "ItemList", "name": "SMLT Leaderboard", "url": "https://smlt.lol/", "itemListElement": items})
	} else {
		name := meta.title
		if english {
			name = meta.title
		}
		raw, _ = json.Marshal(structuredPage{"https://schema.org", "WebPage", name, meta.canonical, meta.description})
	}
	return strings.Replace(page, "</head>", `<script type="application/ld+json">`+string(raw)+`</script></head>`, 1)
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

func serverCountryFlag(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 2 {
		return "🌐"
	}
	return string(rune(code[0]-'A')+0x1F1E6) + string(rune(code[1]-'A')+0x1F1E6)
}

func eventStatusLabel(status string, english bool) string {
	labels := map[string][2]string{
		"ready-verify": {"готов и верифнут", "ready and verified"}, "ready": {"готов", "ready"},
		"in-progress": {"в процессе", "in progress"}, "planned": {"планируется", "planned"},
		"dead": {"мёртв", "dead"}, "frozen": {"заморожен", "frozen"},
	}
	if label, ok := labels[status]; ok {
		if english {
			return label[1]
		}
		return label[0]
	}
	return status
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
		fmt.Fprintf(&out, `<li>%s <a href="%s/player/%d">%s</a> — %.2f points · #%d · hardest: %s</li>`, html.EscapeString(serverCountryFlag(player.Country)), prefix, id, html.EscapeString(player.Name), player.Points, player.GlobalRank, html.EscapeString(player.Demon))
	}
	out.WriteString(`</ol>`)
	if changes, err := dbGetRecentRankChanges(20); err == nil {
		out.WriteString(`<section><h2>`)
		if english {
			out.WriteString(`Recent changes`)
		} else {
			out.WriteString(`Недавние изменения`)
		}
		out.WriteString(`</h2><ul>`)
		for _, change := range changes {
			fmt.Fprintf(&out, `<li>%s <strong>%s</strong> moved from #%d to #%d</li>`, html.EscapeString(serverCountryFlag(change.Country)), html.EscapeString(change.PlayerName), change.OldRank, change.NewRank)
		}
		out.WriteString(`</ul></section>`)
	}
	if events, err := dbGetRecentPlayerEvents(20); err == nil {
		out.WriteString(`<section><h2>`)
		if english {
			out.WriteString(`Player achievements`)
		} else {
			out.WriteString(`Достижения игроков`)
		}
		out.WriteString(`</h2><ul>`)
		for _, event := range events {
			fmt.Fprintf(&out, `<li>%s <strong>%s</strong> — %s%s</li>`, html.EscapeString(serverCountryFlag(event.Country)), html.EscapeString(event.PlayerName), html.EscapeString(event.EventType), func() string {
				if event.Detail != "" {
					return ": " + html.EscapeString(event.Detail)
				}
				return ""
			}())
		}
		out.WriteString(`</ul></section>`)
	}
	out.WriteString(`</main>`)
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
		video := "https://www.youtube.com/watch?v=" + urlQueryEscape(event.VideoID)
		fmt.Fprintf(&out, `<li><strong>%s</strong> — %s · %s · <a href="%s">YouTube</a></li>`, html.EscapeString(event.Title), html.EscapeString(event.Category), html.EscapeString(eventStatusLabel(event.Status, english)), video)
	}
	out.WriteString(`</ul></main>`)
	return out.String()
}

func urlQueryEscape(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "&", "&amp;"), "\"", "&quot;")
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
