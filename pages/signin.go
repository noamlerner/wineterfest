package pages

import (
	"net/http"
	"wineterfest/winedb"
)

type SignIn struct {
	CL *winedb.Client
}

func (s *SignIn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./html/signup.html") // Serve the HTML file
}
