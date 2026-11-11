package http

import (
	"fmt"
	"net/http"
	"time"
)

func NewServer(handler http.Handler, port int) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
