GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
#GO_COMMON_OPTS=-race
GO_GET=$(GO) get $(GO_COMMON_OPTS)
GO_TEST=$(GO) test $(GO_COMMON_OPTS)

MODULES=daemon httpd data document contactform mail image

ALOHA_VERSION=0.23.2

#DAEMON_VERSION=0.6.0
#DOCUMENT_VERSION=0.3.0
#CONTACTFORM_VERSION=0.3.0
#IMAGE_VERSION=0.1.0
#MAIL_VERSION=0.1.0
#DATA_VERSION=0.1.0
#HTTPD_VERSION=0.1.0

DAEMON_VERSION=master
DOCUMENT_VERSION=master
CONTACTFORM_VERSION=master
IMAGE_VERSION=master
MAIL_VERSION=master
DATA_VERSION=master
HTTPD_VERSION=master

MODULE_PROGRAMS=$(MODULES:%=go/bin/monsti-%)
MODULE_SOURCES=$(MODULES:%=go/src/github.com/monsti/monsti-%)
MODULE_TEMPLATES=$(MODULES:%=templates/%)
MODULE_LOCALES=$(MODULES:%=locale/monsti-%.pot)

all: monsti bcrypt

monsti: dep-aloha-editor dep-jquery modules

.PHONY: bcrypt
bcrypt: 
	$(GO_GET) github.com/monsti/monsti-login/bcrypt

modules: $(MODULES)
$(MODULES): %: go/bin/monsti-% locale/monsti-%.pot templates/%

module/daemon.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-daemon/archive/$(DAEMON_VERSION).tar.gz -O module/daemon.tar.gz

module/data.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-data/archive/$(DATA_VERSION).tar.gz -O module/data.tar.gz

module/httpd.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-httpd/archive/$(HTTPD_VERSION).tar.gz -O module/httpd.tar.gz

module/image.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-image/archive/$(IMAGE_VERSION).tar.gz -O module/image.tar.gz

module/mail.tar.gz:
	mkdir -p module/
	wget -nv https://github.com/monsti/monsti-mail/archive/$(IMAGE_VERSION).tar.gz -O module/mail.tar.gz

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

$(MODULE_TEMPLATES):: templates/%: module/%
	ln -sf ../module/$*/templates templates/$*

templates/httpd:: module/httpd
	for i in $(wildcard templates/httpd/*); \
	do \
		ln -sf httpd/`basename $${i}` templates/`basename $${i}`; \
	done; \

$(MODULE_LOCALES): locale/monsti-%.pot: module/%
	mkdir -p locale/
	cp -Rn module/$*/locale .

# Build module executable
.PHONY: $(MODULE_PROGRAMS)
$(MODULE_PROGRAMS): go/bin/monsti-%: go/src/github.com/monsti/monsti-%
	$(GO_GET) github.com/monsti/monsti-$*

.PHONY: tests
tests: $(MODULES:%=test-module-%) monsti-daemon/test-worker util/test-template util/test-testing\
	util/test-l10n rpc/test-client

test-module-%:
	$(GO_TEST) github.com/monsti/monsti-$*

test-%:
	$(GO_TEST) github.com/monsti/$*

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
	wget -nv http://aloha-editor.org/builds/stable/alohaeditor-$(ALOHA_VERSION).zip
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
