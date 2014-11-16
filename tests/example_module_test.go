package browsertests

import "testing"

func TestExampleModule(t *testing.T) {
	b := setup(t)
	Must(b.VisitLink("Node Types"), "Could not open node types", t)
	Must(b.VisitLink("Module Example"), "Could not open module example type", t)
	Must(b.Contains("Foo Field"), "Expected content not present", t)
	Must(b.Contains("Site name: example"), "Site name not present", t)
}
