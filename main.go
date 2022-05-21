package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/room/{room}/player/{player}", handler)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":8001", r)
}
