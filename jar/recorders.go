package jar

import (
	"github.com/headzoo/surf/event"
	"net/http"
)

// RecorderJar records browser requests which can be replayed by the browser in
// the order they were recorded.
type RecorderJar interface {
	event.Eventable

	// Start begins recording browser requests.
	Start()

	// Stop ends recording browser requests.
	Stop()

	// Replay replays the recorded requests.
	Replay() error

	// HandleEvent is called when an event is triggered that the handler is bound to.
	HandleEvent(event event.Event, sender, args interface{}) error
}

// MemoryRecorder implements an in-memory RecorderJar.
type MemoryRecorder struct {
	*event.Dispatcher

	// requests are the recorded browser requests.
	requests []*http.Request

	// recorded is true when recording has started, or false if not.
	recording bool
}

// NewMemoryRecorder creates and returns a new memory recorder.
func NewMemoryRecorder() *MemoryRecorder {
	return &MemoryRecorder{
		Dispatcher: event.NewDispatcher(),
	}
}

// Start begins recording browser requests.
func (r *MemoryRecorder) Start() {
	r.requests = make([]*http.Request, 0, 100)
	r.recording = true
	r.Do(event.RecordStart, r, nil)
}

// Stop ends recording browser requests.
func (r *MemoryRecorder) Stop() {
	r.recording = false
	r.Do(event.RecordStop, r, nil)
}

// Replay replays the recorded requests.
func (r *MemoryRecorder) Replay() error {
	// The browser binds itself to the event.RecordReplay event. This loop
	// will notify the browser of each request that needs to be replayed.
	for _, request := range r.requests {
		err := r.Do(event.RecordReplay, r, request)
		if err != nil {
			return err
		}
	}
	return nil
}

// HandleEvent is called when an event is triggered that the handler is bound to.
//
// The browser will bind the recorder to the event.PostRequest so the recorder is
// notified of each successful request and record it.
func (r *MemoryRecorder) HandleEvent(e event.Event, _, args interface{}) error {
	if r.recording && e == event.PostRequest {
		res := args.(*http.Response)
		r.requests = append(r.requests, res.Request)
	}
	return nil
}
