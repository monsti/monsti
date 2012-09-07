GOPATH=$(PWD)/go/

all: go/ document monsti

go/:
	mkdir -p go/src/datenkarussell.de/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/datenkarussell.de/ ../../../monsti/
	GOPATH=$(GOPATH) go get github.com/hoisie/mustache
	GOPATH=$(GOPATH) go get launchpad.net/goyaml
	GOPATH=$(GOPATH) go get code.google.com/p/gorilla/schema

.PHONY: monsti
monsti: go/
	GOPATH=$(GOPATH) go install datenkarussell.de/monsti

.PHONY: document
document: go/
	GOPATH=$(GOPATH) go install datenkarussell.de/monsti/node/document

.PHONY: clean
clean:
	rm go/ -Rf

.PHONY: tests
tests:
	GOPATH=$(GOPATH) go test datenkarussell.de/monsti/
