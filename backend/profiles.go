package main

import (
	"fmt"
	"html"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func pageStart(title, description, canonical string) string {
	return `<!doctype html><html lang="ru"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta name="theme-color" content="#0d1117"><title>` + html.EscapeString(title) + `</title><meta name="description" content="` + html.EscapeString(description) + `"><link rel="canonical" href="` + canonical + `"><meta property="og:type" content="website"><meta property="og:title" content="` + html.EscapeString(title) + `"><meta property="og:description" content="` + html.EscapeString(description) + `"><meta property="og:url" content="` + canonical + `"><link rel="icon" href="/favicon2.ico?v=3"><link rel="stylesheet" href="/profile.css"></head>`
}

func servePlayersDirectory(w http.ResponseWriter, r *http.Request) {
	players, err := dbGetPlayers()
	if err != nil {
		http.Error(w, "database error", 500)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	countrySet := make(map[string]bool)
	for _, p := range players {
		if p.Country != "" {
			countrySet[p.Country] = true
		}
	}
	countries := make([]string, 0, len(countrySet))
	for country := range countrySet {
		countries = append(countries, country)
	}
	sort.Strings(countries)
	fmt.Fprint(w, pageStart("Игроки SMLT", "Профили, позиции и достижения игроков сообщества SMLT.", "https://smlt.lol/players"))
	fmt.Fprint(w, `<body><main class="profile-shell"><nav><a href="/">← Рейтинг</a><a href="/events">Ивенты</a></nav><header class="directory-head"><p class="eyebrow">SMLT LEADERBOARD</p><h1>Профили игроков</h1><p>Поиск по нику, стране или сложнейшему пройденному уровню.</p><div class="profile-filters"><label class="profile-search">Поиск<input id="profile-search" type="search" autocomplete="off" placeholder="Введите ник или уровень"></label><label>Страна<select id="country-filter"><option value="">Все страны</option>`)
	for _, country := range countries {
		fmt.Fprintf(w, `<option value="%s">%s</option>`, html.EscapeString(strings.ToLower(country)), html.EscapeString(country))
	}
	fmt.Fprint(w, `</select></label><label>Сортировка<select id="profile-sort"><option value="rank">По месту</option><option value="points">По очкам</option><option value="name">По нику</option></select></label></div></header><section class="method"><h2>Как считается рейтинг</h2><p>Очки и глобальная позиция синхронизируются с Demonlist. Внутреннее место SMLT определяется по количеству очков; при равенстве используется ник. История обновляется при каждом изменении записи.</p></section><section class="player-cards" id="player-cards">`)
	for _, p := range players {
		href := "/player/" + url.PathEscape(p.Name)
		fmt.Fprintf(w, `<a class="player-card" href="%s" data-search="%s %s %s" data-country="%s" data-rank="%d" data-points="%.4f" data-name="%s"><span class="player-rank">#%d</span><strong>%s</strong><span>%s · %.2f pts</span><small>%s · global #%d</small></a>`, href, html.EscapeString(strings.ToLower(p.Name+" "+p.Country+" "+p.Demon)), html.EscapeString(p.Country), html.EscapeString(p.Demon), html.EscapeString(strings.ToLower(p.Country)), p.Rank, p.Points, html.EscapeString(strings.ToLower(p.Name)), p.Rank, html.EscapeString(p.Name), html.EscapeString(p.Country), p.Points, html.EscapeString(p.Demon), p.GlobalRank)
	}
	fmt.Fprint(w, `</section><p class="updated">Страница сформирована `+time.Now().UTC().Format("02.01.2006 15:04 UTC")+`</p></main><script src="/profiles.js" defer></script></body></html>`)
}

func servePlayerPage(w http.ResponseWriter, r *http.Request) {
	name, _ := url.PathUnescape(strings.TrimPrefix(r.URL.Path, "/player/"))
	p, err := dbGetPlayer(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	canonical := "https://smlt.lol/player/" + url.PathEscape(p.Name)
	description := fmt.Sprintf("%s — #%d в рейтинге SMLT, %.2f очков, сложнейший демон: %s.", p.Name, p.Rank, p.Points, p.Demon)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, pageStart(p.Name+" · SMLT", description, canonical))
	fmt.Fprintf(w, `<body data-player="%s"><main class="profile-shell"><nav><a href="/">← Рейтинг</a><a href="/players">Все игроки</a></nav><article class="profile-hero"><span class="country">%s</span><p class="eyebrow">ИГРОК SMLT</p><h1>%s</h1><div class="profile-stats"><div><strong>#%d</strong><span>место SMLT</span></div><div><strong>%.2f</strong><span>очков</span></div><div><strong>#%d</strong><span>в мире</span></div></div><p class="hardest">Сложнейший демон: <strong>%s</strong></p></article><section class="chart-card"><div><p class="eyebrow">ИСТОРИЯ</p><h2>Изменение очков</h2></div><svg id="history-chart" viewBox="0 0 800 260" role="img" aria-label="График очков"></svg><p id="history-empty" hidden>История появится после следующих обновлений рейтинга.</p></section></main><script src="/profile.js" defer></script></body></html>`, html.EscapeString(p.Name), html.EscapeString(p.Country), html.EscapeString(p.Name), p.Rank, p.Points, p.GlobalRank, html.EscapeString(p.Demon))
}

func serveSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	locations := []string{"https://smlt.lol/", "https://smlt.lol/events", "https://smlt.lol/about", "https://smlt.lol/players"}
	if players, err := dbGetPlayers(); err == nil {
		for _, player := range players {
			locations = append(locations, "https://smlt.lol/player/"+url.PathEscape(player.Name))
		}
	}
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, location := range locations {
		fmt.Fprintf(w, "<url><loc>%s</loc></url>", html.EscapeString(location))
	}
	fmt.Fprint(w, `</urlset>`)
}
