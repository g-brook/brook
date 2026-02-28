/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/g-brook/brook/common/configs"
	"github.com/g-brook/brook/common/log"
	"github.com/g-brook/brook/common/threading"
	"github.com/g-brook/brook/scmd/web/api"
	"github.com/g-brook/brook/scmd/web/db"
	"github.com/g-brook/brook/scmd/web/sql"
	"github.com/gorilla/mux"
)

//go:embed dist
var embeddedFiles embed.FS
var (
	root = "dist"
)

type Server struct {
}

func NewWebServer(port int) {
	if port <= 4000 || port > 9000 {
		log.Info("port is invalid %d, use default port: 8000", port)
		port = configs.DefWebPort
	}
	doRoute()
	db.Open()
	err := sql.InitSQLDB()
	if err != nil {
		log.Error("init sql db err %v", err)
		return
	}
	//init db check
	err = sql.CheckInfoDB()
	if err != nil {
		log.Error("init sql db err %v", err)
		return
	}
	threading.GoSafe(func() {
		log.Info("start web server on port %d", port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			panic("start web server err: " + err.Error())
		}
	})
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
	// static source
	r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(staticFs)))
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//http.ServeFile(w, r, filepath.Join(root, "index.html"))
		// 直接从 embed 中读取
		file, err := fs.ReadFile(staticFs, "index.html")
		if err != nil {
			http.Error(w, "index not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(file)
	})
	http.Handle("/", r)
}

func Close() {
	db.Close()
}
