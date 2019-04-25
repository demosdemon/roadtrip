package main

import (
	"log"
	"net/http"
	"os"

	"github.com/demosdemon/roadtrip/pkg/server"
)

func main() {
	srv := server.Server{
		Output: os.Stderr,
	}

	handler, err := srv.Handler()
	if err != nil {
		panic(err)
	}

	l, err := srv.Listener()
	if err != nil {
		panic(err)
	}

	addr := l.Addr()
	log.Printf("listening on %s:%s", addr.Network(), addr.String())

	_ = http.Serve(l, handler)
}
