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
	myWines, err := s.CL.UsersWines(r.Context(), (&datamodels.User{username}).Normalize())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Set response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")
	// Marshal the data and write it to the response
	if err := json.NewEncoder(w).Encode(myWines); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
