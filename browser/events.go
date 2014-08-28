package browser

import "net/url"

// EventType describes a specific event.
type EventType int16

const (
	// PreRequestEvent is an event that is called prior to making an HTTP
	// request. The event argument is an instance of *http.Request.
	PreRequestEvent EventType = iota

	// PostRequestEvent is an event that is called after a successful HTTP
	// request. The event argument is an instance of *http.Response.
	PostRequestEvent

	// ClickEvent is an event that is called when a request has been initiated
	// due to clicking a page element such as a link. The event argument is
	// an instance of *url.URL.
	ClickEvent

	// FormSubmitEvent is an event that is called prior to submitting a form.
	// The event arguments are an instance of the *FormArgs type.
	FormSubmitEvent
)

// Eventable describes a type that handles the binding of event handlers to
// event types, and calls the handlers when the event is triggered.
type Eventable interface {
	// On is called to bind an event handler to an event type.
	On(e EventType, handler EventHandler)

	// Do calls the handlers that have been bound to the given event.
	Do(e *Event) error
}

// Event is passed to event handlers when an event is triggered. It contains
// the type of event, the event arguments, and where the event originated.
type Event struct {
	// Type is the type of event being triggered.
	Type EventType

	// Args are optional arguments attached to the event.
	Args interface{}

	// Browser is the browser where the event originated.
	Browser Browsable
}

// FormArgs is a type of event argument used when a form is being submitted.
type FormArgs struct {
	// Values are the form values being submitted.
	Values url.Values

	// Method is the submit method, either "GET" or "POST".
	Method string

	// Action is the URL where the form is being submitted.
	Action *url.URL
}

// EventHandler is a function that is bound to events, and called when the
// event is triggered.
//
// Event dispatching stops when a handler returns an error. The error is then
// returned to the object that triggered the event.
type EventHandler func(e *Event) error

// EventDispatcher handles the binding of event handlers, and the triggering
// of events.
type EventDispatcher struct {
	handlers map[EventType][]EventHandler
}

// NewEventDispatcher creates and returns a new *EventDispatcher type.
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		handlers: make(map[EventType][]EventHandler),
	}
}

// On is called to bind an event handler to an event type.
func (ed *EventDispatcher) On(e EventType, handler EventHandler) {
	ed.handlers[e] = append(ed.handlers[e], handler)
}

// Do calls the handlers that have been bound to the given event.
func (ed *EventDispatcher) Do(e *Event) error {
	for _, handler := range ed.handlers[e.Type] {
		err := handler(e)
		if err != nil {
			return err
		}
	}
	return nil
}
