package template

import (
	"code.google.com/p/gorilla/schema"
	"errors"
)


// FormValidator is a function which validates a string.
type formValidator func(string) error

//  Required is a formValidator to check for non empty values.
func required() formValidator {
	return func(value string) error {
		if len(value) == 0 {
			return errors.New("Required.")
		}
		return nil
	}
}

// FormErrors holds errors for form fields.
//
// If field 'foo.bar' has an error err, then formErrors["foo.bar:error"] ==
// err.
type formErrors map[string]string

// check if the given field's value is valid.
//
// If it's not valid, add an error to the formErrors.
func (f *formErrors) Check(field string, value string, validators ...formValidator) {
	for _, validator := range validators {
		if err := validator(value); err != nil {
			(*f)[field+":error"] = err.Error()
		}
	}
}


// TemplateErrors converts a schema.MultiError to a string map.
//
// An error for the field Foo.Bar will be available under the key
// Foo.Bar:error
func toTemplateErrors(error schema.MultiError) map[string]string {
	vs := make(map[string]string)
	for field, msg := range error {
		vs[field+":error"] = msg.Error()
	}
	return vs
}

