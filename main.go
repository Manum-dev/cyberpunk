package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cyberpunk/api"
)

// Work rappresenta un'opera del portfolio cyberpunk
type Work struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

// Opere predefinite (possono essere estese o collegate a un database in futuro)
var works = []Work{
	{
		ID:          "city",
		Title:       "Neon District",
		Category:    "001 / ENVIRONMENT",
		Description: "Una città senza orizzonte. Tra nebbia, cemento e insegne luminose, ogni finestra custodisce una storia ancora da raccontare.",
		Image:       "assets/city.png",
	},
	{
		ID:          "airship",
		Title:       "Silent Voyager",
		Category:    "002 / VEHICLE DESIGN",
		Description: "Sopra il rumore, un viaggio lento. Un dirigibile attraversa il cielo notturno, portando con sé le luci di una città lontana.",
		Image:       "assets/airship.png",
	},
}

// loggingMiddleware registra ogni richiesta HTTP nel terminale
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s (%s)", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	// 1. Endpoint API per recuperare le opere in formato JSON
	mux.HandleFunc("GET /api/works", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(works); err != nil {
			http.Error(w, `{"error": "impossibile serializzare i dati"}`, http.StatusInternalServerError)
		}
	})

	// 2. Endpoint di stato/health
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "time": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// 3. Endpoint per il form di contatto (compatibile con Vercel)
	mux.HandleFunc("POST /api/contact", handler.Handler)
	mux.HandleFunc("OPTIONS /api/contact", handler.Handler)

	// 4. Servizio della pagina principale e dei file statici (CSS, JS, immagini)
	// Se la richiesta è '/', restituiamo 'index.html'
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}

		// Preveniamo accessi a percorsi al di fuori della cartella corrente
		cleanPath := filepath.Clean(filepath.Join(".", r.URL.Path))
		
		// Verifichiamo se il file esiste
		if info, err := os.Stat(cleanPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, cleanPath)
			return
		}

		http.NotFound(w, r)
	})

	handler := loggingMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("═══════════════════════════════════════════════════════")
	log.Printf("🚀 Server Cyberpunk avviato con successo!")
	log.Printf("🌐 Apri il browser su: http://localhost:%s", port)
	log.Printf("📡 API Opere: http://localhost:%s/api/works", port)
	log.Printf("⚡ API Contatto: http://localhost:%s/api/contact", port)
	log.Printf("═══════════════════════════════════════════════════════")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Errore durante l'esecuzione del server: %v", err)
	}
}
