GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
ALOHA_VERSION=0.22.3

MODULES = daemon document contactform image

all: monsti bcrypt

monsti: dep-aloha-editor dep-jquery modules

.PHONY: bcrypt
bcrypt: 
	$(GO) get github.com/monsti/monsti-daemon/tools/bcrypt

modules: $(MODULES)
$(MODULES): %: go/bin/monsti-%

# Fetch and setup given module
module/%:
	mkdir -p module/
	wget https://github.com/monsti/monsti-$*/archive/master.tar.gz -O module/$*.tar.gz
	cd module; tar xvf $*.tar.gz && mv monsti-$*-master $* && rm $*.tar.gz
	mkdir -p go/src/github.com/monsti/
	ln -s ../../../../module/$* go/src/github.com/monsti/monsti-$*
	cp -Rn module/$*/templates .
	cp -Rn module/$*/locale .

# Build module executable
go/bin/monsti-%: module/%
	$(GO) get github.com/monsti/monsti-$*

.PHONY: tests
tests: $(MODULES:%=test-module-%) daemon/test-worker

test-module-% test-%:
	$(GO) test github.com/monsti/monsti-$*

.PHONY: clean
clean:
	rm go/* -Rf
	rm static/aloha/ -R
	rm module/ -Rf
	rm templates/ -Rf
	rm locale/ -Rf

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
