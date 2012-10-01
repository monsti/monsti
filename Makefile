GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go

all: dep-epic-editor dep-jquery go/ contactform document monsti

go/:
	mkdir -p go/src/datenkarussell.de/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/datenkarussell.de/ ../../../monsti/
	$(GO) get github.com/drbawb/mustache
#	$(GO) get github.com/hoisie/mustache
	$(GO) get github.com/chrneumann/g5t
	$(GO) get github.com/chrneumann/mimemail
	$(GO) get launchpad.net/goyaml
	$(GO) get code.google.com/p/gorilla/schema
	$(GO) get code.google.com/p/gorilla/sessions
	$(GO) get bitbucket.org/zoowar/bcrypt
	$(GO) get github.com/russross/blackfriday


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
	rm static/epiceditor/ -R


tests: test-worker test-template test-contactform test-markup test-monsti

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

.PHONY: test-markup
test-markup: go/
	$(GO) test datenkarussell.de/monsti/markup

dep-epic-editor: static/epiceditor/
static/epiceditor/:
	wget http://epiceditor.com/docs/downloads/EpicEditor-v0.1.1.1.zip
	unzip EpicEditor-v0.1.1.1.zip
	mv EpicEditor-v0.1.1.1/epiceditor/ static/
	rmdir EpicEditor-v0.1.1.1/
	rm EpicEditor-v0.1.1.1.zip

dep-jquery: static/js/jquery.min.js
static/js/jquery.min.js:
	wget http://code.jquery.com/jquery-1.8.2.min.js
	mv jquery-1.8.2.min.js static/js/jquery.min.js
