package main

import (
	"net/http"
	"time"
	"github.com/qiu0130/crayon"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	wctx := crayon.Context(r)
	crayon.R200(
		w,
		wctx.Params, // URI parameters
	)
}

func getRoutes() []*crayon.Route {
	return []*crayon.Route{
		&crayon.Route{
			Name:     "helloworld",                   // A label for the API/URI, this is not used anywhere.
			Method:   http.MethodGet,                 // request type
			Pattern:  "/",                            // Pattern for the route
			Handlers: []http.HandlerFunc{helloWorld}, // route handler
		},

	}
}

func main() {
	cfg := crayon.Config{
		Host:         "",
		Port:         "8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	router := crayon.NewRouter(&cfg, getRoutes())
	router.Start()
}