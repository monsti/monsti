LOCALES=de

all:

locales: $(LOCALES:%=locale/%/LC_MESSAGES/monsti-document.mo)

.PHONY: locale/monsti-document.pot
locale/monsti-document.pot:
	find templates/ *.go -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-document -L C -p locale/ -kG -kGN:1,2 \
	      -o monsti-document.pot -


%.mo: %.po
	  msgfmt -c -v -o $@ $<

%.po: locale/monsti-document.pot
	  msgmerge -s -U $@ $<
