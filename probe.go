package main

import (
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type probe struct {
	client  *http.Client
	ticker  *time.Ticker
	counter *prometheus.CounterVec
	hist    prometheus.Histogram

	cfg *probeCfg
}

func (p *probe) run() {
	for {
		<-p.ticker.C

		// slightly randomize probes
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		startTime := time.Now()
		r, err := p.client.Get(p.cfg.URL)
		if err != nil {
			p.counter.WithLabelValues("request").Inc()
			log.Printf("Error making request to %s: %v\n", p.cfg.URL, err)
			continue
		}
		respTimeMsec := time.Since(startTime) / time.Millisecond
		defer r.Body.Close()

		if r.ContentLength == 0 {
			p.counter.WithLabelValues("zero-len").Inc()
		} else if r.StatusCode != 200 {
			p.counter.WithLabelValues("http-" + strconv.Itoa(r.StatusCode)).Inc()
		} else {
			p.hist.Observe(float64(respTimeMsec))
		}
	}
}

var safeRE = regexp.MustCompile(`[^a-z0-9_]`)

func counterIdentForURL(url string) string {
	url = strings.ToLower(url)
	return safeRE.ReplaceAllString(url, "_")
}

func newProbe(pcfg *probeCfg) *probe {
	pcfg.validate()

	ticker := time.NewTicker(pcfg.Interval)

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: counterIdentForURL(pcfg.URL) + "_errors_total",
			Help: pcfg.URL,
		},
		[]string{"error"},
	)
	chkErr(prometheus.Register(counter))

	histogram := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: counterIdentForURL(pcfg.URL) + "_milliseconds",
			Help: pcfg.URL,
		},
	)
	chkErr(prometheus.Register(histogram))

	p := &probe{
		&http.Client{
			Timeout: pcfg.Timeout,
		},
		ticker,
		counter,
		histogram,
		pcfg,
	}

	go p.run()

	return p
}
