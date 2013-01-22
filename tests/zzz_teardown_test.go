package browsertests

import (
	"testing"
)

// Tear down browser session.
func TestTeardown(t *testing.T) {
	b := setup(t)
	b.Quit()
}
