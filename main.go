package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/muniverse"
)

type Server struct {
	ListenAddr string

	AssetDir    string
	TemplateDir string

	Templates *template.Template
}

func main() {
	var server Server
	flag.StringVar(&server.ListenAddr, "addr", ":8080", "address to listen on")
	flag.StringVar(&server.AssetDir, "assets", "assets", "asset directory path")
	flag.StringVar(&server.TemplateDir, "templates", "templates", "template directory path")
	flag.Parse()

	http.HandleFunc("/", server.HandleRoot)
	http.HandleFunc("/play/", server.HandlePlay)
	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir(server.AssetDir))))

	if err := http.ListenAndServe(server.ListenAddr, nil); err != nil {
		essentials.Die(err)
	}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	var items []map[string]interface{}
	for _, spec := range muniverse.EnvSpecs {
		items = append(items, map[string]interface{}{
			"name": spec.Name,
		})
	}
	s.serveTemplate(w, "index", items)
}

func (s *Server) HandlePlay(w http.ResponseWriter, r *http.Request) {
	expr := regexp.MustCompile("^/play/([A-Za-z0-9-]*)/$")
	submatch := expr.FindStringSubmatch(r.URL.Path)
	if submatch == nil {
		http.NotFound(w, r)
		return
	}
	spec := muniverse.SpecForName(submatch[1])
	if spec == nil {
		http.NotFound(w, r)
		return
	}
	item := map[string]interface{}{
		"name": spec.Name,
	}
	s.serveTemplate(w, "play", item)
}

func (s *Server) serveTemplate(w http.ResponseWriter, name string, data interface{}) {
	filename := filepath.Join(s.TemplateDir, name+".html")
	template, err := template.ParseFiles(filename)
	if err != nil {
		log.Println("load template "+name+":", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := template.Execute(w, data); err != nil {
		log.Println("execute template "+name+":", err)
	}
}
