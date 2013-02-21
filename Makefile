GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go

MODULES=daemon document contactform image

#ALOHA_VERSION=0.22.6
#DAEMON_VERSION=0.4
#DOCUMENT_VERSION=0.1
#CONTACTFORM_VERSION=0.1
#IMAGE_VERSION=0.1

DAEMON_VERSION=master
DOCUMENT_VERSION=master
CONTACTFORM_VERSION=master
IMAGE_VERSION=master

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

module/daemon.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-daemon/archive/$(DAEMON_VERSION).tar.gz -O module/daemon.tar.gz

module/image.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-image/archive/$(IMAGE_VERSION).tar.gz -O module/image.tar.gz

module/document.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-document/archive/$(DOCUMENT_VERSION).tar.gz -O module/document.tar.gz

module/contactform.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-contactform/archive/$(CONTACTFORM_VERSION).tar.gz -O module/contactform.tar.gz

module/%: module/%.tar.gz
	cd module; tar xf $*.tar.gz && mv monsti-$*-* $*


$(MODULE_SOURCES): go/src/github.com/monsti/monsti-%: module/%
	mkdir -p go/src/github.com/monsti/
	ln -sf ../../../../module/$* go/src/github.com/monsti/monsti-$*

$(MODULE_TEMPLATES): templates/%: module/%
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
	util/test-l10n rpc/test-client

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

clean-templates:
	# FIXME rm templates/ -Rf

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
