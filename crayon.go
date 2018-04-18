package crayon

import (
	"net/http"
	"log"
	"crypto/tls"
	"context"
)

const crayonContextKey  = "crayon_context_key"

type CrayonContext struct {
	Params map[string]string
	Route *Route
	AppContext map[string]interface{}
}

func Context(r *http.Request) *CrayonContext {
	cc, _ := r.Context().Value(crayonContextKey).(*CrayonContext)
	return cc
}

func (router *Router) StartHTTPS() {

	cfg := router.config
	if cfg.CertFile == "" {
		log.Fatalln("No certificate provided for HTTPS")
	}
	if cfg.KeyFile == "" {
		log.Fatalln("No key file provided for HTTPS")
	}
	host := cfg.Host
	if len(cfg.HTTPSPort) > 0 {
		host += ":" + cfg.HTTPSPort
	}

	router.httpsServer = &http.Server{
		Addr: host,
		Handler: router,
		ReadTimeout: cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			},
	}
	log.Println("HTTPS server, listening on", host)
	err := router.httpsServer.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	if err != nil && err != http.ErrServerClosed {
		log.Println("HTTPS server exited with error:", err.Error())
	}
}

func (router *Router) Start() {
	cfg := router.config
	host := cfg.Host

	if len(cfg.Port) > 0 {
		host += ":" +  cfg.Port
	}
	router.httpServer = &http.Server{
		Addr: host,
		Handler: router,
		ReadTimeout: cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	log.Println("HTTP server listening on", host)
	err := router.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Println("HTTP server exited with error", err.Error())
	}
}

func (router *Router) Shutdown() error {
	if router.httpServer == nil {
		return nil
	}
	timer := router.config.ShutdownTimeout
	ctx, cancel := context.WithTimeout(context.Background(), timer)
	defer cancel()

	err := router.httpServer.Shutdown(ctx)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (router *Router) ShutdownHTTPS() error {
	if router.httpServer == nil {
		return nil
	}
	timer := router.config.ShutdownTimeout
	ctx, cancel := context.WithTimeout(context.Background(), timer)
	defer cancel()

	err := router.httpsServer.Shutdown(ctx)
	if err != nil && err != http.ErrServerClosed {
		log.Println(err)
	}
	return err
}
