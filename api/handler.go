package main

import (
	"encoding/json"
	"io"
	"net/http"

	vision "cloud.google.com/go/vision/v2/apiv1"
)

const maxUpload = 20 << 20

type server struct {
	vision *vision.ImageAnnotatorClient
}

type box struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type field struct {
	Name        string `json:"name"`
	MatchedText string `json:"matchedText"`
	Box         box    `json:"box"`
}

type scanResponse struct {
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Deskewed bool    `json:"deskewed"`
	Fields   []field `json:"fields"`
}

func (s *server) scan(w http.ResponseWriter, r *http.Request) {
	if s.vision == nil {
		writeError(w, http.StatusServiceUnavailable, "vision non configuré")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)
	if err := r.ParseMultipartForm(maxUpload); err != nil {
		writeError(w, http.StatusBadRequest, "requête multipart invalide")
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "champ image manquant")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "lecture de l'image impossible")
		return
	}

	straight, width, height, deskewed, err := straighten(raw)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	words, err := detectWords(r.Context(), s.vision, straight)
	if err != nil {
		writeError(w, http.StatusBadGateway, "ocr vision impossible")
		return
	}

	writeJSON(w, http.StatusOK, scanResponse{
		Width:    width,
		Height:   height,
		Deskewed: deskewed,
		Fields:   matchFields(words),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
