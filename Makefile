GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go

all: go/ monsti-contactform

go/:
	mkdir -p go/src/github.com/monsti
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/github.com/monsti ../../../../monsti-contactform/

.PHONY: extract-messages
extract-messages:
	mkdir -p locale/
	find templates/ monsti-contactform/ -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-contactform -L C -p locale/ -kG -kGN:1,2 -

.PHONY: monsti-contactform
monsti-contactform: go/
	$(GO) get github.com/monsti/monsti-contactform/

.PHONY: clean
clean:
	rm go/ -Rf

.PHONY: test
test: go/
	$(GO) test github.com/monsti/monsti-contactform
