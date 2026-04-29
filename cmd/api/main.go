package main

import (
	"fmt"
	"log"
	"net/http"

	"concerts/internal/config"
	"concerts/internal/database"
	"concerts/internal/handlers"
	"concerts/internal/middleware"
	"concerts/internal/repository"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("DB connect error:", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	// Repositories
	concertRepo := repository.NewConcertRepo(db)
	seatingRepo := repository.NewSeatingRepo(db)
	reservationRepo := repository.NewReservationRepo(db)
	bookingRepo := repository.NewBookingRepo(db)

	// Handlers
	concertH := handlers.NewConcertHandler(concertRepo)
	seatingH := handlers.NewSeatingHandler(concertRepo, seatingRepo)
	reservationH := handlers.NewReservationHandler(concertRepo, seatingRepo, reservationRepo)
	bookingH := handlers.NewBookingHandler(concertRepo, seatingRepo, reservationRepo, bookingRepo)
	ticketH := handlers.NewTicketHandler(bookingRepo, seatingRepo)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.ColorLogger)
	r.Use(chiMiddleware.Recoverer)

	r.Route("/api/v1", func(r chi.Router) {
		// Concerts
		r.Get("/concerts", concertH.List)
		r.Get("/concerts/{concert-id}", concertH.Get)

		// Seating
		r.Get("/concerts/{concert-id}/shows/{show-id}/seating", seatingH.GetSeating)

		// Reservation
		r.Post("/concerts/{concert-id}/shows/{show-id}/reservation", reservationH.Reserve)

		// Booking
		r.Post("/concerts/{concert-id}/shows/{show-id}/booking", bookingH.Book)

		// Tickets
		r.Post("/tickets", ticketH.GetTickets)
		r.Post("/tickets/{ticket-id}/cancel", ticketH.CancelTicket)
	})

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server started on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
