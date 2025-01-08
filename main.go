package main

import (
	"fmt"
	"log"
	"net/http"
	"wineterfest/pages"
	"wineterfest/winedb"
)

func main() {
	cl := winedb.Conn()
	http.Handle("/", &pages.Home{})
	http.Handle("/signup", &pages.Signup{cl})
	http.Handle("/dashboard", &pages.Dashboard{})
	http.Handle("/register-wine", &pages.RegisterWine{cl})
	http.Handle("/rate-wine", &pages.RateWine{cl})
	http.Handle("/my-ratings", &pages.MyRatings{cl})
	http.Handle("/my-wines", &pages.MyWines{cl})
	fmt.Println("Listening on port 8080!")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
