package browsertests

import (
	"testing"
)

func TestEdit(t *testing.T) {
	b := setup(t)
	login(*b, t)
	Must(b.VisitLink("About"), "Could not visit node", t)
	Must(b.VisitLink("Edit"), "Could not open edit formular", t)
	area, err := b.FindElement(".aloha-editable > p");
	Must(err, "Could not find Aloha editor", t)
	Must(area.Click(), "Could not enter Aloha editor", t)
	Must(area.SendKeys("Hey World!"),
		"Could not enter text into Aloha editor",t)
	Must(b.SubmitForm(nil), "Could not save changed content", t)
	Must(b.Contains("Hey World!"), "Change not saved", t)
}
