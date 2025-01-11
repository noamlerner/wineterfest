package pages

import "net/http"

type StatsDisplay struct {
}

func (s *StatsDisplay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/stats-display.html") // Serve the HTML file
}
