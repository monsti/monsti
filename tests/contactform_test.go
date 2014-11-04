package browsertests

import "testing"

func TestViewContactform(t *testing.T) {
	b := setup(t)
	Must(b.VisitLink("Node Types"), "Could not open node types", t)
	Must(b.VisitLink("Contact Form"), "Could not open contact formular", t)
	Must(b.Contains("A contact form!"), "Contactform body not present", t)
}
