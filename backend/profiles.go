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
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, location := range locations {
		fmt.Fprintf(w, "<url><loc>%s</loc></url>", html.EscapeString(location))
	}
	fmt.Fprint(w, `</urlset>`)
}
