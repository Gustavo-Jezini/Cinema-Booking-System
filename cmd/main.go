package main

import (
	"cinema-booking-system/internal/adapters/redis"
	"cinema-booking-system/internal/booking"
	"cinema-booking-system/internal/utils"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /movies", listMovies)

	mux.Handle("GET /", http.FileServer(http.Dir("static")))

	store := booking.NewRedisStore(redis.NewClient("localhost:6379"))
	svc := booking.NewService(store)

	bookingHandler := booking.NewHandler(svc)

	mux.HandleFunc("GET /movies/{movieID}/seats", bookingHandler.ListSeats)
	mux.HandleFunc("POST /movies/{movieID}/seats/{seatID}/hold", bookingHandler.HoldSeat)
	mux.HandleFunc("PUT /sessions/{sessionID}/confirm", bookingHandler.ConfirmSession)
	mux.HandleFunc("DELETE /sessions/{sessionID}", bookingHandler.ReleaseSession)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

var movies = []movieResponse{
	{ID: "inception", Title: "Inception", Rows: 5, SeatsPerRows: 8},
	{ID: "dune", Title: "Dune: Part Two", Rows: 4, SeatsPerRows: 6},
}

func listMovies(w http.ResponseWriter, r *http.Request) {
	utils.WriteJson(w, http.StatusOK, movies)
}

type movieResponse struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Rows         int    `json:"rows"`
	SeatsPerRows int    `json:"seats_per_row"`
}
