package pages

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Signup struct {
}

type UsernameReq struct {
	Username string `json:"username"`
}

func (s *Signup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, "./html/signup.html") // Serve the HTML file
	} else {
		s.signUp(w, r)
	}
}

func (s *Signup) signUp(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	all, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	req := &UsernameReq{}
	if err := json.Unmarshal(all, req); err != nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	fmt.Println(req.Username)
	// Simulate username validation
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"valid": true}`))
}
