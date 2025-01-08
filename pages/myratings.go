package pages

import (
	"encoding/json"
	"net/http"
	"wineterfest/datamodels"
	"wineterfest/winedb"
)

type MyRatings struct {
	CL *winedb.Client
}

func (s *MyRatings) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wineRatings := []*datamodels.WineRating{
		{
			AnonymizedNumber: 1,
			Rating:           1,
		},
		{
			AnonymizedNumber: 2,
			Rating:           2,
		},
		{
			AnonymizedNumber: 3,
			Rating:           3,
		},
	}
	// Set response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	// Marshal the data and write it to the response
	if err := json.NewEncoder(w).Encode(wineRatings); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
