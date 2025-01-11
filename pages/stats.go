package pages

import (
	"encoding/json"
	"net/http"
	"wineterfest/stats"
)

type Stats struct {
}

func (s *Stats) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := []*stats.JsonStats{
		{
			Title:       "Stat 1",
			Description: "Stat 1 description",
			Items: []string{
				"A", "B", "C", "D", "E", "F", "G",
			},
		},
		{
			Title:       "Stat 2",
			Description: "Stat 2 description",
			Items: []string{
				"A", "B", "C", "D", "E", "F", "G",
			},
		},
	}
	// Marshal the data and write it to the response
	if err := json.NewEncoder(w).Encode(st); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
