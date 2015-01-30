package browsertests

import (
	"testing"

	"github.com/tebeka/selenium"
)

// Login to Monsti.
func login(b Browser, t *testing.T) {
	// types/image/@@login must not redirect to the raw data
	if err := b.Go(appURL + "/types/image/@@login"); err != nil {
		t.Fatal("Could not visit login formular: ", err)
	}
	if err := b.SubmitForm(map[string]string{
		"Login":    "admin",
		"Password": "foofoo"}); err != nil {
		t.Fatal("Could not fill and submit login form: ", err)
	}
	if _, err := b.wd.FindElement(selenium.ByClassName, "admin-bar"); err != nil {
		t.Fatal("Could not find admin bar. Login failed?")
	}
	if err := b.Go(appURL); err != nil {
		t.Fatal("Could not visit main page: ", err)
	}
}

func TestLogin(t *testing.T) {
	b := setup(t)
	login(*b, t)
}
