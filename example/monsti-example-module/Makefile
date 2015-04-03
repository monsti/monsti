VCS_REVISION:=$(shell git rev-list HEAD --count)
VCS_BRANCH:=$(shell git branch | sed -n '/\* /s///p')
MODULE_NAME=example-module
MODULE_VERSION=0.1.0.dev.$(VCS_BRANCH).$(VCS_REVISION)
DEB_VERSION=1

GOPATH=$(CURDIR)/go/

GO=GOPATH=$(GOPATH) go
#GO_COMMON_OPTS=-race
GO_GET=$(GO) get $(GO_COMMON_OPTS)
GO_BUILD=$(GO) build $(GO_COMMON_OPTS)
GO_TEST=$(GO) test $(GO_COMMON_OPTS)

LOCALES=de

MODULE_PROGRAMS=go/bin/monsti-$(MODULE_NAME)
PACKAGE_NAME=monsti-$(MODULE_NAME)-$(MODULE_VERSION)
DIST_PATH=dist/$(PACKAGE_NAME)

all: module

dist: module bcrypt
	rm -Rf $(DIST_PATH)
	mkdir -p $(DIST_PATH)/usr/bin
	cp go/bin/* $(DIST_PATH)/usr/bin
	mkdir -p $(DIST_PATH)/usr/share/monsti-$(MODULE_NAME)
	cp -RL static templates $(DIST_PATH)/usr/share/monsti-$(MODULE_NAME)

	cp -RL locale $(DIST_PATH)/usr/share
	find $(DIST_PATH)/usr/share/locale/ -not -name "*.mo" -exec rm {} \;
	rm -f $(DIST_PATH)/usr/share/locale/*.pot

	cp CHANGES.md COPYING LICENSE README.md $(DIST_PATH)/usr/share/doc/monsti-$(MODULE_NAME)

	find $(DIST_PATH) -type d -exec chmod 755 {} \;
	find $(DIST_PATH) -not -type d -exec chmod 644 {} \;
	chmod 755 $(DIST_PATH)/usr/bin/*

dist-tar: dist
	tar -C dist -czf dist/$(PACKAGE_NAME).tar.gz $(DIST_PATH)

dist-deb: dist
	fpm -s dir -t deb -a all \
		--depends libmagickcore5 \
		-C $(DIST_PATH) \
		-n monsti-$(MODULE_NAME) \
		-p dist/monsti-$(MODULE_NAME)_$(MODULE_VERSION)-$(DEB_VERSION).deb \
		--version $(MMODULE_VERSION)-$(DEB_VERSION) \
		--config-files etc \
		etc usr var

.PHONY: module
module:
	$(GO_GET) -d
	$(GO_BUILD)

.PHONY: test
test: module
	$(GO_TEST) .

.PHONY: clean
clean:
	rm go/ -Rf
	rm dist/ -Rf

locales: $(LOCALES:%=locale/%/LC_MESSAGES/monsti-$(MODULE_NAME).mo)

.PHONY: locale/monsti-$(MODULE_NAME).pot
locale/monsti-$(MODULE_NAME).pot:
	find . -path "./go/*" -prune  -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-$(MODULE_NAME) -L C -p locale/ -kG -kGN:1,2 \
	      -o monsti-$(MODULE_NAME).pot -

%.mo: %.po
	  msgfmt -c -v -o $@ $<

%.po: locale/monsti-$(MODULE_NAME).pot
	  msgmerge -s -U $@ $<
