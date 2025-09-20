package api

import (
	"net/http"
)

type Route struct {
	Url      string
	Method   string
	Handler  http.Handler
	NeedAuth bool
}

func NewRoute(url string, method string) *Route {
	return &Route{Url: url, Method: method, NeedAuth: true}
}

func NewRouteNotAuth(url string, method string) *Route {
	return &Route{Url: url, Method: method, NeedAuth: false}
}

var routes []*Route

func Routes() []*Route {
	return routes
}

func RegisterRoute[T any](route *Route, function WebHandlerFaction[T]) {
	handler := getHandler(function, route.NeedAuth)
	route.Handler = handler
	routes = append(routes, route)
}
