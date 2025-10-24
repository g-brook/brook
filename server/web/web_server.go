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
	"math/rand"
	"net/http"
	"time"

	"github.com/brook/common/log"
	"github.com/brook/server/web/api"
	"github.com/brook/server/web/db"
	"github.com/brook/server/web/sql"
	"github.com/gorilla/mux"
)

//go:embed dist/*
var embeddedFiles embed.FS
var (
	root    = "dist"
	charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNPQRSTUVWXYZ123456789!@#$%^&*()"
)

type Server struct {
}

func NewWebServer(port int) {
	if port <= 0 || port > 30000 {
		log.Info("port is invalid %d, use default port: 8812", port)
		port = 8000
	}
	doRoute()
	db.Open()
	err := sql.InitSQLDB()
	if err != nil {
		log.Error("init sql db err %v", err)
		return
	}
	go func() {
		log.Info("start web server on port %d", port)
		err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			panic("start web server err: " + err.Error())
		}
	}()
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

func randomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}
