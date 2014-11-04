package browsertests

import "testing"

func TestLocalFields(t *testing.T) {
	b := setup(t)
	Must(b.VisitLink("Nodes"), "Could not open nodes", t)
	Must(b.VisitLink("Local fields"), "Could not open local fields example", t)
	Must(b.Contains("A local field"), "Local field  not present", t)
}
