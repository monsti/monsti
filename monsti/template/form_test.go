package template

import "testing"

func TestRequire(t *testing.T) {
	invalid, valid := "", "foo"
	validator := Required()
	err := validator(valid)
	if err != nil {
		t.Errorf("require(%v) = %v, want %v", valid, err, nil)
	}
	err = validator(invalid)
	if err == nil {
		t.Errorf("require(%v) = %v, want %v", invalid, err, "'Required.'")
	}
}

func TestContactFormData(t *testing.T) {
	data := ContactFormData{
		Name:    "Johannes Schmidt",
		Email:   "joe@smithy.de",
		Subject: "",
		Message: "I forgot the subject! D'oh!"}
	errors := data.Check()
	if len(errors) != 1 {
		t.Errorf("len(%v.Check()) = %v, want 1", data, len(errors))
	}
}
