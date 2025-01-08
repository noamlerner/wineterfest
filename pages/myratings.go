package pages

import (
	"encoding/json"
	"net/http"
	"wineterfest/winedb"
)

type MyRatings struct {
	CL *winedb.Client
}

func (s *MyRatings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	wineRatings, err := s.CL.MyWineRatings(r.Context(), username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Set response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Marshal the data and write it to the response
	if err := json.NewEncoder(w).Encode(wineRatings); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
