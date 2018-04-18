package crayon

import (
	"net/http"
	"fmt"
	"encoding/json"
	"log"
	"html/template"
)

const (
	HeaderContentType = "Context-Type"
	JSONContextType = "application/json"
	HTMLContextType = "text/html charset=UTF-8"
	ErrInternalServer = "Internal server error"
)
type ErrorData struct {
	ErrorCode int
	ErrDescription string
}
type dOutput struct {
	Data interface{} `json:"data"`
	Status int
}
type errOutput struct {
	Errors interface{} `json:"errors"`
	Status int
}
type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw responseWriter) Write(data []byte) (int, error) {
	rw.WriteHeader(rw.code)
	return rw.ResponseWriter.Write(data)
}

func (rw responseWriter) WriteHeader(code int) {
	rw.ResponseWriter.Header().Set(HeaderContentType, JSONContextType)
	rw.ResponseWriter.WriteHeader(code)
}

func SendHeader(w http.ResponseWriter, rCode int) {
	w.WriteHeader(rCode)
}
func Send(w http.ResponseWriter, contextType string, data interface{}, rCode int) {
	w.Header().Set(HeaderContentType, contextType)
	w.WriteHeader(rCode)
	_, err := fmt.Fprint(w, data)
	if err != nil {
		R500(w, ErrInternalServer)
	}
}

func SendResponse(w http.ResponseWriter, data interface{}, rCode int) {
	rw := responseWriter{
		ResponseWriter: w,
		code: rCode,
	}
	err := json.NewEncoder(rw).Encode(dOutput{
		Data: data, Status: rCode,
	})
	if err != nil {
		log.Println(err)
		R500(w, ErrInternalServer)
	}
}

func SendError(w http.ResponseWriter, data interface{}, rCode int) {
	rw := responseWriter{
		ResponseWriter: w,
		code: rCode,
	}
	err := json.NewEncoder(rw).Encode(errOutput{data, rCode})
	if err != nil {
		log.Println(err)
		R500(w, ErrInternalServer)
	}
}

func Render(w http.ResponseWriter, data interface{}, rCode int, tpl *template.Template) {
	w.Header().Set(HeaderContentType, HTMLContextType)
	w.WriteHeader(rCode)
	tpl.Execute(w, data)
}

func Render404(w http.ResponseWriter, tpl *template.Template) {
	Render(w, ErrorData {
		404,
		"Sorry, the URL you requested was not found on this server... Or you're lost :-/",
		},
			404,
			tpl,
	)
}

func R200(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, 200)
}
func R201(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, 201)
}
func R204(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, 204)
}
func R302(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, 302)
}
func R400(w http.ResponseWriter, data interface{}) {
	SendResponse(w, data, 400)
}
func R403(w http.ResponseWriter, data interface{}) {
	SendError(w, data, 403)
}
func R404(w http.ResponseWriter, data interface{}) {
	SendError(w, data, 404)
}
func R406(w http.ResponseWriter, data interface{}) {
	SendError(w, data, 406)
}
func R451(w http.ResponseWriter, data interface{}) {
	SendError(w, data, 451)
}
func R500(w http.ResponseWriter, data interface{}) {
	SendError(w, data, 500)
}