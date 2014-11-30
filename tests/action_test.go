package browsertests

import "testing"

func TestEdit(t *testing.T) {
	b := setup(t)
	login(*b, t)
	Must(b.VisitLink("Edit"), "Could not open edit formular", t)
	_, err := b.FindElement(".html-field")
	Must(err, "Could not find HTML field", t)
	Must(b.SubmitForm(nil), "Could not save changed content", t)
}
