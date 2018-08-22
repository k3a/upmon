build:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure
	go build -ldflags "-X main.Version=$(shell git rev-parse --short HEAD)@$(shell date -I)"
