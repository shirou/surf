package browser

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf/errors"
	"net/url"
	"strings"
)

// Submittable represents an element that may be submitted, such as a form.
type Submittable interface {
	Method() string
	Action() string
	Input(name, value string) error
	Click(button string) error
	Submit() error
	Dom() *goquery.Selection
}

// Form is the default form element.
type Form struct {
	bow           Browsable
	selection     *goquery.Selection
	method        string
	action        string
	definedFields []string
	fields        url.Values
	buttons       url.Values
}

// NewForm creates and returns a *Form type.
func NewForm(bow Browsable, s *goquery.Selection) *Form {
	definedFields, fields, buttons := serializeForm(s)
	method, action := formAttributes(bow, s)

	return &Form{
		bow:           bow,
		selection:     s,
		method:        method,
		action:        action,
		definedFields: definedFields,
		fields:        fields,
		buttons:       buttons,
	}
}

// Method returns the form method, eg "GET" or "POST".
func (f *Form) Method() string {
	return f.method
}

// Action returns the form action URL.
// The URL will always be absolute.
func (f *Form) Action() string {
	return f.action
}

// stringInSlice judges wheter slice has the given string.
// Original: http://stackoverflow.com/questions/15323767/how-to-if-x-in-array-in-golang
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Input sets the value of a form field.
func (f *Form) Input(name, value string) error {
	if stringInSlice(name, f.definedFields) {
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
	method, ok := f.selection.Attr("method")
	if !ok {
		method = "GET"
	}
	action, ok := f.selection.Attr("action")
	if !ok {
		action = f.bow.Url().String()
	}
	aurl, err := url.Parse(action)
	if err != nil {
		return err
	}
	aurl = f.bow.ResolveUrl(aurl)

	values := make(url.Values, len(f.fields)+1)
	for name, vals := range f.fields {
		values[name] = vals
	}
	if buttonName != "" {
		values.Set(buttonName, buttonValue)
	}

	if strings.ToUpper(method) == "GET" {
		return f.bow.OpenForm(aurl.String(), values)
	} else {
		return f.bow.PostForm(aurl.String(), values)
	}

	return nil
}

// Serialize converts the form fields into a url.Values type.
// Returns two url.Value types. The first is the form field values, and the
// second is the form button values.
func serializeForm(sel *goquery.Selection) ([]string, url.Values, url.Values) {
	input := sel.Find("input,button")
	var definedfields []string
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
				} else if typ == "radio" || typ == "checkbox" {
					definedfields = append(definedfields, name)
					_, ok := s.Attr("checked")
					if ok {
						val, ok := s.Attr("value")
						if ok {
							fields.Add(name, val)
						}
					}
				} else {
					definedfields = append(definedfields, name)
					val, ok := s.Attr("value")
					if ok {
						fields.Add(name, val)
					}
				}
			}
		}
	})

	selec := sel.Find("select")

	selec.Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok {
			return
		}
		definedfields = append(definedfields, name)
		s.Find("option[selected]").Each(func(_ int, so *goquery.Selection) {
			val, ok := so.Attr("value")
			if ok {
				fields.Add(name, val)
			}
		})
	})

	textarea := sel.Find("textarea")
	textarea.Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok {
			return
		}
		definedfields = append(definedfields, name)
		fields.Add(name, s.Text())
	})

	return definedfields, fields, buttons
}

func formAttributes(bow Browsable, s *goquery.Selection) (string, string) {
	method, ok := s.Attr("method")
	if !ok {
		method = "GET"
	}
	action, ok := s.Attr("action")
	if !ok {
		action = bow.Url().String()
	}
	aurl, err := url.Parse(action)
	if err != nil {
		return "", ""
	}
	aurl = bow.ResolveUrl(aurl)

	return strings.ToUpper(method), aurl.String()
}
