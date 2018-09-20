package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
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

	// register promhttp
	http.Handle("/", promhttp.Handler())
	http.Handle("/upmon", promhttp.Handler())

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
