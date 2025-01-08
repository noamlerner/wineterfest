package pages

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wineterfest/datamodels"
	"wineterfest/winedb"
)

type RegisterWine struct {
	CL *winedb.Client
}

func (s *RegisterWine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "./html/register-wine.html") // Serve the HTML file
	} else {
		s.register(w, r)
	}
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

	wineRegistration := datamodels.Wine{}
	err = json.Unmarshal(all, &wineRegistration)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
	}

	if wineRegistration.WineName == "" {
		http.Error(w, "Wine must have a name!", 400)
		return
	}
	if wineRegistration.WinePrice < 0.0 {
		http.Error(w, "Wine must have a positive price!", 400)
		return
	}
	if wineRegistration.WinePrice > 500 {
		http.Error(w, "No chance you bought a wine that expensive.", 400)
		return
	}
	if wineRegistration.AnonymizedNumber < 0 {
		http.Error(w, "No anonymous wine has a negative number!", 400)
		return
	}
	if len(wineRegistration.WineName) > 1000 {
		http.Error(w, "Wine must have a maximum length of 1000!", 400)
		return
	}

	err = s.CL.CreateWine(r.Context(), &wineRegistration)
	if err != nil {
		http.Error(w, "Number already in use", 403)
		return
	}
	http.Redirect(w, r, "/", 302)
}
