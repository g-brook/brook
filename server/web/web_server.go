package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var embeddedFiles embed.FS
var (
	root = "dist"
)

type Server struct {
}

func NewWebServer(port int) {
	//readPages()
	staticFs, _ := fs.Sub(embeddedFiles, root)
	http.Handle("/", http.FileServer(http.FS(staticFs)))
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		return
	}
}
