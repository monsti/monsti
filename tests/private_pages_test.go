package browsertests

import "testing"

func TestViewPrivatePage(t *testing.T) {
	b := setup(t)
	Must(b.VisitLink("Nodes"), "", t)
	MustErr(b.Contains("Private document"), "does not contain",
		"Private document should not be visible", t)
	Must(b.Go(appURL+"/nodes/private"), "", t)
	MustErr(b.Contains("This document is only visible"), "does not contain",
		"Private document should not be reachable", t)
	login(*b, t)
	Must(b.VisitLink("Nodes"), "", t)
	Must(b.VisitLink("Private document"), "", t)
	Must(b.Contains("This document is only visible"), "", t)
}
