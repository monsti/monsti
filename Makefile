LOCALES=de

all:

locales: $(LOCALES:%=locale/%/LC_MESSAGES/monsti-httpd.mo)

.PHONY: locale/monsti-httpd.pot
locale/monsti-httpd.pot:
	find templates/ *.go -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-httpd -L C -p locale/ -kG -kGN:1,2 \
	  -o monsti-httpd.pot -


%.mo: %.po
	msgfmt -c -v -o $@ $<

%.po: locale/monsti-httpd.pot
	msgmerge -s -U $@ $<
