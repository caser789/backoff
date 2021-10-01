package backoff

import (
	"runtime"
	"time"
)

type Ticker struct {
	C    <-chan time.Time
	c    chan time.Time
	b    BackOff
	stop chan struct{}
}

func NewTicker(b BackOff) *Ticker {
	c := make(chan time.Time)
	t := &Ticker{
		C:    c,
		c:    c,
		b:    b,
		stop: make(chan struct{}, 1),
	}
	go t.run()
	runtime.SetFinalizer(t, func(x *Ticker) { x.Stop() })
	return t
}

func (t *Ticker) Stop() {
	select {
	case t.stop <- struct{}{}:
	default:
	}
}

func (t *Ticker) run() {
	var (
		next   time.Duration
		afterC <-chan time.Time
		tick   time.Time
	)

	defer close(t.c)
	t.b.Reset()

	send := func() {
		select {
		case t.c <- tick:
		case <-t.stop:
			return
		}

		next = t.b.NextBackOff()
		if next == Stop {
			t.Stop()
			return
		}
		afterC = time.After(next)
	}

	send() // Ticker is guaranteed to tick at least once.
	for {
		select {
		case tick = <-afterC:
			send()
		case <-t.stop:
			return
		}
	}
}
