package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// BugReport - Represents the HTTP/POST request for the feedbackHandler.
type BugReport struct {
	Title   string
	Body    string
	Link    string
	Version string
}

// main - Main entry point for the application.
func main() {
	config := loadConfig()
	setupHttpServer(config)
}

// loadConfig - Loads application secrets using VIPER and assigns
// them to the database.go variables.
func loadConfig() Config {
	config, err := LoadConfig(".")

	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	Hostname = config.Hostname
	Port = config.Port
	Username = config.Username
	Password = config.Password
	Database = config.Database
	return config
}

// setupHttpServer - Creates the routes and handlers for the server.
// Also creates the goroutine which will keep the application alive.
// Terminates gracefully on os.Signal interrupt 1.
func setupHttpServer(config Config) {
	baseMux := mux.NewRouter()

	s := http.Server{
		Addr:         config.HttpPort,
		Handler:      baseMux,
		ErrorLog:     nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	postMux := baseMux.Methods(http.MethodPost).Subrouter()
	postMux.HandleFunc("/feedback", feedbackHandler)

	go func() {
		log.Println("Listening to", config.HttpPort)
		err := s.ListenAndServe()
		if err != nil {
			log.Printf("Error starting server: %s\n", err)
			return
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	sig := <-sigs
	log.Println("Received sigterm:", sig)
	_ = s.Shutdown(nil)
}

// feedbackHandler - Handles the feedback request from the Stellar
// client. Can return a 400, 201 or 503.
func feedbackHandler(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)

	var data BugReport
	err := decoder.Decode(&data)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Incorrect request format."))
		return
	}

	if len(data.Title) == 0 || len(data.Body) == 0 || len(data.Link) == 0 || len(data.Version) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("{\"result\": \"Failure\"}\n"))
		return
	}

	didInsert := LogFeedback(data)

	w.Header().Set("Content-Type", "application/json")
	if didInsert {
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte("{\"result\": \"Success\"}\n"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err = w.Write([]byte("{\"result\": \"Success\"}\n"))
	}
}
