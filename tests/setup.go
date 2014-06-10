package browsertests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tebeka/selenium"
)

var (
	appURL = "http://localhost:8080"
	wd     selenium.WebDriver
)

type Browser struct {
	wd selenium.WebDriver
}

// setup sets up a clean browser session.
func setup(t *testing.T) *Browser {
	if wd == nil {
		executor := "http://localhost:9515"
		if len(os.Getenv("SAUCE_USERNAME")) > 0 {
			executor = fmt.Sprintf(
				"http://%s:%s@localhost:4445/wd/hub",
				os.Getenv("SAUCE_USERNAME"),
				os.Getenv("SAUCE_ACCESS_KEY"),
			)
		}
		caps := selenium.Capabilities{"browserName": "chrome"}
		if len(os.Getenv("TRAVIS_JOB_NUMBER")) > 0 {
			caps["tunnel-identifier"] = os.Getenv("TRAVIS_JOB_NUMBER")
		}
		var err error
		wd, err = selenium.NewRemote(caps, executor)
		if err != nil {
			t.Fatal("Could not get new remote: ", err)
		}
		if err := wd.SetImplicitWaitTimeout(5 * time.Second); err != nil {
			t.Fatal("Could not set implicit wait timeout: ", err)
		}
		if err != nil {
			t.Fatal("Could not setup selenium remote: ", err)
		}
	} else {
		if err := wd.DeleteAllCookies(); err != nil {
			t.Fatal("Could not delete cookies: ", err)
		}
	}
	return &Browser{wd}
}

func (b *Browser) Quit() {
	if err := b.wd.Quit(); err != nil {
		panic("Could not quit browser session: " + err.Error())
	}
}

// FillById fills the element with the given id.
func (b *Browser) FillById(id, content string) error {
	elem, err := b.wd.FindElement(selenium.ById, id)
	if err != nil {
		return fmt.Errorf("browser: Could not find element to fill: %v")
	}
	if err := elem.Clear(); err != nil {
		return fmt.Errorf("browser: Could not clear element to fill: %v")
	}
	if err := elem.SendKeys(content); err != nil {
		return fmt.Errorf("browser: Could not fill element: %v", err)
	}
	return nil
}

// SubmitForm fills the given form with the fields in the map and submits it.
// The keys in fields are the field ids.
func (b *Browser) SubmitForm(fields map[string]string) error {
	for id, value := range fields {
		if err := b.FillById(id, value); err != nil {
			return fmt.Errorf("browser: Could not fill field %q: %v", id, err)
		}
	}
	btn, err := b.wd.FindElements(selenium.ByCSSSelector, "button[type=submit]")
	if err != nil {
		return fmt.Errorf("browser: Could not find submit button: %v", err)
	}
	if len(btn) != 1 {
		return fmt.Errorf("browser: Found %d submit button(s)", len(btn))
	}
	if err := btn[0].Submit(); err != nil {
		return fmt.Errorf("browser: Could not click submit button: %v", err)
	}
	return nil
}

// Go opens the given url.
func (b *Browser) Go(url string) error {
	if err := b.wd.Get(url); err != nil {
		return fmt.Errorf("browser: Could not open url: %v", err)
	}
	return nil
}

// Visit links opens the link with the given text.
func (b *Browser) VisitLink(text string) error {
	link, err := b.wd.FindElement(selenium.ByLinkText, text)
	if err != nil {
		return fmt.Errorf("browser: Could not find link %q: %v", text, err)
	}
	if err := link.Click(); err != nil {
		return fmt.Errorf("browser: Could not click link %q: %v", text, err)
	}
	return nil
}

// FindElement finds an element using the given css selector.
func (b *Browser) FindElement(selector string) (selenium.WebElement, error) {
	elem, err := b.wd.FindElement(selenium.ByCSSSelector, selector)
	if err != nil {
		return nil, fmt.Errorf("browser: Could not find element %q: %v", selector, err)
	}
	return elem, nil
}

// Contains checks if the given text is in the page source. If not, it returns
// an error.
func (b *Browser) Contains(text string) error {
	root, err := b.FindElement("html")
	if err != nil {
		return fmt.Errorf("browser: Could not get root element: %v", err)
	}
	src, err := root.Text()
	if err != nil {
		return fmt.Errorf("browser: Could not get page source: %v", err)
	}
	if !strings.Contains(src, text) {
		return fmt.Errorf("browser: Page source does not contain string.")
	}
	return nil
}

// Must expects err to be nil. If err is not nil, it calls t.Fail with the given
// error msg and the error.
func Must(err error, msg string, t *testing.T) {
	if err != nil {
		t.Fatalf("%v: %v", msg, err)
	}
}
