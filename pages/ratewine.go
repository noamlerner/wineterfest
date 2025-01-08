package pages

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wineterfest/datamodels"
	"wineterfest/winedb"
)

type RateWine struct {
	CL *winedb.Client
}

func (s *RateWine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "html/rate-wine.html")
	} else {
		s.rateWine(w, r)
	}
}

func (s *RateWine) rateWine(w http.ResponseWriter, r *http.Request) {
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
	wineRating := datamodels.WineRating{}
	err = json.Unmarshal(all, &wineRating)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
	}

	err = s.CL.CreateWineRating(r.Context(), &wineRating)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
	}

}
