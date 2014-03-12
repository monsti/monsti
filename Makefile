GOPATH=$(PWD)/go/
GO=GOPATH=$(GOPATH) go
#GO_COMMON_OPTS=-race
GO_GET=$(GO) get $(GO_COMMON_OPTS)
GO_BUILD=$(GO) build $(GO_COMMON_OPTS)
GO_TEST=$(GO) test $(GO_COMMON_OPTS)

MODULES=daemon httpd data document contactform mail image

MONSTI_VERSION=unstable
DIST_PATH=dist/monsti-$(MONSTI_VERSION)

ALOHA_VERSION=0.23.2

MODULE_PROGRAMS=$(MODULES:%=go/bin/monsti-%)

all: monsti bcrypt

monsti: modules templates locales templates/master.html dep-aloha-editor dep-jquery

.PHONY: bcrypt
bcrypt: 
	mkdir -p $(GOPATH)/bin
	cd utils/bcrypt && $(GO_GET) -d . && $(GO_BUILD) -o $(GOPATH)/bin/bcrypt .

modules: $(MODULES)
$(MODULES): %: go/bin/monsti-%

locales: $(MODULES:%=locales-monsti-%)

locales-monsti-%:
	mkdir -p locale/
	mkdir -p core/$*/locale/
	cp -Rn core/$*/locale .

templates: $(MODULES:%=templates-monsti-%)

templates-monsti-%:
	mkdir -p templates/
	mkdir -p core/$*/templates/
	ln -nsf ../core/$*/templates templates/$*

templates/master.html: templates/httpd/master.html
	for i in $(wildcard templates/httpd/*); \
	do \
		ln -nsf httpd/`basename $${i}` templates/`basename $${i}`; \
	done; \
  #rm templates/httpd/templates

core/%:
	git clone git@gitorious.org:monsti/$*.git core/$*

dist: monsti bcrypt
	mkdir -p $(DIST_PATH)/bin
	cp go/bin/* $(DIST_PATH)/bin
	mkdir -p $(DIST_PATH)/share
	cp -RL locale static templates $(DIST_PATH)/share
	mkdir -p $(DIST_PATH)/doc
	cp CHANGES COPYING LICENSE README $(DIST_PATH)/doc
	mkdir -p $(DIST_PATH)/etc
	cp -R example/config/* $(DIST_PATH)/etc
	mv $(DIST_PATH)/etc/sites/example $(DIST_PATH)/etc/sites/default
	mkdir -p $(DIST_PATH)/run
	mkdir -p $(DIST_PATH)/data
	cp -R example/data/example $(DIST_PATH)/data/default
	cp example/start.sh $(DIST_PATH)/
	sed -i 's/\.\.\/go\///' $(DIST_PATH)/start.sh
	sed -i 's/config/etc/' $(DIST_PATH)/start.sh
	tar -C dist -czf dist/monsti-$(MONSTI_VERSION).tar.gz monsti-$(MONSTI_VERSION)

go/src/pkg.monsti.org/monsti:
	mkdir -p $(GOPATH)/src/pkg.monsti.org
	ln -sf ../../.. $(GOPATH)/src/pkg.monsti.org/monsti

# Build module executable
.PHONY: $(MODULE_PROGRAMS)
$(MODULE_PROGRAMS): go/bin/monsti-%: go/src/pkg.monsti.org/monsti
	mkdir -p $(GOPATH)/bin
	$(GO_GET) -d pkg.monsti.org/monsti/core/$*
	cd core/$* && $(GO_BUILD) -o $(GOPATH)/bin/monsti-$* .

.PHONY: tests
tests: $(MODULES:%=test-module-%) util/test-template util/test-testing\
	util/test-l10n rpc/test-client

test-module-%:
	cd core/$* && $(GO_TEST) .

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
