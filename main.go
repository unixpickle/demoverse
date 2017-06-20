package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/muniverse"
)

type Server struct {
	ListenAddr string

	AssetDir    string
	TemplateDir string
	OutputDir   string

	FrameTime time.Duration
	Cursor    bool
	Filter    EventFilter

	Templates *template.Template
}

func main() {
	var server Server
	flag.StringVar(&server.ListenAddr, "addr", ":8080", "address to listen on")
	flag.StringVar(&server.AssetDir, "assets", "assets", "asset directory path")
	flag.StringVar(&server.TemplateDir, "templates", "templates", "template directory path")
	flag.StringVar(&server.OutputDir, "outdir", "recordings", "recordings directory")
	flag.DurationVar(&server.FrameTime, "frametime", time.Second/10, "time per frame")
	flag.BoolVar(&server.Cursor, "cursor", false, "render cursor")
	flag.Var(&server.Filter, "filter", "event filter (NoFilter or DeltaFilter)")
	flag.Parse()

	if info, err := os.Stat(server.OutputDir); os.IsNotExist(err) {
		if err := os.Mkdir(server.OutputDir, 0755); err != nil {
			essentials.Die(err)
		}
		log.Println("created directory:", server.OutputDir)
	} else if err != nil {
		essentials.Die(err)
	} else if !info.IsDir() {
		essentials.Die("not a directory:", server.OutputDir)
	}

	http.HandleFunc("/", server.HandleRoot)
	http.HandleFunc("/play/", server.HandlePlay)
	http.HandleFunc("/env/", server.HandleEnv)
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
	sort.Slice(items, func(i int, j int) bool {
		iName := items[i]["name"].(string)
		jName := items[j]["name"].(string)
		return strings.Compare(iName, jName) < 0
	})
	s.serveTemplate(w, "index", items)
}

func (s *Server) HandlePlay(w http.ResponseWriter, r *http.Request) {
	specName := s.pathEnvName(r)
	spec := muniverse.SpecForName(specName)
	if spec == nil {
		http.NotFound(w, r)
		return
	}
	item := map[string]interface{}{
		"name":     spec.Name,
		"width":    spec.Width,
		"height":   spec.Height,
		"keys":     spec.KeyWhitelist,
		"interval": int(s.FrameTime / time.Millisecond),
	}
	s.serveTemplate(w, "play", item)
}

func (s *Server) HandleEnv(w http.ResponseWriter, r *http.Request) {
	envName := s.pathEnvName(r)
	spec := muniverse.SpecForName(envName)
	if spec == nil {
		http.NotFound(w, r)
		return
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade:", err)
		return
	}
	handler := &EnvHandler{
		Server: s,
		Conn:   conn,
		Spec:   spec,
	}
	log.Println("env handler:", handler.Handle())
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

func (s *Server) pathEnvName(r *http.Request) string {
	expr := regexp.MustCompile("^/[a-z]*/([A-Za-z0-9-]*)/?$")
	submatch := expr.FindStringSubmatch(r.URL.Path)
	if submatch == nil {
		return ""
	}
	return submatch[1]
}
