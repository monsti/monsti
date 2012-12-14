LOCALES=de

all:

locales: $(LOCALES:%=locale/%/LC_MESSAGES/monsti-daemon.mo)

.PHONY: locale/monsti-daemon.pot
locale/monsti-daemon.pot:
	find templates/ *.go -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-daemon -L C -p locale/ -kG -kGN:1,2 \
	      -o monsti-daemon.pot -


%.mo: %.po
	  msgfmt -c -v -o $@ $<

%.po: locale/monsti-daemon.pot
	  msgmerge -s -U $@ $<
