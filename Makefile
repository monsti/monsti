GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
ALOHA_VERSION=0.22.3

all: dep-aloha-editor dep-jquery go/ document monsti bcrypt

go/:
	mkdir -p go/src/github.com/monsti/
	mkdir -p go/bin
	mkdir -p go/pkg
	ln -s -t go/src/github.com/monsti/ ../../../../monsti/

.PHONY: extract-messages
extract-messages:
	mkdir -p locale/
	find templates/ monsti/ -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti -L C -p locale/ -kG -kGN:1,2 -

.PHONY: bcrypt
bcrypt: go/
	$(GO) get github.com/monsti/monsti/tools/bcrypt

.PHONY: monsti
monsti: go/
	$(GO) get github.com/monsti/monsti

.PHONY: document
document: go/
	$(GO) get github.com/monsti/monsti/node/document

.PHONY: clean
clean:
	rm go/ -Rf
	rm static/aloha/ -R

tests: test-worker test-monsti

.PHONY: test-monsti
test-monsti: go/
	$(GO) test github.com/monsti/monsti

.PHONY: test-worker
test-worker: go/
	$(GO) test github.com/monsti/monsti/worker

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
