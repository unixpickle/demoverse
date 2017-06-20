package main

import (
	"fmt"
	"time"

	"github.com/unixpickle/muniverse"
	"github.com/unixpickle/muniverse/chrome"
)

// An EventFilter defines a way of pruning or modifying
// events before they are sent to an environment.
//
// EventFilters are intended to make actions more similar
// to those that could be made by an agent.
// For example, agents might only be parameterized to make
// one mouse movement per frame.
type EventFilter int

const (
	// NoFilter does no event filtering.
	NoFilter EventFilter = iota

	// DeltaFilter uses the minimum set of events to
	// describe how the mouse and keyboard state changed
	// from the beginning to the end of the frame.
	DeltaFilter
)

// String returns a string representation of the filter.
func (e *EventFilter) String() string {
	switch *e {
	case NoFilter:
		return "NoFilter"
	case DeltaFilter:
		return "DeltaFilter"
	}
	panic("invalid EventFilter")
}

// Set sets the filter from a string representation.
func (e *EventFilter) Set(s string) error {
	switch s {
	case "NoFilter":
		*e = NoFilter
	case "DeltaFilter":
		*e = DeltaFilter
	default:
		return fmt.Errorf("set EventFilter: unknown filter '%s'", s)
	}
	return nil
}

type filteredEnv struct {
	muniverse.Env

	filter EventFilter

	mouseX       int
	mouseY       int
	mousePressed bool

	keysPressed map[string]bool
}

// FilterEnv creates an environment with filtered events.
//
// When the result is closed, e is closed as well.
func FilterEnv(e muniverse.Env, filter *EventFilter) muniverse.Env {
	if *filter == NoFilter {
		return e
	}
	return &filteredEnv{
		Env:    e,
		filter: *filter,
	}
}

func (f *filteredEnv) Reset() error {
	if err := f.Env.Reset(); err != nil {
		return err
	}
	f.mouseX = f.Env.Spec().Width / 2
	f.mouseY = f.Env.Spec().Height / 2
	f.mousePressed = false
	f.keysPressed = map[string]bool{}
	return nil
}

func (f *filteredEnv) Step(t time.Duration, events ...interface{}) (float64,
	bool, error) {
	switch f.filter {
	case DeltaFilter:
		events = f.deltaFilter(events)
	}
	return f.Env.Step(t, events...)
}

func (f *filteredEnv) deltaFilter(events []interface{}) []interface{} {
	newX, newY, newMousePressed := f.mouseX, f.mouseY, f.mousePressed
	newKeysPressed := map[string]bool{}
	for _, evt := range events {
		switch evt := evt.(type) {
		case *chrome.MouseEvent:
			newX, newY = evt.X, evt.Y
			if evt.Type == chrome.MousePressed {
				newMousePressed = true
			} else if evt.Type == chrome.MouseReleased {
				newMousePressed = false
			}
		case *chrome.KeyEvent:
			if evt.Type == chrome.KeyDown {
				newKeysPressed[evt.Code] = true
			} else if evt.Type == chrome.KeyUp {
				newKeysPressed[evt.Code] = false
			}
		}
	}

	var newEvents []interface{}
	if newX != f.mouseX || newY != f.mouseY {
		f.mouseX, f.mouseY = newX, newY
		evt := &chrome.MouseEvent{
			Type: chrome.MouseMoved,
			X:    newX,
			Y:    newY,
		}
		if f.mousePressed {
			evt.Button = chrome.LeftButton
		}
		newEvents = append(newEvents, evt)
	}
	if newMousePressed != f.mousePressed {
		f.mousePressed = newMousePressed
		evt := &chrome.MouseEvent{
			Type:   chrome.MousePressed,
			X:      newX,
			Y:      newY,
			Button: chrome.LeftButton,

			// TODO: better way to compute if clicked.
			ClickCount: 1,
		}
		if !newMousePressed {
			evt.Type = chrome.MouseReleased
		}
		newEvents = append(newEvents, evt)
	}

	for key, pressed := range newKeysPressed {
		if pressed != f.keysPressed[key] {
			f.keysPressed[key] = pressed
			evt := chrome.KeyEvents[key]
			if pressed {
				evt.Type = chrome.KeyDown
			} else {
				evt.Type = chrome.KeyUp
			}
			newEvents = append(newEvents, evt)
		}
	}

	return newEvents
}
