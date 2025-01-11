package pages

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wineterfest/winedb"
)

type Signup struct {
	CL *winedb.Client
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

	if req.Username == "" || len(req.Username) > 200 {
		http.Error(w, "Please send a request body", 400)
	}
	fmt.Println(req.Username)

	err = s.CL.CreateUser(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Could not create user, try a different name", 400)
		return
	}

	// Simulate username validation
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"valid": true}`))
}
