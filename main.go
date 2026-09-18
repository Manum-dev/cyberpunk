package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Driver SQLite in puro Go (zero dipendenze esterne)
)

// Work rappresenta un'opera del portfolio cyberpunk
type Work struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

// Message rappresenta un messaggio salvato nel database
type Message struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Category  string    `json:"category"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ContactRequest definisce i dati attesi dal form di contatto
type ContactRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

// Opere predefinite in memoria
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

// Variabile globale per la connessione al database
var db *sql.DB

// initDB inizializza la connessione a SQLite e crea la tabella messaggi
func initDB(dataSourceName string) (*sql.DB, error) {
	database, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("apertura database fallita: %w", err)
	}

	// Query SQL per creare la tabella 'messages' se non esiste già
	query := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		category TEXT DEFAULT 'SEGNALE GENERICO',
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := database.Exec(query); err != nil {
		return nil, fmt.Errorf("creazione tabella fallita: %w", err)
	}

	// Assicura che la colonna 'category' esista anche per database già creati
	_, _ = database.Exec("ALTER TABLE messages ADD COLUMN category TEXT DEFAULT 'SEGNALE GENERICO'")

	return database, nil
}

// sendDiscordNotification invia un messaggio formattato in stile Cyberpunk a Discord
func sendDiscordNotification(webhookURL string, req ContactRequest) {
	type DiscordField struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Inline bool   `json:"inline"`
	}

	type DiscordFooter struct {
		Text string `json:"text"`
	}

	type DiscordEmbed struct {
		Title     string         `json:"title"`
		Color     int            `json:"color"`
		Fields    []DiscordField `json:"fields"`
		Footer    DiscordFooter  `json:"footer"`
		Timestamp string         `json:"timestamp"`
	}

	type DiscordPayload struct {
		Username string         `json:"username"`
		Content  string         `json:"content,omitempty"`
		Embeds   []DiscordEmbed `json:"embeds"`
	}

	// Colori ed emoji dinamici in base al tipo di trasmissione
	color := 16737876 // #ff6654 Coral per generale
	categoryEmoji := "📡"
	pingContent := ""

	switch req.Category {
	case "COMMISSIONE ART":
		color = 16724582 // #ff3366 Neon Magenta
		categoryEmoji = "💎"
		pingContent = "@here 🚨 **[PRIORITÀ ALTA]** Nuova richiesta di commissione ricevuta!"
	case "COLLABORAZIONE":
		color = 61695 // #00f0ff Neon Cyan
		categoryEmoji = "🤝"
		pingContent = "⚡ **[PROPOSTA]** Nuova collaborazione in arrivo!"
	default:
		req.Category = "SEGNALE GENERICO"
	}

	embed := DiscordEmbed{
		Title: fmt.Sprintf("⚡ Nuova trasmissione ricevuta da NOVA [%s]", req.Category),
		Color: color,
		Fields: []DiscordField{
			{Name: "👤 Mittente / Callsign", Value: fmt.Sprintf("`%s`", req.Name), Inline: true},
			{Name: "📡 Frequenza / Contatto", Value: fmt.Sprintf("`%s`", req.Email), Inline: true},
			{Name: fmt.Sprintf("%s Protocollo", categoryEmoji), Value: fmt.Sprintf("**%s**", req.Category), Inline: true},
			{Name: "💬 Dati Trasmissione (Payload)", Value: req.Message, Inline: false},
		},
		Footer: DiscordFooter{
			Text: "NOVA Cyberpunk Terminal • SQLite // Go Server",
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	payload := DiscordPayload{
		Username: "NOVA Neural Uplink",
		Content:  pingContent,
		Embeds:   []DiscordEmbed{embed},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("⚠️ Errore serializzazione Discord: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("⚠️ Errore invio a Discord: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Printf("🚀 Notifica Discord recapitata! (Status: %d)", resp.StatusCode)
}

// loggingMiddleware registra ogni richiesta HTTP nel terminale
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s (%s)", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// loadEnv carica le variabili da un file .env locale (ignorato da git)
func loadEnv(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // Se .env non esiste (es. su Vercel), ignora silenziosamente
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}

func main() {
	// Carica credenziali da .env locale (se presente)
	loadEnv(".env")

	// 1. Inizializzazione Database SQLite
	var err error
	db, err = initDB("cyberpunk.db")
	if err != nil {
		log.Fatalf("Errore inizializzazione database: %v", err)
	}
	defer db.Close()
	log.Println("💾 Database SQLite (cyberpunk.db) connesso con successo!")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()

	// 2. Endpoint API per ottenere le opere in JSON
	mux.HandleFunc("GET /api/works", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(works)
	})

	// 3. Endpoint per inviare un messaggio (salvataggio in SQLite + Discord opzionale)
	mux.HandleFunc("POST /api/contact", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req ContactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "JSON non valido"})
			return
		}

		if req.Name == "" || req.Message == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Nome e messaggio sono obbligatori"})
			return
		}

		if req.Category == "" {
			req.Category = "SEGNALE GENERICO"
		}

		// Inserimento nel database SQLite con prepared statement
		query := "INSERT INTO messages (name, email, category, content) VALUES (?, ?, ?, ?)"
		res, err := db.Exec(query, req.Name, req.Email, req.Category, req.Message)
		if err != nil {
			log.Printf("Errore salvataggio messaggio: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Impossibile salvare il messaggio nel database"})
			return
		}

		id, _ := res.LastInsertId()
		log.Printf("📨 Trasmissione #%d [%s] registrata da %s (%s)", id, req.Category, req.Name, req.Email)

		// GOROUTINE: se DISCORD_WEBHOOK_URL è impostato, esegui l'invio in background
		webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
		if webhookURL != "" {
			go sendDiscordNotification(webhookURL, req)
		} else {
			log.Println("⚠️ [AVVISO DISCORD] DISCORD_WEBHOOK_URL non è impostato in .env. La notifica Discord non può essere inviata.")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":  true,
			"id":       id,
			"category": req.Category,
			"message":  "Segnale salvato nel database locale!",
		})
	})

	// 4. Endpoint per consultare tutti i messaggi archiviati nel database SQLite
	mux.HandleFunc("GET /api/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		rows, err := db.Query("SELECT id, name, email, COALESCE(category, 'SEGNALE GENERICO'), content, created_at FROM messages ORDER BY id DESC")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Errore lettura messaggi dal database"})
			return
		}
		defer rows.Close()

		var messages []Message
		for rows.Next() {
			var m Message
			if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Category, &m.Content, &m.CreatedAt); err != nil {
				log.Printf("Errore scansione riga: %v", err)
				continue
			}
			messages = append(messages, m)
		}

		if messages == nil {
			messages = []Message{} // Restituisce [] invece di null in JSON
		}

		_ = json.NewEncoder(w).Encode(messages)
	})

	// 5. Servizio file statici e pagina principale
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}

		cleanPath := filepath.Clean(filepath.Join(".", r.URL.Path))
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
	log.Printf("📡 API Opere:    http://localhost:%s/api/works", port)
	log.Printf("📨 API Messaggi: http://localhost:%s/api/messages", port)
	if os.Getenv("DISCORD_WEBHOOK_URL") != "" {
		log.Printf("⚡ Discord Webhook: ATTIVO (notifiche abilitate)")
	} else {
		log.Printf("⚠️ Discord Webhook: NON IMPOSTATO (crea .env con DISCORD_WEBHOOK_URL=...)")
	}
	log.Printf("═══════════════════════════════════════════════════════")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Errore durante l'esecuzione del server: %v", err)
	}
}
