package debounce_test

import (
	"testing"
	"time"

	"github.com/marcus999/go-config/pkg/debounce"

	"github.com/marcus999/go-testpredicate"
	"github.com/marcus999/go-testpredicate/pred"
)

// ---------------------------------------------------------------------------
// debounce.New()
// ---------------------------------------------------------------------------

func drain(c <-chan struct{}) (count int) {
	for {
		_, ok := <-c
		if !ok {
			return
		}
		count++
	}
}

func TestEmpty(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.New(2*time.Millisecond, 20*time.Millisecond)
	close(in)

	r := drain(out)
	assert.That(r, pred.IsEqualTo(0))
}

func TestWithMaxDelay(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.New(2*time.Millisecond, 20*time.Millisecond)

	go func() {
		for i := 0; i < 30; i++ {
			in <- debounce.Event
			time.Sleep(1 * time.Millisecond)
		}
		close(in)
	}()

	count := drain(out)
	assert.That(count, pred.IsEqualTo(2))
}

func TestWithNoMax(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.New(3*time.Millisecond, 0)

	go func() {
		for i := 0; i < 10; i++ {
			in <- debounce.Event
			time.Sleep(1 * time.Millisecond)
		}

		time.Sleep(5 * time.Millisecond)

		for i := 10; i < 20; i++ {
			time.Sleep(1 * time.Millisecond)
			in <- debounce.Event
		}
		close(in)
	}()

	count := drain(out)
	assert.That(count, pred.IsEqualTo(2))
}

// ---------------------------------------------------------------------------
// debounce.NewGrouped()
// ---------------------------------------------------------------------------

func drainGrouped(c <-chan []interface{}) (r [][]interface{}) {

	for {
		v, ok := <-c
		if !ok {
			return
		}
		r = append(r, v)
	}
}

func TestGroupedEmpty(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewGrouped(2*time.Millisecond, 20*time.Millisecond)
	close(in)

	r := drainGrouped(out)
	assert.That(r, pred.IsEmpty())
}

func TestGroupedWithMaxDelay(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewGrouped(2*time.Millisecond, 20*time.Millisecond)

	go func() {
		for i := 0; i < 30; i++ {
			in <- i
			time.Sleep(1 * time.Millisecond)
		}
		close(in)
	}()

	r := drainGrouped(out)
	assert.That(r, pred.Length(pred.IsEqualTo(2)))
}

func TestGroupedWithNoMax(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewGrouped(2*time.Millisecond, 0)

	go func() {
		for i := 0; i < 10; i++ {
			in <- i
			time.Sleep(1 * time.Millisecond)
		}

		time.Sleep(4 * time.Millisecond)

		for i := 10; i < 20; i++ {
			time.Sleep(1 * time.Millisecond)
			in <- i
		}
		close(in)
	}()

	r := drainGrouped(out)
	assert.That(r, pred.Length(pred.IsEqualTo(2)))
}

// ---------------------------------------------------------------------------
// debounce.NewLast()
// ---------------------------------------------------------------------------

func drainLast(c <-chan interface{}) (r []interface{}) {

	for {
		v, ok := <-c
		if !ok {
			return
		}
		r = append(r, v)
	}
}

func TestLastEmpty(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewLast(2*time.Millisecond, 20*time.Millisecond)
	close(in)

	r := drainLast(out)
	assert.That(r, pred.IsEmpty())
}

func TestLastWithMaxDelay(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewLast(2*time.Millisecond, 20*time.Millisecond)

	go func() {
		for i := 0; i < 30; i++ {
			in <- i
			time.Sleep(1 * time.Millisecond)
		}
		close(in)
	}()

	r := drainLast(out)
	assert.That(r, pred.Length(pred.IsEqualTo(2)))
	assert.That(r[0], pred.CloseTo(15, 2))
	assert.That(r[1], pred.IsEqualTo(29))
}

func TestLastWithNoMax(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewLast(2*time.Millisecond, 0)

	go func() {
		for i := 0; i < 10; i++ {
			in <- i
			time.Sleep(1 * time.Millisecond)
		}

		time.Sleep(1 * time.Millisecond)

		for i := 10; i < 20; i++ {
			time.Sleep(1 * time.Millisecond)
			in <- i
		}
		close(in)
	}()

	r := drainLast(out)
	assert.That(r, pred.IsEqualTo([]int{9, 19}))
}

// ---------------------------------------------------------------------------
// debounce.NewCounted()
// ---------------------------------------------------------------------------

func drainCounted(c <-chan int) (r []int) {

	for {
		v, ok := <-c
		if !ok {
			return
		}
		r = append(r, v)
	}
}

func TestCountedEmpty(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewCounted(2*time.Millisecond, 20*time.Millisecond)
	close(in)

	r := drainCounted(out)
	assert.That(r, pred.IsEmpty())
}

func TestCountedWithMaxDelay(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewCounted(2*time.Millisecond, 20*time.Millisecond)

	go func() {
		for i := 0; i < 30; i++ {
			in <- debounce.Event
			time.Sleep(1 * time.Millisecond)
		}
		close(in)
	}()

	r := drainCounted(out)
	assert.That(r, pred.Length(pred.IsEqualTo(2)))
	assert.That(r[0], pred.CloseTo(15, 2))
	assert.That(r[1], pred.CloseTo(15, 2))
}

func TestCountedWithNoMax(t *testing.T) {
	assert := testpredicate.NewAsserter(t)
	assert.That(nil, pred.IsNil())

	in, out := debounce.NewCounted(2*time.Millisecond, 0)

	go func() {
		for i := 0; i < 10; i++ {
			in <- debounce.Event
			time.Sleep(1 * time.Millisecond)
		}

		time.Sleep(1 * time.Millisecond)

		for i := 10; i < 20; i++ {
			time.Sleep(1 * time.Millisecond)
			in <- debounce.Event
		}
		close(in)
	}()

	r := drainCounted(out)
	assert.That(r, pred.IsEqualTo([]int{10, 10}))
}
