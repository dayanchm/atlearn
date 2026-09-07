package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type SessionRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func main() {
	local := flag.Bool("local", false, "publish locally every minute instead of starting the HTTP server")
	flag.Parse()
	// Load local settings without overriding existing environment variables.
	_ = godotenv.Load()
	if *local {
		if err := runLocalBot(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/daily", dailyHandler(publishScheduledPoem))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Daily poetry bot is running.\n"))
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
