GOPATH=$(CURDIR)/go/
GO=GOPATH=$(GOPATH) go
#GO_COMMON_OPTS=-race
GO_GET=$(GO) get $(GO_COMMON_OPTS)
GO_BUILD=$(GO) build $(GO_COMMON_OPTS)
GO_TEST=$(GO) test $(GO_COMMON_OPTS)

MODULES=daemon

LOCALES=de

VCS_REVISION:=$(shell git rev-list HEAD --count)
VCS_BRANCH:=$(shell git branch | sed -n '/\* /s///p')
MONSTI_VERSION=0.8.0.rc.$(VCS_BRANCH).$(VCS_REVISION)
DEB_VERSION=1

DIST_PATH=dist/monsti-$(MONSTI_VERSION)

TINYMCE_VERSION=4.1.7
WEBSHIM_VERSION=1.15.5

MODULE_PROGRAMS=$(MODULES:%=go/bin/monsti-%)

all: monsti bcrypt example-module

monsti: modules dep-tinymce-editor dep-jquery dep-webshim

.PHONY: bcrypt
bcrypt: 
	mkdir -p $(GOPATH)/bin
	cd utils/bcrypt && $(GO_GET) -d . && $(GO_BUILD) -o $(GOPATH)/bin/bcrypt .

.PHONY: upgrade
upgrade:
	$(GO_GET) pkg.monsti.org/monsti/utils/upgrade

modules: $(MODULES)
$(MODULES): %: go/bin/monsti-%

dist: monsti bcrypt
	rm -Rf $(DIST_PATH)
	mkdir -p $(DIST_PATH)/bin
	cp go/bin/* $(DIST_PATH)/bin
	mkdir -p $(DIST_PATH)/share
	cp -RL locale static templates $(DIST_PATH)/share
	mkdir -p $(DIST_PATH)/doc
	cp CHANGES COPYING LICENSE README.md $(DIST_PATH)/doc
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

dist-deb: monsti bcrypt
	rm -Rf $(DIST_PATH)
	mkdir -p $(DIST_PATH)/usr/bin
	cp go/bin/* $(DIST_PATH)/usr/bin
	mkdir -p $(DIST_PATH)/usr/share/monsti
	cp -RL static templates $(DIST_PATH)/usr/share/monsti
	cp -RL locale $(DIST_PATH)/usr/share
	find $(DIST_PATH)/usr/share/locale/ -not -name "*.mo" -exec rm {} \;
	rm -f $(DIST_PATH)/usr/share/locale/*.pot
	mkdir -p $(DIST_PATH)/usr/share/doc/monsti/examples
	cp example/start.sh $(DIST_PATH)/usr/share/doc/monsti/examples
	sed -i 's/\.\.\/go\///' $(DIST_PATH)/usr/share/doc/monsti/examples/start.sh
	sed -i 's/config/etc\/monsti/' $(DIST_PATH)/usr/share/doc/monsti/examples/start.sh
	cp CHANGES COPYING LICENSE README.md $(DIST_PATH)/usr/share/doc/monsti
	mkdir -p $(DIST_PATH)/etc/monsti/sites
	cp -R example/config/* $(DIST_PATH)/etc/monsti
	sed -i 's/\.\.\/share/\/usr\/share\/monsti/' $(DIST_PATH)/etc/monsti/monsti.yaml
	sed -i 's/\.\.\/data/\/var\/lib\/monsti/' $(DIST_PATH)/etc/monsti/monsti.yaml
	sed -i 's/\.\.\/run/\/var\/run\/monsti/' $(DIST_PATH)/etc/monsti/monsti.yaml
	mv $(DIST_PATH)/etc/monsti/sites/example $(DIST_PATH)/etc/monsti/sites/default
	mkdir -p $(DIST_PATH)/var/run/monsti
	mkdir -p $(DIST_PATH)/var/lib/monsti
	cp -R example/data/example $(DIST_PATH)/usr/share/doc/monsti/examples/default
	find $(DIST_PATH) -type d -exec chmod 755 {} \;
	find $(DIST_PATH) -not -type d -exec chmod 644 {} \;
	chmod 755 $(DIST_PATH)/usr/bin/*
	fpm -s dir -t deb -a all \
		-C $(DIST_PATH) \
		-n monsti \
		-p dist/monsti_$(MONSTI_VERSION)-$(DEB_VERSION).deb \
		--version $(MONSTI_VERSION)-$(DEB_VERSION) \
		--config-files etc \
		etc usr var

go/src/pkg.monsti.org/monsti:
	mkdir -p $(GOPATH)/src/pkg.monsti.org
	ln -sf ../../.. $(GOPATH)/src/pkg.monsti.org/monsti

# Build module executable
.PHONY: $(MODULE_PROGRAMS)
$(MODULE_PROGRAMS): go/bin/%: go/src/pkg.monsti.org/monsti
	$(GO_GET) pkg.monsti.org/monsti/core/$*

.PHONY: test
test: monsti
	cd $(GOPATH)/src/pkg.monsti.org/monsti/api && $(GO_TEST) ./...
	cd $(GOPATH)/src/pkg.monsti.org/monsti/core && $(GO_TEST) ./...
	cd $(GOPATH)/src/pkg.monsti.org/monsti/utils && $(GO_TEST) ./...

.PHONY: test-browser
test-browser: monsti
	$(GO_GET) github.com/tebeka/selenium
	cd tests && $(GO_TEST)

.PHONY: clean
clean:
	rm go/* -Rf
	rm static/lib/ -Rf
	rm dist/ -Rf
	$(MAKE) -C example/monsti-example-module clean

dep-tinymce-editor: static/lib/tinymce/
static/lib/tinymce/:
	wget -nv http://download.moxiecode.com/tinymce/tinymce_$(TINYMCE_VERSION).zip
	unzip -q tinymce_$(TINYMCE_VERSION).zip
	rm tinymce_$(TINYMCE_VERSION).zip
	mkdir -p static/lib
	mv tinymce/js/tinymce static/lib
	rm tinymce -R

dep-jquery: static/js/jquery.min.js
static/js/jquery.min.js:
	wget -nv http://code.jquery.com/jquery-1.8.2.min.js
	mkdir -p static/lib
	mv jquery-1.8.2.min.js static/lib/jquery.min.js

dep-webshim: static/lib/webshim/
static/lib/webshim/:
	wget -nv https://github.com/aFarkas/webshim/archive/$(WEBSHIM_VERSION).zip
	unzip -q $(WEBSHIM_VERSION).zip
	rm $(WEBSHIM_VERSION).zip
	mkdir -p static/lib
	mv webshim-$(WEBSHIM_VERSION)/ static/lib/webshim

locales: $(LOCALES:%=locale/%/LC_MESSAGES/monsti-daemon.mo)

.PHONY: locale/monsti-daemon.pot
locale/monsti-daemon.pot:
	find templates/ core/ -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-daemon -L C -p locale/ -kG -kGN:1,2 \
	      -o monsti-daemon.pot -

%.mo: %.po
	  msgfmt -c -v -o $@ $<

%.po: locale/monsti-daemon.pot
	  msgmerge -s -U $@ $<

doc: doc/manual.html

doc/%.html: doc/%.adoc
	asciidoc $<

.PHONY: example/monsti-example-module/monsti-example-module
example/monsti-example-module/monsti-example-module:
	$(MAKE) -C example/monsti-example-module

example-module: go/bin/monsti-example-module

go/bin/monsti-example-module: example/monsti-example-module/monsti-example-module
	cp example/monsti-example-module/monsti-example-module $(GOPATH)/bin
	ln -sf ../example/monsti-example-module/templates templates/example
