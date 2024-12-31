package pages

import (
	"fmt"
	"io"
	"net/http"
)

type RegisterWine struct {
}

func (s *RegisterWine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "./html/register-wine.html") // Serve the HTML file
	} else {
		s.register(w, r)
	}
}

type WineRegistration struct {
	WineName         string `json:"wineName"`
	WinePrice        string `json:"winePrice"`
	AnonymizedNumber string `json:"anonymizedNumber"`
	Username         string `json:"username"`
}

func (s *RegisterWine) register(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	all, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	fmt.Println(string(all))
	http.Redirect(w, r, "/", 302)
}
