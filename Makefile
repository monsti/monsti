LOCALES=de

all:

locales: $(LOCALES:%=locale/%/LC_MESSAGES/monsti-contactform.mo)

.PHONY: locale/monsti-contactform.pot
locale/monsti-contactform.pot:
	find templates/ *.go -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-contactform -L C -p locale/ -kG -kGN:1,2 \
	      -o monsti-contactform.pot -


%.mo: %.po
	  msgfmt -c -v -o $@ $<

%.po: locale/monsti-contactform.pot
	  msgmerge -s -U $@ $<
