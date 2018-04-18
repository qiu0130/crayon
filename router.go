package crayon

import (
	"net/http"
	"regexp"
	"fmt"
	"strings"
	"log"
	"context"
)

const (
	URLCHARS = `([^/]+)`
	URLWILDCARD = `(.*)`

)

var validHttpMethods = []string{
	http.MethodOptions,
	http.MethodHead,
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

type customResponseWriter struct {
	http.ResponseWriter
	Statuscode int
	flag bool
}

func (crw *customResponseWriter) WriteHeader(code int) {
	if crw.flag {
		return
	}
	crw.Statuscode = code
	crw.ResponseWriter.WriteHeader(code)
}

func (crw *customResponseWriter) Write(body []byte) (int, error) {
	if crw.flag {
		return 0, nil
	}
	crw.flag = true
	return crw.Write(body)
}

type Route struct {
	Name string
	Method string
	Pattern string
	TrailingSlash bool
	FallTroughPostResponse bool
	Handlers []http.HandlerFunc
	keys []string
	PatternString string
	PatternRegexp *regexp.Regexp
}

func (r *Route) extractParameter(patternString string, hasWildcard bool, key string) string {
	regexPattern := ""
	patternKey := ""

	if hasWildcard {
		patternKey = fmt.Sprintf(":%s*", key)
		regexPattern = URLWILDCARD
	} else {
		patternKey = fmt.Sprintf(":%s", key)
		regexPattern = URLCHARS
	}
	patternString = strings.Replace(patternString, patternKey, regexPattern, 1)
	for idx, k := range r.keys {
		if key == k {
			log.Fatal(errDuplicateKey, "\nURI: ", r.Pattern, "\nKey:", k, ", Position:", idx+1)
		}
	}
	r.keys = append(r.keys, key)
	return patternString
}

func (r * Route) init() error {
	patternString := r.Pattern

	if strings.Contains(r.Pattern, ":") {
		key := ""
		hasKey := false
		hasWildcard := false

		for i, n := 0, len(r.Pattern); i < n; i++ {
			char := string(r.Pattern[i])
			if char == ":" {
				hasKey = true
			} else if char == "*" {
				hasWildcard = true
			} else if hasKey && char != "/" {
				key += char
			} else if hasKey && len(key) > 0 {
				patternString = r.extractParameter(patternString, hasWildcard, key)
				hasWildcard, hasKey, key = false, false, ""
			}

		}
		if hasKey && len(key) > 0 {
			patternString = r.extractParameter(patternString, hasWildcard, key)

		}
	}

	if r.TrailingSlash {
		patternString = fmt.Sprintf("^%s%s$", patternString, )
	} else {
		patternString = fmt.Sprintf("^%s$", patternString)
	}
	patternRegexp, err := regexp.Compile(patternString)
	if err != nil {
		return err
	}

	r.PatternRegexp = patternRegexp
	r.PatternString = patternString

	return nil

}

func (r *Route) matchURI(requestURI string) (bool, map[string]string) {
	if r.Pattern == requestURI {
		return true, nil
	}

	if !r.PatternRegexp.Match([]byte(requestURI)) {
		return false, nil
	}
	values := r.PatternRegexp.FindStringSubmatch(requestURI)
	n := len(values)
	if n == 0 {
		return true, nil
	}
	uriValues := make(map[string]string, n-1)
	for i := 1; i < n; i++ {
		uriValues[r.keys[i-1]] = values[i]
	}
	return true, uriValues
}


type Router struct {
	Options []*Route
	Head []*Route
	Get []*Route
	Post []*Route
	Put []*Route
	Patch []*Route
	Delete []*Route

	NotFound http.HandlerFunc
	AppContext map[string]interface{}
	config *Config
	serveHandler http.HandlerFunc
	httpServer *http.Server
	httpsServer *http.Server
}

func NewRouter(cfg *Config, routes []*Route) *Router {
	methods := make(map[string][]*Route, len(validHttpMethods))

	for _, method := range validHttpMethods {
		methods[method] = []*Route{}
	}

	for idx, route := range routes {
		found := false
		for _, method := range validHttpMethods {
			if route.Method == method {
				found = true
			}
		}
		if !found {
			log.Fatalln("Unsupported HTTP request method provided. Method:", route.Method)
		}
		if route.Handlers == nil || len(route.Handlers) == 0 {
			log.Fatalln("No handlers provided for the route '", route.Pattern, "', method '", route.Method, "'")
		}

		if err := route.init(); err != nil {
			log.Fatalln("Unsupported URI pattern.", route.Pattern, err)
		}
		// TODO ??
		for i := 0; i < idx; i++ {
			rt := routes[i]
			if rt.Name == route.Name {
				log.Println("Duplicate route name(\"" + rt.Name + "\") detected. Route name should be unique.")
			}
			if rt.Method == route.Method {
				if ok, _ := rt.matchURI(route.Pattern); ok {
					log.Println("Duplicate URI pattern detected.\nPattern: '" + rt.Pattern + "'\nDuplicate pattern: '" + route.Pattern + "'")
				}
			}
		}
		methods[route.Method] = append(methods[route.Method], route)
	}

	r := &Router{
		Options: methods[http.MethodOptions],
		Head: methods[http.MethodHead],
		Get: methods[http.MethodGet],
		Post: methods[http.MethodPost],
		Put: methods[http.MethodPut],
		Patch: methods[http.MethodPatch],
		Delete: methods[http.MethodDelete],

		NotFound: http.NotFound,
		AppContext: make(map[string]interface{}, 0),
		config: cfg,
	}
	r.serveHandler = r.serve

	return r
}

func (rtr *Router) serve(rw http.ResponseWriter, req *http.Request) {
	var rr []*Route

	switch req.Method {
	case http.MethodOptions:
		rr = rtr.Options
	case http.MethodHead:
		rr = rtr.Head
	case http.MethodGet:
		rr = rtr.Get
	case http.MethodPost:
		rr = rtr.Post
	case http.MethodPut:
		rr = rtr.Put
	case http.MethodPatch:
		rr = rtr.Patch
	case http.MethodDelete:
		rr = rtr.Delete
	default:
		return
	}

	var route *Route
	ok := false
	params := make(map[string]string, 0)
	path := req.URL.EscapedPath()
	for _, r := range rr {
		if ok, params = r.matchURI(path); ok {
			route = r
			break
		}
	}
	if !ok {
		rtr.NotFound(rw, req)
		return
	}
	crw := &customResponseWriter{
		ResponseWriter: rw,
	}
	reqContext := req.WithContext(
		context.WithValue(
			req.Context(),
			crayonContextKey,
			&CrayonContext{
				Params: params,
				Route: route,
				AppContext: rtr.AppContext,
			},
		),
	)


	for _, handler := range route.Handlers {
		if crw.flag == false {
			handler(crw, reqContext)
		} else if route.FallTroughPostResponse {
			handler(crw, reqContext)
		} else {
			break
		}
	}
}

func (rtr *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rtr.serveHandler(rw, req)
}

func (rtr *Router) AddMiddleware(f func(w http.ResponseWriter, r *http.Request, handlerFunc http.HandlerFunc)) {
	srv := rtr.serveHandler
	rtr.serveHandler = func(rw http.ResponseWriter, req *http.Request) {
		f(rw, req, srv)
	}
}