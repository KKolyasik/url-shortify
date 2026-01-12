package main

import (
	"log"
	"net/http"

	"github.com/KKolyasik/url-shortify/internal/handler"
	"github.com/KKolyasik/url-shortify/internal/service"
	"github.com/KKolyasik/url-shortify/internal/storage"
)

func main() {
	st := storage.NewMemoryStore()
	svc := service.New(st)
	h := handler.New("http://localhost:8080", svc)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			h.Shortify(w, r)
			return
		}
		h.Redirect(w, r)
	})

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
