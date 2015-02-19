// This file is part of Monsti, a web content management system.
// Copyright 2012-2015 Christian Neumann
//
// Monsti is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option) any
// later version.
//
// Monsti is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
// A PARTICULAR PURPOSE.  See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Monsti.  If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/chrneumann/htmlwidgets"
	"pkg.monsti.org/gettext"
	"pkg.monsti.org/monsti/api/util/i18n"
)

func init() {
	gob.Register(new(TextFieldType))
	gob.Register(new(HTMLFieldType))
	gob.Register(new(BoolFieldType))
	gob.Register(new(DateTimeFieldType))
	gob.Register(new(FileFieldType))
}

type NestedMap map[string]interface{}

func (n NestedMap) Get(id string) interface{} {
	parts := strings.Split(id, ".")
	field := interface{}(map[string]interface{}(n))
	for _, part := range parts {
		var ok bool
		field, ok = field.(map[string]interface{})[part]
		if !ok {
			return nil
		}
	}
	return field
}

func (n NestedMap) Set(id string, value interface{}) {
	parts := strings.Split(id, ".")
	field := interface{}(map[string]interface{}(n))
	for _, part := range parts[:len(parts)-1] {
		next := field.(map[string]interface{})[part]
		if next == nil {
			next = make(map[string]interface{})
			field.(map[string]interface{})[part] = next
		}
		field = next
	}
	field.(map[string]interface{})[parts[len(parts)-1]] = value
}

type Field interface {
	// Init initializes the field.
	Init(*MonstiClient, string) error
	// RenderHTML returns a string or template.HTML to be used in a html
	// template.
	RenderHTML() interface{}
	// Value returns the value of the field, e.g. a boolean value for
	// Bool fields.
	Value() interface{}
	// Load loads the field data using the given function (see also `Dump`).
	//
	// The passed function unmarshals the raw value (as returned by an
	// earlier `Dump`) into the given value.
	Load(func(interface{}) error) error
	// Dump dumps the field data.
	//
	// The dumped value must be something that can be marshalled into
	// JSON by encoding/json.
	Dump() interface{}
	// Adds a form field to the given form.
	//
	// The nested map stores the field values used by the form. Locale
	// is used for translations.
	ToFormField(form *htmlwidgets.Form, values NestedMap, field *FieldConfig,
		locale string)
	// Load values from the form submission
	FromFormField(NestedMap, *FieldConfig)
}

type BoolFieldType int

func (_ BoolFieldType) Field() Field {
	return new(BoolField)
}

// BoolField is a basic boolean field rendered as checkbox.
type BoolField bool

func (t BoolField) Init(*MonstiClient, string) error {
	return nil
}

func (t BoolField) Value() interface{} {
	return bool(t)
}

func (t BoolField) RenderHTML() interface{} {
	return t.Value()
}

func (t *BoolField) Load(f func(interface{}) error) error {
	return f(t)
}

func (t BoolField) Dump() interface{} {
	return t
}

func (t BoolField) ToFormField(form *htmlwidgets.Form, data NestedMap,
	field *FieldConfig, locale string) {
	data.Set(field.Id, t)
	form.AddWidget(new(htmlwidgets.BoolWidget), "Fields."+field.Id,
		field.Name.Get(locale), "")
}

func (t *BoolField) FromFormField(data NestedMap, field *FieldConfig) {
	*t = BoolField(data.Get(field.Id).(bool))
}

func (t *BoolField) Bool() bool {
	return bool(*t)
}

type TextFieldType int

func (_ TextFieldType) Field() Field {
	return new(TextField)
}

// TextField is a basic unicode text field
type TextField string

func (t TextField) Init(*MonstiClient, string) error {
	return nil
}

func (t TextField) Value() interface{} {
	return string(t)
}

func (t TextField) RenderHTML() interface{} {
	return t
}

func (t *TextField) Load(f func(interface{}) error) error {
	return f(t)
}

func (t TextField) Dump() interface{} {
	return string(t)
}

func (t TextField) ToFormField(form *htmlwidgets.Form, data NestedMap,
	field *FieldConfig, locale string) {
	data.Set(field.Id, string(t))
	G, _, _, _ := gettext.DefaultLocales.Use("", locale)
	widget := new(htmlwidgets.TextWidget)
	if field.Required {
		widget.MinLength = 1
		widget.ValidationError = G("Required.")
	}
	form.AddWidget(widget, "Fields."+field.Id, field.Name.Get(locale), "")
}

func (t *TextField) FromFormField(data NestedMap, field *FieldConfig) {
	*t = TextField(data.Get(field.Id).(string))
}

