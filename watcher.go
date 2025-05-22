package main

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/deckarep/golang-set/v2"
)

type EventType int

const (
	EventConnect EventType = iota
	EventDisconnect
	EventError
)

type Event struct {
	Type  EventType
	Path  string
	Error error
}

type watcher struct {
	events chan Event
}

func newWatcher() (*watcher, error) {
	return &watcher{
		events: make(chan Event, 1),
	}, nil
}

func (w *watcher) Events() <-chan Event {
	return w.events
}

func (w *watcher) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	first := true
	prev := mapset.NewSet[string]()
	tick := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}

		cur, err := listPorts()
		if err != nil {
			w.events <- Event{Type: EventError, Error: err}
			return
		}

		if first {
			prev = cur
			first = false
			continue
		}

		log.Debugf("found %d port(s): %s", cur.Cardinality(), slices.Sorted(slices.Values(cur.ToSlice())))

		for _, p := range cur.Difference(prev).ToSlice() {
			w.events <- Event{Type: EventConnect, Path: p}
		}

		for _, p := range prev.Difference(cur).ToSlice() {
			w.events <- Event{Type: EventDisconnect, Path: p}
		}

		prev = cur
	}
}
