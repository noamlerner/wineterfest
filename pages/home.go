package pages

import (
	"net/http"
)

type Home struct {
}

func (s *Home) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/home.html")
}
