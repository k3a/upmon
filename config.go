package main

import (
	"fmt"
	"time"
)

type probeCfg struct {
	URL      string        `yaml:"url"`
	Interval time.Duration `yaml:"interval,omitempty"`
	Timeout  time.Duration `yaml:"timeout,omitempty"`
}

func (pcfg *probeCfg) validate() {
	if len(pcfg.URL) == 0 {
		chkErr(fmt.Errorf("url cannot be empty"))
	}
	if pcfg.Interval == 0 {
		pcfg.Interval = 1 * time.Minute
	}
	if pcfg.Timeout == 0 {
		pcfg.Timeout = 30 * time.Second
	}
}

type configStruct struct {
	Listen string      `yaml:"listen,omitempty"`
	Port   int         `yaml:"port,omitempty"`
	Bearer string      `yaml:"bearer"` // authentication bearer
	Probes []*probeCfg `yaml:"probes"`
}
