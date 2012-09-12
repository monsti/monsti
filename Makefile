GOPATH=$(PWD)/go/

all: go/ contactform document monsti locales

go/:
	mkdir -p go/src/datenkarussell.de/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/datenkarussell.de/ ../../../monsti/
	GOPATH=$(GOPATH) go get github.com/drbawb/mustache
#	GOPATH=$(GOPATH) go get github.com/hoisie/mustache
	GOPATH=$(GOPATH) go get launchpad.net/goyaml
	GOPATH=$(GOPATH) go get code.google.com/p/gorilla/schema

.PHONY: locales
locales:
	mkdir -p locale/
	find templates/ -name "*.html"| xargs cat \
	  | sed 's|{{#_}}\(.*\){{/_}}|gettext("\1");|g' \
	  | xgettext -d monsti -L C -p locale/ -

.PHONY: monsti
monsti: go/
	GOPATH=$(GOPATH) go install datenkarussell.de/monsti

.PHONY: document
document: go/
	GOPATH=$(GOPATH) go install datenkarussell.de/monsti/node/document

.PHONY: contactform
contactform: go/
	GOPATH=$(GOPATH) go install datenkarussell.de/monsti/node/contactform

.PHONY: clean
clean:
	rm go/ -Rf

.PHONY: tests
tests:
	GOPATH=$(GOPATH) go test datenkarussell.de/monsti/
