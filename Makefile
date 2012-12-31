GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
ALOHA_VERSION=0.22.6

MODULES=daemon document contactform image

MODULE_PROGRAMS=$(MODULES:%=go/bin/monsti-%)
MODULE_SOURCES=$(MODULES:%=go/src/github.com/monsti/monsti-%)
MODULE_TEMPLATES=$(MODULES:%=templates/%)
MODULE_LOCALES=$(MODULES:%=locale/monsti-%.pot)

all: monsti bcrypt

monsti: dep-aloha-editor dep-jquery modules

.PHONY: bcrypt
bcrypt: 
	$(GO) get github.com/monsti/monsti-daemon/tools/bcrypt

modules: $(MODULES)
$(MODULES): %: go/bin/monsti-% locale/monsti-%.pot templates/%

module/%:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-$*/archive/master.tar.gz -O module/$*.tar.gz
	cd module; tar xf $*.tar.gz && mv monsti-$*-master $* && rm $*.tar.gz

$(MODULE_SOURCES): go/src/github.com/monsti/monsti-%: module/%
	mkdir -p go/src/github.com/monsti/
	ln -sf ../../../../module/$* go/src/github.com/monsti/monsti-$*

$(MODULE_TEMPLATES): templates/%: module/%
	mkdir -p templates/
	ln -sf ../module/$*/templates templates/$*

$(MODULE_LOCALES): locale/monsti-%.pot: module/%
	mkdir -p locale/
	cp -Rn module/$*/locale .

# Build module executable
.PHONY: $(MODULE_PROGRAMS)
$(MODULE_PROGRAMS): go/bin/monsti-%: go/src/github.com/monsti/monsti-%
	$(GO) get github.com/monsti/monsti-$*

.PHONY: tests
tests: $(MODULES:%=test-module-%) monsti-daemon/test-worker util/test-template util/test-testing\
	util/test-l10n

test-module-%:
	$(GO) test github.com/monsti/monsti-$*

test-%:
	$(GO) test github.com/monsti/$*

.PHONY: clean
clean: clean-templates
	rm go/* -Rf
	rm static/aloha/ -R
	rm module/ -Rf
	rm locale/ -Rf

.PHONY: clean-templates
clean-templates:
	rm templates/ -Rf

dep-aloha-editor: static/aloha/
static/aloha/:
	wget -nv https://github.com/downloads/alohaeditor/Aloha-Editor/alohaeditor-$(ALOHA_VERSION).zip
	unzip -q alohaeditor-$(ALOHA_VERSION).zip
	mkdir static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/lib static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/css static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/img static/aloha
	mv alohaeditor-$(ALOHA_VERSION)/aloha/plugins static/aloha
	rm alohaeditor-$(ALOHA_VERSION) -R
	rm alohaeditor-$(ALOHA_VERSION).zip

dep-jquery: static/js/jquery.min.js
static/js/jquery.min.js:
	wget -nv http://code.jquery.com/jquery-1.8.2.min.js
	mv jquery-1.8.2.min.js static/js/jquery.min.js
