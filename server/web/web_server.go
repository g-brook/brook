package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/brook/common/log"
	"github.com/brook/server/web/api"
	"github.com/brook/server/web/db"
	"github.com/gorilla/mux"
)

//go:embed dist/*
var embeddedFiles embed.FS
var (
	root = "dist"
)

type Server struct {
}

func NewWebServer(port int) {
	//staticFs, _ := fs.Sub(embeddedFiles, root)
	//http.Handle("/", http.FileServer(http.FS(staticFs)))
	doRoute()
	db.Open()
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		return
	}
}

func doRoute() {
	staticFs, _ := fs.Sub(embeddedFiles, root)
	r := mux.NewRouter()
	// api source
	routes := api.Routes()
	apiRouter := r.PathPrefix("/api").Subrouter()
	for _, item := range routes {
		apiRouter.Handle(item.Url, item.Handler).Methods(item.Method)
		log.Debug("register route: %s %s", item.Method, "/api"+item.Url)
	}
	//static source
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticFs)))
	http.Handle("/", r)
}
