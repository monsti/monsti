.PHONY: extract-messages
extract-messages:
	mkdir -p locale/
	find . -name "*.html" -or -name "*.go" | xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-daemon -L C -p locale/ -kG -kGN:1,2 -
