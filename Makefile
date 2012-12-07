GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
ALOHA_VERSION=0.22.3

all: dep-aloha-editor dep-jquery monsti node-types bcrypt

.PHONY: extract-messages
extract-messages:
	mkdir -p locale/
	find templates/ -name "*.html" | xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-cms -L C -p locale/ -kG -kGN:1,2 -

.PHONY: bcrypt
bcrypt: 
	$(GO) get github.com/monsti/monsti-daemon/tools/bcrypt

.PHONY: monsti
monsti:
	$(GO) get github.com/monsti/monsti-daemon

.PHONY: node-types
node-types:
	$(GO) get github.com/monsti/monsti-contactform
	$(GO) get github.com/monsti/monsti-document
	$(GO) get github.com/monsti/monsti-image

.PHONY: tests
tests:
	$(GO) test github.com/monsti/monsti-daemon
	$(GO) test github.com/monsti/monsti-daemon/worker
	$(GO) test github.com/monsti/monsti-contactform
	$(GO) test github.com/monsti/monsti-document
	$(GO) test github.com/monsti/monsti-image

.PHONY: clean
clean:
	rm go/bin/*
	rm go/pkg/* -R
	rm static/aloha/ -R

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
