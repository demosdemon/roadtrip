package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type cacher interface {
	clearCache()
}

func registerHUP(ctx context.Context, c cacher) {
	ch := make(chan os.Signal, 1) // must not block

	go func() {
		defer signal.Stop(ch)

		for {
			select {
			case s := <-ch:
				log.Printf("%v signal received", s)
				c.clearCache()
				log.Printf("%T cache cleared", c)
			case <-ctx.Done():
				log.Printf("context done: %v", ctx.Err())
				return
			}

		}
	}()

	signal.Notify(ch, syscall.SIGHUP)
}
