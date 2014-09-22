package viewer

import (
	"html/template"
	"net/http"
)

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/view", func(rw http.ResponseWriter, req *http.Request) {
		source := req.FormValue("doc")
		tmpl.Execute(rw, source)
	})

	http.Handle("/", http.FileServer(http.Dir(".")))
}
