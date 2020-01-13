build:
	go build -ldflags "-X main.Version=$(shell git rev-parse --short HEAD)@$(shell date -I)"
