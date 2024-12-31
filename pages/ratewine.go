package pages

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RateWine struct {
}
type WineRating struct {
	AnonymizedNumber int    `json:"anonymizedNumber"`
	Rating           int    `json:"rating"`
	Wineuser         string `json:"wineuser"`
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

	wineRating := WineRating{}
	err = json.Unmarshal(all, &wineRating)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
	}

}
