package main

import (
	"net/http"
	_ "net/http/pprof"
)

func startPprof() {
	if cfg.AddrPprof == "" {
		return
	}

	logger.Printf("starting pprof http server at %s/debug/pprof", cfg.AddrPprof)

	go func() {
		logger.Printf("error running pprof: %s", http.ListenAndServe(cfg.AddrPprof, nil))
	}()
}
