package main

import (
	"testing"
)

func TestContactFormData(t *testing.T) {
	data := contactFormData{
		Name:    "Johannes Schmidt",
		Email:   "joe@smithy.de",
		Subject: "",
		Message: "I forgot the subject! D'oh!"}
	errors := data.Check()
	if len(errors) != 1 {
		t.Errorf("len(%v.Check()) = %v, want 1", data, len(errors))
	}
}
