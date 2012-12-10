GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
ALOHA_VERSION=0.22.3

NODE_TYPES = document contactform image

all: dep-aloha-editor dep-jquery monsti node-types bcrypt

.PHONY: bcrypt
bcrypt: 
	$(GO) get github.com/monsti/monsti-daemon/tools/bcrypt

.PHONY: monsti
monsti:
	$(GO) get github.com/monsti/monsti-daemon

node-types: $(NODE_TYPES)
$(NODE_TYPES): %: ext/% go/bin/monsti-%

ext/%:
	mkdir -p ext/
	wget https://github.com/monsti/monsti-$*/archive/master.tar.gz -O ext/$*.tar.gz
	cd ext; tar xvf $*.tar.gz && mv monsti-$*-master $* && rm $*.tar.gz
	mkdir -p go/src/github.com/monsti/
	ln -s ../../../../ext/$* go/src/github.com/monsti/monsti-$*

go/bin/monsti-%:
	$(GO) install github.com/monsti/monsti-$*

.PHONY: tests
tests: $(NODE_TYPES:%=test-ext-%) test-daemon daemon/test-worker

test-ext-% test-%:
	$(GO) test github.com/monsti/monsti-$*

.PHONY: clean
clean:
	rm go/bin/* -Rf
	rm go/pkg/* -Rf
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
