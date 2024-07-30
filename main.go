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

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		templates.ExecuteTemplate(w, "index.html", nil)
	})

	router.Handle("GET /assets/", http.StripPrefix("/assets/", fs))

	log.Println("Running server on http://localhost:8080")
	http.ListenAndServe(":8080", router)
}
