package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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
	router.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	router.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/robots.txt")
	})

	// Homepage and catch-all.
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", nil)
	})

	log.Printf("Running server on http://localhost:%s", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), router)
}
