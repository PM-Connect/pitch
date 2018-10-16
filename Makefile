.PHONY: build run-dev build-docker clean

GIT_BRANCH := $(subst heads/,,$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
DEV_IMAGE := pitch-dev$(if $(GIT_BRANCH),:$(subst /,-,$(GIT_BRANCH)))
DEV_DOCKER_IMAGE := pitch-bin-dev$(if $(GIT_BRANCH),:$(subst /,-,$(GIT_BRANCH)))

default: clean install coverage crossbinary

clean:
	rm -rf dist/

binary: install
	GOOS=linux CGO_ENABLED=0 GOGC=off GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o "$(CURDIR)/dist/pitch"

crossbinary: binary
	GOOS=linux GOARCH=amd64 go build -o "$(CURDIR)/dist/pitch-linux-amd64"
	GOOS=linux GOARCH=386 go build -o "$(CURDIR)/dist/pitch-linux-386"
	GOOS=darwin GOARCH=amd64 go build -o "$(CURDIR)/dist/pitch-darwin-amd64"
	GOOS=darwin GOARCH=386 go build -o "$(CURDIR)/dist/pitch-darwin-386"
	GOOS=windows GOARCH=amd64 go build -o "$(CURDIR)/dist/pitch-windows-amd64.exe"
	GOOS=windows GOARCH=386 go build -o "$(CURDIR)/dist/pitch-windows-386.exe"

install: clean
	go mod vendor
	go generate

test:
	go test ./...

coverage:
	"$(CURDIR)/script/coverage.sh"

dist:
	mkdir dist

run-dev:
	go generate
	go test ./...
	go build -o "pitch"
	./pitch

build-docker:
	docker build -t "$(DEV_DOCKER_IMAGE)" .
