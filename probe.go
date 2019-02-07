package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var errCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "upmon_errors_total",
		Help: "Number of failed http requests",
	},
	[]string{"url", "error"},
)

var reqSummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "upmon_milliseconds",
		Help: "Response time",
	},
	[]string{"url"},
)

type probe struct {
	client *http.Client
	ticker *time.Ticker

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
			errCounter.WithLabelValues(p.cfg.URL, "request").Inc()
			log.Printf("Error making request to %s: %v\n", p.cfg.URL, err)
			continue
		}
		respTimeMsec := time.Since(startTime) / time.Millisecond
		r.Body.Close()

		if r.ContentLength == 0 {
			errCounter.WithLabelValues(p.cfg.URL, "zero-len").Inc()
		} else if r.StatusCode != 200 {
			errCounter.WithLabelValues(p.cfg.URL, "http-"+strconv.Itoa(r.StatusCode)).Inc()
		} else {
			reqSummary.WithLabelValues(p.cfg.URL).Observe(float64(respTimeMsec))
		}
	}
}

func newProbe(pcfg *probeCfg) *probe {
	pcfg.validate()

	ticker := time.NewTicker(pcfg.Interval)

	p := &probe{
		&http.Client{
			Timeout: pcfg.Timeout,
		},
		ticker,
		pcfg,
	}

	go p.run()

	return p
}

func init() {
	chkErr(prometheus.Register(errCounter))
	chkErr(prometheus.Register(reqSummary))
}
