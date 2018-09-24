package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.in/yaml.v2"
)

func chkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err) //nolint:gas
		os.Exit(1)
	}
}

// Version holds compilation datetime
var Version = ""

func httpHandler(cfg *configStruct) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no bearer token configured, return immediately
		if len(cfg.Bearer) == 0 {
			promhttp.Handler().ServeHTTP(w, r)
			return
		}

		// auth token required
		authHeader := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authHeader) == 2 && authHeader[1] == cfg.Bearer {
			promhttp.Handler().ServeHTTP(w, r)
			return
		}

		// auth token invalid
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized\n")) //nolint:gosec
	})

	return promhttp.Handler()
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: %s config.yml\n", os.Args[0])
		return
	}

	rand.Seed(time.Now().UnixNano())
	log.SetOutput(os.Stderr)
	log.Printf("upmon version %s\n", Version)

	// open cfg file
	f, err := os.Open(os.Args[1])
	chkErr(err)
	defer f.Close()

	// parse config
	var cfg configStruct

	err = yaml.NewDecoder(f).Decode(&cfg)
	chkErr(err)

	if len(cfg.Probes) == 0 {
		chkErr(fmt.Errorf("no probes defined"))
	}

	for _, pcfg := range cfg.Probes {
		newProbe(pcfg)
	}

	// register http handler
	http.Handle("/", httpHandler(&cfg))

	// listen
	if len(cfg.Listen) == 0 {
		cfg.Listen = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	addr := fmt.Sprintf("%s:%d", cfg.Listen, cfg.Port)

	log.Printf("Watching %d probes on %s\n", len(cfg.Probes), addr)
	panic(http.ListenAndServe(addr, nil))
}
