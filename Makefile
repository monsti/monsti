GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
#GO_COMMON_OPTS=-race
GO_GET=$(GO) get $(GO_COMMON_OPTS)
GO_TEST=$(GO) test $(GO_COMMON_OPTS)

MODULES=daemon httpd data document contactform mail image

MONSTI_VERSION=unstable
DIST_PATH=dist/monsti-$(MONSTI_VERSION)

ALOHA_VERSION=0.23.2

MODULE_PROGRAMS=$(MODULES:%=go/bin/monsti-%)
MODULE_SOURCES=$(MODULES:%=go/src/pkg.monsti.org/monsti-%)

all: monsti bcrypt

monsti: modules templates locales templates/master.html dep-aloha-editor dep-jquery

.PHONY: bcrypt
bcrypt: 
	$(GO_GET) pkg.monsti.org/monsti-login/bcrypt

modules: $(MODULES)
$(MODULES): %: go/bin/monsti-%

locales: $(MODULES:%=locales-monsti-%)

locales-monsti-%:
	mkdir -p locale/
	mkdir -p modules/monsti-$*/locale/
	cp -Rn modules/monsti-$*/locale .

templates: $(MODULES:%=templates-monsti-%)

templates-monsti-%:
	mkdir -p templates/
	mkdir -p modules/monsti-$*/templates/
	ln -nsf ../modules/monsti-$*/templates templates/$*

templates/master.html: templates/httpd/master.html
	for i in $(wildcard templates/httpd/*); \
	do \
		ln -nsf httpd/`basename $${i}` templates/`basename $${i}`; \
	done; \
  #rm templates/httpd/templates

modules/monsti-%:
	git clone git@gitorious.org:monsti/monsti-$*.git modules/monsti-$*

$(MODULE_SOURCES): go/src/pkg.monsti.org/monsti-%: modules/monsti-%
	mkdir -p go/src/pkg.monsti.org
	ln -sf ../../../modules/monsti-$* go/src/pkg.monsti.org/monsti-$*

dist: monsti bcrypt
	mkdir -p $(DIST_PATH)/bin
	cp go/bin/* $(DIST_PATH)/bin
	mkdir -p $(DIST_PATH)/share
	cp -RL locale static templates $(DIST_PATH)/share
	mkdir -p $(DIST_PATH)/doc/examples
	cp CHANGES COPYING LICENSE README $(DIST_PATH)/doc
	cp -R example/config/* example/data example/start.sh $(DIST_PATH)/doc/examples
	mkdir -p $(DIST_PATH)/etc
	mkdir -p $(DIST_PATH)/run
	mkdir -p $(DIST_PATH)/data
	tar -C dist -czf dist/monsti-$(MONSTI_VERSION).tar.gz monsti-$(MONSTI_VERSION)

# Build module executable
.PHONY: $(MODULE_PROGRAMS)
$(MODULE_PROGRAMS): go/bin/monsti-%: go/src/pkg.monsti.org/monsti-%
	$(GO_GET) pkg.monsti.org/monsti-$*

.PHONY: tests
tests: $(MODULES:%=test-module-%) monsti-daemon/test-worker util/test-template util/test-testing\
	util/test-l10n rpc/test-client

test-module-%:
	$(GO_TEST) pkg.monsti.org/monsti-$*

test-%:
	$(GO_TEST) pkg.monsti.org/$*

.PHONY: clean
clean:
	rm go/* -Rf
	rm static/aloha/ -Rf
	rm locale/ -Rf
	rm dist/ -Rf
	rm templates/ -Rf

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
