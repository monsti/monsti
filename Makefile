GOPATH=$(CURDIR)/go/
GO=GOPATH=$(GOPATH) go
#GO_COMMON_OPTS=-race
GO_GET=$(GO) get $(GO_COMMON_OPTS)
GO_BUILD=$(GO) build $(GO_COMMON_OPTS)
GO_TEST=$(GO) test $(GO_COMMON_OPTS)

MODULES=daemon base taxonomy

LOCALES=de

VCS_REVISION:=$(shell git rev-list HEAD --count)
VCS_BRANCH:=$(shell git branch | sed -n '/\* /s///p')
MONSTI_VERSION=0.14.0.dev.$(VCS_BRANCH).$(VCS_REVISION)
DEB_VERSION=1

DIST_PATH=dist/monsti-$(MONSTI_VERSION)

WEBSHIM_VERSION=1.15.5

MODULE_PROGRAMS=$(MODULES:%=go/bin/monsti-%)

all: monsti bcrypt example-module

monsti: modules dep-webshim

.PHONY: bcrypt
bcrypt: 
	mkdir -p $(GOPATH)/bin
	cd utils/bcrypt && $(GO_GET) -d . && $(GO_BUILD) -o $(GOPATH)/bin/bcrypt .

.PHONY: upgrade
upgrade:
	$(GO_GET) pkg.monsti.org/monsti/utils/upgrade

modules: $(MODULES)
$(MODULES): %: go/bin/monsti-%

dist: all
	rm -Rf $(DIST_PATH)
	mkdir -p $(DIST_PATH)/bin
	cp go/bin/* $(DIST_PATH)/bin
	mkdir -p $(DIST_PATH)/share
	cp -RL locale static templates $(DIST_PATH)/share
	mkdir -p $(DIST_PATH)/doc
	cp CHANGES COPYING LICENSE README.md $(DIST_PATH)/doc
	mkdir -p $(DIST_PATH)/etc
	cp -R example/config/* $(DIST_PATH)/etc
	mkdir -p $(DIST_PATH)/run
	mkdir -p $(DIST_PATH)/data
	cp -R example/data/localhost $(DIST_PATH)/data/localhost
	cp -p example/start.sh $(DIST_PATH)/start.sh
	sed -e 's/\.\.\/go\///' -e 's/config/etc/' example/start.sh \
		> $(DIST_PATH)/start.sh
	tar -C dist -czf dist/monsti-$(MONSTI_VERSION).tar.gz monsti-$(MONSTI_VERSION)

dist-deb: all
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
	sed -e 's/config/etc\/monsti/' -e 's/\.\.\/go\///' \
		example/start.sh > $(DIST_PATH)/usr/share/doc/monsti/examples/start.sh
	cp CHANGES COPYING LICENSE README.md $(DIST_PATH)/usr/share/doc/monsti
	mkdir -p $(DIST_PATH)/etc/monsti/sites
	cp -R example/config/* $(DIST_PATH)/etc/monsti
	sed -e 's/\.\.\/run/\/var\/run\/monsti/' -e 's/\.\.\/data/\/var\/lib\/monsti/' \
		-e 's/\.\.\/share/\/usr\/share\/monsti/' example/config/monsti.yaml \
		> $(DIST_PATH)/etc/monsti/monsti.yaml
	mkdir -p $(DIST_PATH)/var/run/monsti
	mkdir -p $(DIST_PATH)/var/lib/monsti
	cp -R example/data/localhost $(DIST_PATH)/usr/share/doc/monsti/examples/localhost
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
	$(GO_TEST) ./core/...
	$(GO_TEST) ./api/...
	$(GO_TEST) ./utils/...

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
	find templates/ core/ api/ -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|\ngettext("\1");\n|g' \
	  | xgettext -d monsti-daemon -L C -p locale/ -kG -kGN:1,2 \
	      -o monsti-daemon.pot -

%.mo: %.po
	  msgfmt -c -v -o $@ $<

%.po: locale/monsti-daemon.pot
	  msgmerge -s -U $@ $<

doc: doc/manual.html doc/release_notes.html

doc/%.html: doc/%.adoc
	asciidoc $<

.PHONY: example/monsti-example-module/monsti-example-module
example/monsti-example-module/monsti-example-module:
	ln -sf ../../go example/monsti-example-module/go
	$(MAKE) -C example/monsti-example-module

example-module: go/bin/monsti-example-module

go/bin/monsti-example-module: example/monsti-example-module/monsti-example-module
	cp -f example/monsti-example-module/monsti-example-module $(GOPATH)/bin
	rm -f templates/example
	ln -sf ../example/monsti-example-module/templates templates/example
