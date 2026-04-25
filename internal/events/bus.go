// Package events provides an in-process async event bus.
// Callers publish events without blocking; registered handlers receive them
// in a dedicated goroutine so multiple subscribers can hook in independently.
package events

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Kind identifies the type of event.
type Kind string

const (
	KindSavedViewCreated Kind = "saved_view.created"
	KindSavedViewUpdated Kind = "saved_view.updated"
	KindSavedViewDeleted Kind = "saved_view.deleted"
)

// Event carries everything a subscriber needs to react to an action.
type Event struct {
	Kind        Kind
	At          time.Time
	WorkspaceID uuid.UUID
	ActorID     string
	ActorEmail  string
	ObjectID    string
	ObjectName  string
	Meta        map[string]any
}

// Handler is a function called for every event delivered to a subscriber.
// Handlers run sequentially per event inside the bus goroutine — keep them fast.
// For slow work (e.g. HTTP webhooks) spawn a goroutine inside the handler.
type Handler func(Event)

// Bus is a buffered, single-goroutine dispatcher.
// Publish never blocks the caller — events are dropped when the buffer is full.
type Bus struct {
	ch   chan Event
	subs []Handler
	mu   sync.RWMutex
	done chan struct{}
}

// New creates a Bus with the given buffer size and starts its dispatch loop.
func New(bufSize int) *Bus {
	if bufSize <= 0 {
		bufSize = 256
	}
	b := &Bus{
		ch:   make(chan Event, bufSize),
		done: make(chan struct{}),
	}
	go b.loop()
	return b
}

// Subscribe registers h to receive all future events.
// Safe to call at any time, including after the first Publish.
func (b *Bus) Subscribe(h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, h)
}

// Publish enqueues e for delivery. If At is zero it is set to now.
// Returns immediately; never blocks the caller.
func (b *Bus) Publish(e Event) {
	if e.At.IsZero() {
		e.At = time.Now()
	}
	select {
	case b.ch <- e:
	default:
		// buffer full — drop rather than stall the HTTP request
	}
}

// Shutdown drains the channel and waits for the dispatch loop to finish.
// No more events should be published after calling Shutdown.
func (b *Bus) Shutdown() {
	close(b.ch)
	<-b.done
}

func (b *Bus) loop() {
	defer close(b.done)
	for e := range b.ch {
		b.mu.RLock()
		subs := make([]Handler, len(b.subs))
		copy(subs, b.subs)
		b.mu.RUnlock()
		for _, h := range subs {
			h(e)
		}
	}
}
