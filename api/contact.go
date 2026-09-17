package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// ContactRequest definisce i dati attesi dal form di contatto
type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type DiscordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type DiscordEmbed struct {
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color"`
	Fields      []DiscordEmbedField `json:"fields"`
	Footer      struct {
		Text string `json:"text"`
	} `json:"footer"`
	Timestamp string `json:"timestamp"`
}

type DiscordWebhookPayload struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds"`
}

// Handler è l'entrypoint per Vercel Serverless Function
func Handler(w http.ResponseWriter, r *http.Request) {
	// Intestazioni CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Metodo non consentito. Richiesto POST."}`, http.StatusMethodNotAllowed)
		return
	}

	var req ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "JSON non valido"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Message == "" {
		http.Error(w, `{"error": "Nome e messaggio sono obbligatori"}`, http.StatusBadRequest)
		return
	}

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		// Se il webhook non è ancora configurato, simuliamo la ricezione corretta
		log.Printf("⚠️ [LOCAL] DISCORD_WEBHOOK_URL non impostato. Messaggio da %s (%s): %s\n", req.Name, req.Email, req.Message)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "message": "Segnale ricevuto in locale (imposta DISCORD_WEBHOOK_URL per inoltrarlo a Discord)"}`))
		return
	}

	// Costruiamo il messaggio formattato con stile Cyberpunk per Discord
	embed := DiscordEmbed{
		Title: "⚡ Nuova trasmissione ricevuta da NOVA",
		Color: 16737876, // Colore accento #ff6654 in decimale
		Fields: []DiscordEmbedField{
			{Name: "👤 Mittente / Handle", Value: req.Name, Inline: true},
			{Name: "📡 Canale / Email", Value: req.Email, Inline: true},
			{Name: "💬 Dati di Trasmissione", Value: req.Message, Inline: false},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	embed.Footer.Text = "NOVA Cyberpunk Terminal • Ricevuto via Vercel Go"

	payload := DiscordWebhookPayload{
		Username: "NOVA System",
		Embeds:   []DiscordEmbed{embed},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, `{"error": "Errore interno durante la serializzazione"}`, http.StatusInternalServerError)
		return
	}

	// Invio HTTP POST a Discord
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Errore chiamata Discord: %v", err)
		http.Error(w, `{"error": "Impossibile contattare Discord"}`, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Discord ha restituito status: %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf(`{"error": "Discord ha restituito errore %d"}`, resp.StatusCode), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true, "message": "Trasmissione inviata al terminale Discord!"}`))
}
