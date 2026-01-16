package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// GetTemplates returns a pointer of all available templates.
func GetTemplates() *template.Template {
	t := template.New("main")

	_, err := t.ParseGlob("views/*.html")
	if err != nil {
		panic(fmt.Sprintf("error loading templates: %s", err.Error()))
	}

	return t
}

func main() {
	router := http.NewServeMux()
	templates := GetTemplates()
	fs := http.FileServer(http.Dir("./assets"))
	port := "8080"

	// Assets such as images and stylesheets.
	router.Handle("GET /assets/", http.StripPrefix("/assets/", cacheControlHandler(fs)))

	router.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/robots.txt")
	})

	// Health check endpoint for deployment platforms
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Speed text reader page
	router.HandleFunc("GET /speed-text", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "speed-text.html", nil)
	})

	// Homepage and catch-all.
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", nil)
	})

	log.Printf("Running server on http://localhost:%s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}

// cacheControlHandler adds cache control headers to non-JS and non-CSS files
func cacheControlHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the file extension
		if !strings.HasSuffix(r.URL.Path, ".js") && !strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}

		next.ServeHTTP(w, r)
	})
}
