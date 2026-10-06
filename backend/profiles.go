package main

import (
	"fmt"
	"html"
	"net/http"
)

func serveSitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	locations := []string{"https://smlt.lol/", "https://smlt.lol/events", "https://smlt.lol/about"}
	locations = append(locations, "https://smlt.lol/en", "https://smlt.lol/en/events", "https://smlt.lol/en/about")
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, location := range locations {
		fmt.Fprintf(w, "<url><loc>%s</loc></url>", html.EscapeString(location))
	}
	if players, err := dbGetPlayers(); err == nil {
		for _, player := range players {
			id := player.GDLID
			if id == 0 {
				id = int64(player.ID)
			}
			last := ""
			if !player.UpdatedAt.IsZero() {
				last = fmt.Sprintf("<lastmod>%s</lastmod>", player.UpdatedAt.UTC().Format("2006-01-02"))
			}
			fmt.Fprintf(w, "<url><loc>https://smlt.lol/player/%d</loc>%s</url>", id, last)
			fmt.Fprintf(w, "<url><loc>https://smlt.lol/en/player/%d</loc>%s</url>", id, last)
		}
	}
	fmt.Fprint(w, `</urlset>`)
}
