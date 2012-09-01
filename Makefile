GOPATH=$(PWD)/go/

all: go/ brassica-serve bootstrap

go/:
	mkdir -p go/src/datenkarussell.de/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/datenkarussell.de/ ../../../brassica/

.PHONY: brassica-serve
brassica-serve: $(shell find brassica/ -name "*.go")
	go get github.com/hoisie/mustache
	go get launchpad.net/goyaml
	go install datenkarussell.de/brassica/brassica-serve

bootstrap:
	echo "Implement me!"

.PHONY:
clean:
	rm go/ -Rf
