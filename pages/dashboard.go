package pages

import (
	"net/http"
)

type Dashboard struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/dashboard.html") // Serve the HTML file
}
