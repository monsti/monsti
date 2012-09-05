GOPATH=$(PWD)/go/

all: go/ monsti-serve

go/:
	mkdir -p go/src/datenkarussell.de/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/datenkarussell.de/ ../../../monsti/
	GOPATH=$(GOPATH) go get github.com/hoisie/mustache
	GOPATH=$(GOPATH) go get launchpad.net/goyaml
	GOPATH=$(GOPATH) go get code.google.com/p/gorilla/schema

.PHONY:
monsti-serve: $(shell find monsti/ -name "*.go")
	GOPATH=$(GOPATH) go install datenkarussell.de/monsti/monsti-serve

.PHONY:
clean:
	rm go/ -Rf

.PHONY:
tests:
	GOPATH=$(GOPATH) go test datenkarussell.de/monsti/
