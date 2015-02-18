package browsertests

import "testing"

func TestFields(t *testing.T) {
	b := setup(t)
	login(*b, t)
	Must(b.VisitLink("Fields"), "Could not open nodes", t)
	Must(b.Contains("31. Dec 2014 23:55"), "Can't find local datetime", t)
	Must(b.VisitLink("Edit"), "Could not open edit formular", t)
	field, err := b.FindElement("#Fields\\.example\\.DateTime")
	Must(err, "Could not find datetime field", t)
	date, err := field.GetAttribute("value")
	Must(err, "Could not get value of datetime field", t)
	expected := "2014-12-31T23:55"
	if date != expected {
		t.Errorf("Value of DateTime field should be %v, got %v",
			expected, date)
	}
}
