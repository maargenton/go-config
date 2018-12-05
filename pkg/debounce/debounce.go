/*
Package debounce implements debounce logic for external events.

The debounce logic in controlled by a debounce interval during which events
are aggregated together, and an optional max delay that interrupts long
streaks of events.

The 4 variations provided deal with different event and aggregated event
formats.

All variations provide an input and an output channel. Events are fed through
the input channel, and come out of the ouput channel after the debouncing is
applied. Closing the input channel will close the ouput channel after any
pending event has been propagated.
*/
package debounce

import "time"

// Event is a convience value to feed into channels of empty structs
var Event struct{}

// New returns a pair of input / output channels surrounding
// the debounce function logic, taking an empty struct{} as input values
// and emitting a single empty struct{} per grouped input.
func New(
	interval, maxDelay time.Duration) (
	chan<- struct{}, <-chan struct{}) {

	in := make(chan struct{})
	out := make(chan struct{})

	go func() {
		var pending bool
		var t = debounceTimers{
			interval: interval,
			maxDelay: maxDelay,
		}

	loop:
		for {
			select {
			case _, ok := <-in:
				t.clearInterval()
				if ok {
					pending = true
					t.resetInterval()
				} else {
					t.clearInterval()
					break loop
				}
				t.setMaxDelay()

			case <-t.intervalChan:
				out <- Event
				pending = false
				t.clearMaxDelay()

			case <-t.maxDelayChan:
				if pending {
					out <- Event
					pending = false
				}
				t.clearMaxDelay()
				t.clearInterval()
			}
		}

		if pending {
			out <- Event
		}
		close(out)

	}()

	return in, out
}

// NewGrouped returns a pair of input / output channels surrounding
// the debounce function logic, taking a generic interface{} as input values
// and emitting lists of grouped inputs as []interface{}.
func NewGrouped(
	interval, maxDelay time.Duration) (
	chan<- interface{}, <-chan []interface{}) {

	in := make(chan interface{})
	out := make(chan []interface{})

	go func() {
		var pending []interface{}
		var t = debounceTimers{
			interval: interval,
			maxDelay: maxDelay,
		}

	loop:
		for {
			select {
			case v, ok := <-in:
				t.clearInterval()
				if ok {
					pending = append(pending, v)
					t.resetInterval()
				} else {
					t.clearInterval()
					break loop
				}
				t.setMaxDelay()

			case <-t.intervalChan:
				out <- pending
				pending = nil
				t.clearMaxDelay()

			case <-t.maxDelayChan:
				out <- pending
				pending = nil
				t.clearMaxDelay()
				t.clearInterval()
			}
		}

		if len(pending) != 0 {
			out <- pending
		}
		close(out)

	}()

	return in, out
}

// NewLast returns a pair of input / output channels surrounding
// the debounce function logic, taking a generic interface{} as input values
// and emitting the last value of the grouped inputs as an interface{}.
func NewLast(
	interval, maxDelay time.Duration) (
	chan<- interface{}, <-chan interface{}) {

	in := make(chan interface{})
	out := make(chan interface{})

	go func() {
		var last interface{}
		var t = debounceTimers{
			interval: interval,
			maxDelay: maxDelay,
		}

	loop:
		for {
			select {
			case v, ok := <-in:
				t.clearInterval()
				if ok {
					last = v
					t.resetInterval()
				} else {
					t.clearInterval()
					break loop
				}
				t.setMaxDelay()

			case <-t.intervalChan:
				out <- last
				last = nil
				t.clearMaxDelay()

			case <-t.maxDelayChan:
				if last != nil {
					out <- last
				}
				last = nil
				t.clearMaxDelay()
				t.clearInterval()
			}
		}

		if last != nil {
			out <- last
		}
		close(out)

	}()

	return in, out
}

// NewCounted returns a pair of input / output channels surrounding
// the debounce function logic, taking an empty struct{} as input values
// and emitting the number of grouped inputs as an int
func NewCounted(
	interval, maxDelay time.Duration) (
	chan<- struct{}, <-chan int) {

	in := make(chan struct{})
	out := make(chan int)

	go func() {
		var count int
		var t = debounceTimers{
			interval: interval,
			maxDelay: maxDelay,
		}

	loop:
		for {
			select {
			case _, ok := <-in:
				t.clearInterval()
				if ok {
					count++
					t.resetInterval()
				} else {
					t.clearInterval()
					break loop
				}
				t.setMaxDelay()

			case <-t.intervalChan:
				out <- count
				count = 0
				t.clearMaxDelay()

			case <-t.maxDelayChan:
				out <- count
				count = 0
				t.clearMaxDelay()
				t.clearInterval()
			}
		}

		if count != 0 {
			out <- count
		}
		close(out)

	}()

	return in, out
}

// ---------------------------------------------------------------------------
// Shared logic between debounce functions
// ---------------------------------------------------------------------------

type debounceTimers struct {
	interval      time.Duration
	maxDelay      time.Duration
	intervalTimer *time.Timer
	intervalChan  <-chan time.Time
	maxDelayTimer *time.Timer
	maxDelayChan  <-chan time.Time
}

func (t *debounceTimers) resetInterval() {
	t.intervalTimer = time.NewTimer(t.interval)
	t.intervalChan = t.intervalTimer.C
}

func (t *debounceTimers) clearInterval() {
	if t.intervalTimer == nil {
		return
	}

	t.intervalTimer.Stop()
	t.intervalTimer = nil
	t.intervalChan = nil
}

func (t *debounceTimers) setMaxDelay() {
	if t.maxDelayTimer == nil && t.maxDelay != 0 {
		t.maxDelayTimer = time.NewTimer(t.maxDelay)
		t.maxDelayChan = t.maxDelayTimer.C
	}
}

func (t *debounceTimers) clearMaxDelay() {
	if t.maxDelayTimer == nil {
		return
	}

	t.maxDelayTimer.Stop()
	t.maxDelayTimer = nil
	t.maxDelayChan = nil
}
