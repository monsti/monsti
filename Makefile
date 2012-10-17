GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
ALOHA_VERSION=0.22.1

all: dep-aloha-editor dep-jquery go/ contactform document monsti

go/:
	mkdir -p go/src/datenkarussell.de/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/datenkarussell.de/ ../../../monsti/
	$(GO) get github.com/chrneumann/g5t
	$(GO) get github.com/chrneumann/mimemail
	$(GO) get launchpad.net/goyaml
	$(GO) get code.google.com/p/gorilla/schema
	$(GO) get code.google.com/p/gorilla/sessions
#	$(GO) get bitbucket.org/zoowar/bcrypt

.PHONY: extract-messages
extract-messages:
	mkdir -p locale/
	find templates/ monsti/ -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{#_}}\(.*\){{/_}}|gettext("\1");|g' \
	  | xgettext -d monsti -L C -p locale/ -kG -kGN:1,2 -

.PHONY: monsti
monsti: go/
	$(GO) install datenkarussell.de/monsti

.PHONY: document
document: go/
	$(GO) install datenkarussell.de/monsti/node/document

.PHONY: contactform
contactform: go/
	$(GO) install datenkarussell.de/monsti/node/contactform

.PHONY: clean
clean:
	rm go/ -Rf
	rm static/aloha/ -R

tests: test-worker test-template test-form test-contactform test-monsti

.PHONY: test-monsti
test-monsti: go/
	$(GO) test datenkarussell.de/monsti

.PHONY: test-worker
test-worker: go/
	$(GO) test datenkarussell.de/monsti/worker

.PHONY: test-contactform
test-contactform: go/
	$(GO) test datenkarussell.de/monsti/node/contactform

.PHONY: test-template
test-template: go/
	$(GO) test datenkarussell.de/monsti/template

.PHONY: test-form
test-form: go/
	$(GO) test datenkarussell.de/monsti/form

dep-aloha-editor: static/aloha/
static/aloha/:
	wget https://github.com/downloads/alohaeditor/Aloha-Editor/alohaeditor-$(ALOHA_VERSION).zip
	unzip alohaeditor-$(ALOHA_VERSION).zip
	mkdir static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/lib static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/css static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/img static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/plugins static/aloha
	rm alohaeditor-$(ALOHA_VERSION) -R
	rm alohaeditor-$(ALOHA_VERSION).zip

dep-jquery: static/js/jquery.min.js
static/js/jquery.min.js:
	wget http://code.jquery.com/jquery-1.8.2.min.js
	mv jquery-1.8.2.min.js static/js/jquery.min.js
