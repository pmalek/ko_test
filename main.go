package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

const (
	gracefulShutdownDuration = 3 * time.Second
	serverAddress            = "127.0.0.1:8000"
)

func main() {
	r := mux.NewRouter()
	r.Path("/products/{key}").Methods(http.MethodGet).HandlerFunc(ProductHandler)
	r.Path("/").Methods(http.MethodGet).HandlerFunc(HomeHandler)

	srv := CreateServer(r, serverAddress)
	log.Printf("Starting server at %v", serverAddress)
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Error running server: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownDuration)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("Shutting down")
	os.Exit(0)
}

func CreateServer(r *mux.Router, addr string) *http.Server {
	return &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func ProductHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	k, ok := vars["key"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `Product key not provided`)
		return
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `Product '`+k+`'`)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `Home`)
}
