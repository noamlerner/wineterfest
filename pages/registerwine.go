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

	if wineRegistration.WineName == "" ||
		wineRegistration.WinePrice < 0.0 || wineRegistration.WinePrice > 500 ||
		wineRegistration.AnonymizedNumber < 0 || wineRegistration.AnonymizedNumber > 100 ||
		len(wineRegistration.WineName) > 1000 {
		http.Error(w, "Please send a request body", 400)
	}

	err = s.CL.CreateWine(r.Context(), &wineRegistration)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	http.Redirect(w, r, "/", 302)
}
