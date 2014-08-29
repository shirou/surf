package browser

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/errors"
	"github.com/headzoo/surf/event"
	"net/url"
	"strings"
)

// Submittable represents an element that may be submitted, such as a form.
type Submittable interface {
	Method() string
	Action() *url.URL
	Input(name, value string) error
	Click(button string) error
	Submit() error
	Dom() *goquery.Selection
}

// Form is the default form element.
type Form struct {
	bow       Browsable
	selection *goquery.Selection
	method    string
	action    *url.URL
	fields    url.Values
	buttons   url.Values
}

// NewForm creates and returns a *Form type.
func NewForm(bow Browsable, s *goquery.Selection) *Form {
	fields, buttons := serializeForm(s)
	method, action := formAttributes(bow, s)

	return &Form{
		bow:       bow,
		selection: s,
		method:    method,
		action:    action,
		fields:    fields,
		buttons:   buttons,
	}
}

// Method returns the form method, eg "GET" or "POST".
func (f *Form) Method() string {
	return f.method
}

// Action returns the form action URL.
// The URL will always be absolute.
func (f *Form) Action() *url.URL {
	return f.action
}

// Input sets the value of a form field.
func (f *Form) Input(name, value string) error {
	if _, ok := f.fields[name]; ok {
		f.fields.Set(name, value)
		return nil
	}
	return errors.NewElementNotFound(
		"No input found with name '%s'.", name)
}

// Submit submits the form.
// Clicks the first button in the form, or submits the form without using
// any button when the form does not contain any buttons.
func (f *Form) Submit() error {
	if len(f.buttons) > 0 {
		for name := range f.buttons {
			return f.Click(name)
		}
	}
	return f.send("", "")
}

// Click submits the form by clicking the button with the given name.
func (f *Form) Click(button string) error {
	if _, ok := f.buttons[button]; !ok {
		return errors.NewInvalidFormValue(
			"Form does not contain a button with the name '%s'.", button)
	}
	return f.send(button, f.buttons[button][0])
}

// Dom returns the inner *goquery.Selection.
func (f *Form) Dom() *goquery.Selection {
	return f.selection
}

// send submits the form.
func (f *Form) send(buttonName, buttonValue string) error {
	values := make(url.Values, len(f.fields)+1)
	for name, vals := range f.fields {
		values[name] = vals
	}
	if buttonName != "" {
		values.Set(buttonName, buttonValue)
	}
	f.bow.Do(event.Submit, f, &event.SubmitArgs{
		Values: values,
		Method: f.method,
		Action: f.action,
	})

	if f.method == "GET" {
		return f.bow.OpenForm(f.action.String(), values)
	} else {
		return f.bow.PostForm(f.action.String(), values)
	}

	return nil
}

// Serialize converts the form fields into a url.Values type.
// Returns two url.Value types. The first is the form field values, and the
// second is the form button values.
func serializeForm(sel *goquery.Selection) (url.Values, url.Values) {
	input := sel.Find("input,button")
	if input.Length() == 0 {
		return url.Values{}, url.Values{}
	}

	fields := make(url.Values)
	buttons := make(url.Values)
	input.Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if ok {
			typ, ok := s.Attr("type")
			if ok {
				if typ == "submit" {
					val, ok := s.Attr("value")
					if ok {
						buttons.Add(name, val)
					} else {
						buttons.Add(name, "")
					}
				} else {
					val, ok := s.Attr("value")
					if ok {
						fields.Add(name, val)
					}
				}
			}
		}
	})

	return fields, buttons
}

// formAttributes returns the method and action on the form.
func formAttributes(bow Browsable, s *goquery.Selection) (string, *url.URL) {
	method := strings.ToUpper(attrOrDefault("method", "GET", s))
	action := attrOrDefault("action", bow.Url().String(), s)
	action, err := bow.ResolveStringUrl(action)
	if err != nil {
		action = bow.Url().String()
	}
	au, err := url.Parse(action)
	if err != nil {
		au = bow.Url()
	}

	return method, au
}
