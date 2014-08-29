package event

import "net/url"

// Event describes a type of event.
type Event int16

const (
	// PreRequest is an event that is called prior to making an HTTP
	// request. The event argument is an instance of *http.Request.
	PreRequest Event = iota

	// PostRequest is an event that is called after a successful HTTP
	// request. The event argument is an instance of *http.Response.
	PostRequest

	// Click is an event that is called when a request has been initiated
	// due to clicking a page element such as a link. The event argument is
	// an instance of *url.URL.
	Click

	// Submit is an event that is called prior to submitting a form.
	// The event arguments are an instance of the *SubmitArgs type.
	Submit

	// RecordPlay is an event that is triggered when a browser recorder
	// starts recording. The event argument is the recorder.
	RecordPlay

	// RecordStopEvent is an event that is triggered when a browser recorder
	// stops recording. The event argument is the recorder.
	RecordStop

	// RecordReplayEvent is an event that is triggered when a browser recorder
	// starts playback. The event arguments is an instance of []*State, which
	// are the recorded states.
	RecordReplay
)

// Handler is an interface that handles triggered events.
//
// Event dispatching stops when the Triggered() method returns an error. The
// error is then returned to the object that triggered the event.
type Handler interface {
	HandleEvent(event Event, sender, args interface{}) error
}

// HandlerFunc is a function that handles triggered events.
//
// Event dispatching stops when a handler returns an error. The error is then
// returned to the object that triggered the event.
type HandlerFunc func(event Event, sender, args interface{}) error

// HandlerMap is a map of event handler functions.
type HandlerMap map[Event][]HandlerFunc

// Eventable describes a type that handles the binding of events to event
// handlers and calls the handlers when the event is triggered.
type Eventable interface {
	// On binds an event to an event handler.
	On(event Event, handler Handler)

	// OnFunc binds an event to an event handling function.
	OnFunc(event Event, handler HandlerFunc)

	// Do calls the handlers that have been bound to the given event.
	Do(event Event, sender, args interface{}) error
}

// Dispatcher implements the Eventable interface.
type Dispatcher struct {
	handlers HandlerMap
}

// NewDispatcher creates and returns a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(HandlerMap),
	}
}

// On binds an event to an event handler.
func (ed *Dispatcher) On(event Event, handler Handler) {
	ed.handlers[event] = append(ed.handlers[event], func(e Event, sender, args interface{}) error {
		return handler.HandleEvent(e, sender, args)
	})
}

// OnEventFunc binds an event to an event handling function.
func (ed *Dispatcher) OnFunc(event Event, handler HandlerFunc) {
	ed.handlers[event] = append(ed.handlers[event], handler)
}

// Do calls the handlers that have been bound to the given event.
func (ed *Dispatcher) Do(event Event, sender, args interface{}) error {
	for _, handler := range ed.handlers[event] {
		err := handler(event, sender, args)
		if err != nil {
			return err
		}
	}
	return nil
}

// SubmitArgs is an event argument used when a form is being submitted.
type SubmitArgs struct {
	// Values are the form values being submitted.
	Values url.Values

	// Method is the submit method, either "GET" or "POST".
	Method string

	// Action is the URL where the form is being submitted.
	Action *url.URL
}