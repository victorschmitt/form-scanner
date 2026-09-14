package main

import (
	"context"
	"log"
	"net/http"
	"os"

	vision "cloud.google.com/go/vision/v2/apiv1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	ctx := context.Background()
	s := &server{}
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Printf("Vision indisponible: %v (gcloud auth login --update-adc, puis GOOGLE_CLOUD_PROJECT)", err)
	} else {
		s.vision = client
		defer client.Close()
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Post("/scan", s.scan)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("écoute sur %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
