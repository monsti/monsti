package form

import (
	"code.google.com/p/gorilla/schema"
	"fmt"
	"github.com/chrneumann/g5t"
	"html"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

var schemaDecoder = schema.NewDecoder()

var G func(string) string = g5t.String

// FieldRenderData contains the data needed for field rendering.
type FieldRenderData struct {
	// Lebel is the field's label.
	Label string
	// LabelTag is the html code for the field's label, e.g.
	// `<label for="the_id">The Label</label>`.
	LabelTag string
	// Input is the input html for the field.
	Input string
	// Help is the help string.
	Help string
	// Errors contains any validation errors.
	Errors []string
}

// RenderData contains the data needed for form rendering. 
type RenderData struct {
	Fields []FieldRenderData
}

type Widget interface {
	HTML(name string, value interface{}) string
}

type Text int

func (t Text) HTML(field string, value interface{}) string {
	return fmt.Sprintf(`<input id="%v" type="text" name="%v" value="%v"/>`,
		strings.ToLower(field), field, html.EscapeString(
			fmt.Sprintf("%v", value)))
}

// Field contains settings for a form field.
type Field struct {
	Label, Help string
	Validator   Validator
	Widget      Widget
}

// Fields is a map of field names to field settings.
type Fields map[string]Field

// Form represents an html form.
type Form struct {
	Fields map[string]Field
	data   interface{}
	errors map[string][]string
}

// NewForm creates a new Form with the given fields with data stored in the
// given pointer to a structure.
func NewForm(data interface{}, fields Fields) *Form {
	if dataType := reflect.TypeOf(data); dataType.Kind() != reflect.Ptr ||
		dataType.Elem().Kind() != reflect.Struct {
		panic("NewForm(data, fields) expects data to be a pointer to a struct.")
	}
	form := Form{data: data, Fields: fields,
		errors: make(map[string][]string, len(fields))}
	return &form
}

// RenderData returns a RenderData struct for the form.
func (f Form) RenderData() (renderData RenderData) {
	dataVal := reflect.ValueOf(f.data).Elem()
	renderData.Fields = make([]FieldRenderData, 0, dataVal.NumField()-1)
	for i := 0; i < dataVal.NumField(); i++ {
		fieldType := dataVal.Type().Field(i)
		fmt.Println(fieldType.Name)
		fieldVal := dataVal.Field(i)
		name := strings.ToLower(fieldType.Name)
		setup, ok := f.Fields[fieldType.Name]
		if !ok {
			panic("Field " + fieldType.Name + " has not been set up.")
		}
		var widget Text
		renderData.Fields = append(renderData.Fields, FieldRenderData{
			Label: setup.Label,
			LabelTag: fmt.Sprintf(`<label for="%v">%v</label>`, name,
				setup.Label),
			Input:  widget.HTML(fieldType.Name, fieldVal.Interface()),
			Help:   setup.Help,
			Errors: f.errors[fieldType.Name]})
	}
	return
}

// Fill fills the form data with the given values and validates the form.
//
// Returns true iff the form validates.
func (f *Form) Fill(values url.Values) bool {
	error := schemaDecoder.Decode(f.data, values)
	switch e := error.(type) {
	case nil:
		return f.validate()
	case schema.MultiError:
		fmt.Println("MultiError", e)
		for field, msg := range e {
			if f.errors[field] == nil {
				f.errors[field] = []string{msg.Error()}
			} else {
				f.errors[field] = append(f.errors[field], msg.Error())
			}
		}
		return false
	default:
		panic(error.Error())
	}
	return true
}

// validate validates the currently present data.
//
// Resets any previous errors.
func (f *Form) validate() bool {
	anyError := false
	for name, field := range f.Fields {
		value := reflect.ValueOf(f.data).Elem().FieldByName(name)
		if value == reflect.ValueOf(nil) {
			panic(fmt.Sprintf("Field '%v' not present in form data structure.",
				name))
		}
		if errors := field.Validator(value.Interface()); errors != nil {
			f.errors[name] = errors
			anyError = true
		}
	}
	return anyError
}

// Validator is a function which validates the given data and returns error
// messages if the data does not validate.
type Validator func(interface{}) []string

// And is a Validator that collects errors of all given validators.
func And(vs ...Validator) Validator {
	return func(value interface{}) []string {
		errors := []string{}
		for _, v := range vs {
			errors = append(errors, v(value)...)
			fmt.Println(errors)
		}
		if len(errors) == 0 {
			return nil
		}
		return errors
	}
}

// Required creates a Validator to check for non empty values.
func Required() Validator {
	return func(value interface{}) []string {
		if value == reflect.Zero(reflect.TypeOf(value)).Interface() {
			return []string{G("Required.")}
		}
		return nil
	}
}

// Regex creates a Validator to check a string for a matching regexp.
//
// If the expression does not match the string to be validated,
// the given error msg is returned.
func Regex(exp, msg string) Validator {
	return func(value interface{}) []string {
		if matched, _ := regexp.MatchString(exp, value.(string)); !matched {
			return []string{msg}
		}
		return nil
	}
}
