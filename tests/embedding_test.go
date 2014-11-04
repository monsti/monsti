package browsertests

import "testing"

func TestEmbedding(t *testing.T) {
	b := setup(t)
	Must(b.VisitLink("Nodes"), "Could not open nodes", t)
	Must(b.VisitLink("Embedding"), "Could not open embedding example", t)
	Must(b.Contains("I want you to embed me!"), "Embedded content not present", t)
}
