package main

import (
	"net/http"

	"github.com/limeleaf-coop/knbn/pkg"
)

func main() {
	http.HandleFunc("GET /boards/{id}", pkg.BoardHandler)
	http.HandleFunc("GET /boards", pkg.BoardsHandler)
	http.HandleFunc("GET /", pkg.SignInHandler)
	http.ListenAndServe(":8080", nil)
}
