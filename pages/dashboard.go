package pages

import (
	"net/http"
)

type Dashboard struct {
}

func (s *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/dashboard.html") // Serve the HTML file
}
