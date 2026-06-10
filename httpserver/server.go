// Package httpserver constructs and starts the HTTP server for an Adele application.
//
// It builds a net/http.Server configured from the application's environment,
// routes, and logger, then listens for and serves incoming connections.
package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cidekar/adele-framework"
	"github.com/sirupsen/logrus"
)

// Create a new http server for use with the adele skeleton application.
func NewServer(adele *adele.Adele) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%s", adele.Helpers.Getenv("HTTP_PORT", "4000")),
		ErrorLog:     log.New(adele.Log.WriterLevel(logrus.ErrorLevel), "", 0),
		Handler:      adele.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}
}

// Creates a new http server, listens on the TCP network address srv.Addr and then calls
// server to handle requests on incoming connections. Accepted connections are configured
// to enable TCP keep-alives.
func Start(adele *adele.Adele) error {
	server := NewServer(adele)
	return server.ListenAndServe()
}
