package main

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"

	"github.com/demosdemon/roadtrip/pkg/server"
)

var (
	debug = flag.Bool("debug", false, "enable debugging features")
)

const (
	logFlags = log.LstdFlags | log.Llongfile
)

func main() {
	flag.Parse()
	if *debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	log.SetFlags(logFlags)
	log.SetPrefix("[RoadTrip] ")

	srv := server.Server{
		Context: context.Background(),
		Debug:   *debug,
		Output:  os.Stderr,
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
