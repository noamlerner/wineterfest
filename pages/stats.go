package pages

import (
	"encoding/json"
	"net/http"
	"wineterfest/stats"
	"wineterfest/winedb"
)

type Stats struct {
	CL *winedb.Client
}

func (s *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	allWines, err := s.CL.AllWines(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	allRatings, err := s.CL.AllRatings(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	st := stats.Calc(allWines, allRatings)
	// Marshal the data and write it to the response
	if err := json.NewEncoder(w).Encode(st); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
