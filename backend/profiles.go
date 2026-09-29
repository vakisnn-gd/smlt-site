package main

import (
	"fmt"
	"html"
	"net/http"
	"time"
)

func serveSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	locations := []string{"https://smlt.lol/", "https://smlt.lol/events", "https://smlt.lol/about"}
	locations = append(locations, "https://smlt.lol/en", "https://smlt.lol/en/events", "https://smlt.lol/en/about")
	if players, err := dbGetPlayers(); err == nil {
		for _, player := range players {
			id := player.GDLID
			if id == 0 {
				id = int64(player.ID)
			}
			locations = append(locations, fmt.Sprintf("https://smlt.lol/player/%d", id))
			locations = append(locations, fmt.Sprintf("https://smlt.lol/en/player/%d", id))
		}
	}
	lastModified := time.Now().UTC().Format("2006-01-02")
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, location := range locations {
		fmt.Fprintf(w, "<url><loc>%s</loc><lastmod>%s</lastmod></url>", html.EscapeString(location), lastModified)
	}
	fmt.Fprint(w, `</urlset>`)
}
