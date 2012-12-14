LOCALES=de

all:

locales: $(LOCALES:%=locale/%/LC_MESSAGES/monsti-image.mo)

.PHONY: locale/monsti-image.pot
locale/monsti-image.pot:
	find templates/ *.go -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-image -L C -p locale/ -kG -kGN:1,2 \
	      -o monsti-image.pot -


%.mo: %.po
	  msgfmt -c -v -o $@ $<

%.po: locale/monsti-image.pot
	  msgmerge -s -U $@ $<
