package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tomtorh96/grocery-app/internal/auth"
	"github.com/tomtorh96/grocery-app/internal/db"
	"github.com/tomtorh96/grocery-app/internal/handlers"
	"github.com/tomtorh96/grocery-app/internal/ws"
)

func main() {
	// connect to database
	if err := db.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	fmt.Println("connected to database")

	// websocket hub
	hub := ws.NewHub()

	// router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// public routes
	r.Post("/auth/register", handlers.Register)
	r.Post("/auth/login", handlers.Login)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware)

		// lists
		r.Post("/lists", handlers.CreateList)
		r.Get("/lists", handlers.GetLists)
		r.Get("/lists/{id}", handlers.GetList)
		r.Post("/lists/join", handlers.JoinList)

		// items
		r.Post("/lists/{id}/items", handlers.AddItem(hub))
		r.Delete("/lists/{id}/items/{itemId}", handlers.DeleteItem(hub))
		r.Patch("/lists/{id}/items/{itemId}", handlers.MarkItem(hub))
		r.Post("/lists/{id}/reset", handlers.ResetItems(hub))
		r.Put("/lists/{id}/items/{itemId}", handlers.EditItem(hub))

		// history
		r.Get("/lists/{id}/history", handlers.GetHistory)

		// websocket
		r.Get("/lists/{id}/ws", ws.ServeWS(hub))

		// members
		r.Get("/lists/{id}/members", handlers.GetMembers)
		r.Delete("/lists/{id}/members/{userId}", handlers.RemoveMember(hub))
		r.Patch("/lists/{id}/members/{userId}", handlers.UpdateMemberRole(hub))
		r.Patch("/lists/{id}", handlers.RenameList(hub))
		r.Delete("/lists/{id}/leave", handlers.LeaveList(hub))
		r.Delete("/lists/{id}/members/{userId}", handlers.RemoveMember(hub))

	})

	// HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("server running on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	fmt.Println("shutting down...")
}
