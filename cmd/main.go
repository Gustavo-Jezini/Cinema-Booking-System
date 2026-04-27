package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /movies", listMovies)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

var movies = []movieResponse{
	{ID: "inception", Title: "Inception", Rows: 5, SeatsPerRows: 8},
	{ID: "dune", Title: "Dune: Part Two", Rows: 4, SeatsPerRows: 6},
}

func listMovies(w http.ResponseWriter, r *http.Request) {
	WriteJson(w, http.StatusOK, movies)
}

func WriteJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(v)
}

type movieResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Rows         int    `json:"rows"`
	SeatsPerRows int    `json:"seats_per_row"`
}
