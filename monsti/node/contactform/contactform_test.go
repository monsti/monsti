package main

import (
	"datenkarussell.de/monsti/template"
	"testing"
)

func TestContactFormData(t *testing.T) {
	data := contactFormData{
		Name:    "Johannes Schmidt",
		Email:   "joe@smithy.de",
		Subject: "",
		Message: "I forgot the subject! D'oh!"}
	errors := make(template.FormErrors)
	data.Check(&errors)
	if len(errors) != 1 {
		t.Errorf("len(%v.Check()) = %v, want 1", data, len(errors))
	}
}
