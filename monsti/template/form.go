package template

import (
	"code.google.com/p/gorilla/schema"
	"github.com/chrneumann/g5t"
	"errors"
)

var G func(string) string = g5t.String

// FormValidator is a function which validates a string.
type FormValidator func(string) error

//  Required is a formValidator to check for non empty values.
func Required() FormValidator {
	return func(value string) error {
		if len(value) == 0 {
			return errors.New(G("Required."))
		}
		return nil
	}
}

// FormErrors holds errors for form fields.
//
// If field 'foo.bar' has an error err, then formErrors["foo.bar:error"] ==
// err.
type FormErrors map[string]string

// check if the given field's value is valid.
//
// If it's not valid, add an error to the formErrors.
func (f *FormErrors) Check(field string, value string, validators ...FormValidator) {
	for _, validator := range validators {
		if err := validator(value); err != nil {
			(*f)[field+":error"] = err.Error()
		}
	}
}


// ToTemplateErrors converts a schema.MultiError to a string map.
//
// An error for the field Foo.Bar will be available under the key
// Foo.Bar:error
func ToTemplateErrors(error schema.MultiError) map[string]string {
	vs := make(map[string]string)
	for field, msg := range error {
		vs[field+":error"] = msg.Error()
	}
	return vs
}