type HTMLFieldType int

func (_ HTMLFieldType) Field() Field {
	return new(HTMLField)
}

// HTMLField is a text area containing HTML code
type HTMLField string

func (t HTMLField) Init(*MonstiClient, string) error {
	return nil
}

func (t HTMLField) Value() interface{} {
	return string(t)
}

func (t HTMLField) RenderHTML() interface{} {
	return template.HTML(t)
}

func (t *HTMLField) Load(f func(interface{}) error) error {
	return f(t)
}

func (t HTMLField) Dump() interface{} {
	return string(t)
}

func (t HTMLField) ToFormField(form *htmlwidgets.Form, data NestedMap,
	field *FieldConfig, locale string) {
	//G, _, _, _ := gettext.DefaultLocales.Use("", locale)
	data.Set(field.Id, string(t))
	widget := form.AddWidget(new(htmlwidgets.TextAreaWidget), "Fields."+field.Id,
		field.Name.Get(locale), "")
	widget.Base().Classes = []string{"html-field"}
}

func (t *HTMLField) FromFormField(data NestedMap, field *FieldConfig) {
	*t = HTMLField(data.Get(field.Id).(string))
}

type FileFieldType int

func (_ FileFieldType) Field() Field {
	return new(FileField)
}

type FileField string

func (t FileField) Init(*MonstiClient, string) error {
	return nil
}

func (t FileField) Value() interface{} {
	return string(t)
}

func (t FileField) RenderHTML() interface{} {
	return template.HTML(t)
}

func (t *FileField) Load(f func(interface{}) error) error {
	return f(t)
}

func (t FileField) Dump() interface{} {
	return ""
}

func (t FileField) ToFormField(form *htmlwidgets.Form, data NestedMap,
	field *FieldConfig, locale string) {
	data.Set(field.Id, "")
	form.AddWidget(new(htmlwidgets.FileWidget), "Fields."+field.Id,
		field.Name.Get(locale), "")
}

func (t *FileField) FromFormField(data NestedMap, field *FieldConfig) {
	*t = FileField(data.Get(field.Id).(string))
}

type DateTimeFieldType int

func (_ DateTimeFieldType) Field() Field {
	return &DateTimeField{}
}

type DateTimeField struct {
	Time     time.Time
	Location *time.Location
}

func (t *DateTimeField) Init(m *MonstiClient, site string) error {
	var timezone string
	err := m.GetSiteConfig(site, "core.timezone", &timezone)
	if err != nil {
		return fmt.Errorf("Could not get timezone: %v", err)
	}
	t.Location, err = time.LoadLocation(timezone)
	if err != nil {
		t.Location = time.UTC
	}
	return nil
}

func (t DateTimeField) RenderHTML() interface{} {
	return t.Time.String()
}

func (t DateTimeField) Value() interface{} {
	return t.Time
}

func (t *DateTimeField) Load(f func(interface{}) error) error {
	var date string
	if err := f(&date); err != nil {
		return err
	}
	val, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return fmt.Errorf("Could not parse the date value: %v", err)
	}
	t.Time = val.In(t.Location)
	return nil
}

func (t DateTimeField) Dump() interface{} {
	return t.Time.UTC().Format(time.RFC3339)
}

func (t DateTimeField) ToFormField(form *htmlwidgets.Form, data NestedMap,
	field *FieldConfig, locale string) {
	data.Set(field.Id, t.Time)
	form.AddWidget(&htmlwidgets.TimeWidget{Location: t.Location},
		"Fields."+field.Id, field.Name.Get(locale), "")
}

func (t *DateTimeField) FromFormField(data NestedMap, field *FieldConfig) {
	time := data.Get(field.Id).(time.Time)
	*t = DateTimeField{Time: time}
}

func initFields(fields map[string]Field, configs []*FieldConfig,
	m *MonstiClient, site string) error {
	for _, config := range configs {
		val := config.Type.Field()
		err := val.Init(m, site)
		if err != nil {
			return fmt.Errorf("Could not init field %q: %v", config.Id, err)
		}
		fields[config.Id] = val
	}
	return nil
}

type FieldType interface {
	// Field returns a new field for the type.
	Field() Field
}

// FieldConfig is the configuration of a field.
type FieldConfig struct {
	// The id of the field, e.g. `core.Title`.
	Id string
	// The type of the field.
	Type FieldType
	// The name of the field as shown in the web interface.
	Name i18n.LanguageMap
	// True if the user has to set this field (if applicable).
	Required bool
	// Hidden fields won't show up in the web interface.
	Hidden bool
}
