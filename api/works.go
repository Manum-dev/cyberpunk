package handler

import (
	"encoding/json"
	"net/http"
)

// Work rappresenta un'opera del portfolio cyberpunk
type Work struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

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

// Handler per l'endpoint /api/works su Vercel
func HandlerWorks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(works)
}
