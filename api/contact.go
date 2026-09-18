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

// ContactRequest definisce i dati attesi dal form
type ContactRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

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

// Handler è l'entrypoint per la Serverless Function Go di Vercel
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Metodo non consentito. Richiesto POST."})
		return
	}

	var req ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "JSON payload non valido"})
		return
	}

	if req.Name == "" || req.Message == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Nome e messaggio sono obbligatori"})
		return
	}

	if req.Category == "" {
		req.Category = "SEGNALE GENERICO"
	}

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		log.Printf("⚠️ DISCORD_WEBHOOK_URL non configurato su Vercel. Ricevuto: %s da %s", req.Category, req.Name)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "Segnale ricevuto! (Configura DISCORD_WEBHOOK_URL nelle impostazioni di Vercel per inoltrarlo su Discord)",
		})
		return
	}

	// Colori ed emoji per la notifica Discord
	color := 16737876 // #ff6654 Coral (default)
	categoryEmoji := "📡"
	pingContent := ""

	switch req.Category {
	case "COMMISSIONE ART":
		color = 16724582 // #ff3366 Magenta Neon
		categoryEmoji = "💎"
		pingContent = "@here 🚨 **[PRIORITÀ ALTA]** Nuova richiesta di commissione ricevuta!"
	case "COLLABORAZIONE":
		color = 61695 // #00f0ff Ciano Neon
		categoryEmoji = "🤝"
		pingContent = "⚡ **[PROPOSTA]** Nuova collaborazione in arrivo!"
	default:
		req.Category = "SEGNALE GENERICO"
	}

	embed := DiscordEmbed{
		Title: fmt.Sprintf("⚡ Nuova trasmissione da NOVA [%s]", req.Category),
		Color: color,
		Fields: []DiscordField{
			{Name: "👤 Mittente / Callsign", Value: fmt.Sprintf("`%s`", req.Name), Inline: true},
			{Name: "📡 Frequenza / Contatto", Value: fmt.Sprintf("`%s`", req.Email), Inline: true},
			{Name: fmt.Sprintf("%s Protocollo", categoryEmoji), Value: fmt.Sprintf("**%s**", req.Category), Inline: true},
			{Name: "💬 Dati Trasmissione", Value: req.Message, Inline: false},
		},
		Footer: DiscordFooter{
			Text: "NOVA Cyberpunk Terminal • Vercel Go Serverless",
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Errore preparazione payload Discord"})
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Errore chiamata Discord: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Impossibile contattare Discord"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Discord ha risposto con codice %d", resp.StatusCode)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"message": "Trasmissione inoltrata al terminale Discord con successo!",
	})
}
