package main

import (
	"fmt"
	"log"
	"net/http"
	"wineterfest/pages"
)

func main() {
	http.Handle("/", &pages.Home{})
	http.Handle("/signup", &pages.Signup{})
	http.Handle("/dashboard", &pages.Dashboard{})
	http.Handle("/register-wine", &pages.RegisterWine{})
	http.Handle("/rate-wine", &pages.RateWine{})
	fmt.Println("Listening on port 8080!")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
