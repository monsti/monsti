GOPATH=$(PWD)/go/

all: go/ brassica-serve

go/:
	mkdir -p go/src/datenkarussell.de/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/datenkarussell.de/ ../../../brassica/
	GOPATH=$(GOPATH) go get github.com/hoisie/mustache
	GOPATH=$(GOPATH) go get launchpad.net/goyaml

.PHONY: brassica-serve
brassica-serve: $(shell find brassica/ -name "*.go")
	GOPATH=$(GOPATH) go install datenkarussell.de/brassica/brassica-serve

.PHONY:
clean:
	rm go/ -Rf
