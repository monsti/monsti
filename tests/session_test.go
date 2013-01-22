package browsertests

import (
	"testing"
	"bitbucket.org/tebeka/selenium"
)

// Login to Monsti.
func login(b Browser, t *testing.T) {
	if err := b.Go(appURL + "/@@login"); err != nil {
		t.Fatal("Could not visit login formular: ", err);
	}
	if err := b.SubmitForm(map[string]string{
		"login": "admin",
		"password": "foofoo"}); err != nil {
		t.Fatal("Could not fill and submit login form: ", err)
	}
	if _, err := b.wd.FindElement(selenium.ById, "admin-bar"); err != nil {
		t.Fatal("Could not find admin bar. Login failed?")
	}
}

func TestLogin(t *testing.T) {
	b := setup(t)
	login(*b, t)
}
