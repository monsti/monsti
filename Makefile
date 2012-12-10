.PHONY: extract-messages
extract-messages:
	find templates/ *.go -name "*.html" -o -name "*.go"| xargs cat \
	  | sed 's|{{G "\(.*\)"}}|gettext("\1");|g' \
	  | xgettext -d monsti-image -L C -p locale/ -kG -kGN:1,2 -
