package browsertests

import "testing"

func TestViewContactform(t *testing.T) {
	b := setup(t)
	Must(b.VisitLink("Contact"), "Could not open contact formular", t)
	Must(b.Contains("A contact form!"), "Contactform body not present", t)
}
