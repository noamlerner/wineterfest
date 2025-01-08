package pages

import (
	"encoding/json"
	"net/http"
	"wineterfest/datamodels"
	"wineterfest/winedb"
)

type MyWines struct {
	CL *winedb.Client
}

func (s *MyWines) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	wines, err := s.CL.AllWines(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Set response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	myWines := make([]datamodels.Wine, 0, len(wines))
	for _, wine := range wines {
		if wine.Username == username {
			myWines = append(myWines, wine)
		}
	}
	// Marshal the data and write it to the response
	if err := json.NewEncoder(w).Encode(myWines); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
