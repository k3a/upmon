build:
	go build -ldflags "-X main.buildstamp=$(shell git rev-parse --short HEAD)@$(shell date -I)"
