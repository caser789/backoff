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
	var next time.Duration
	var afterC <-chan time.Time

	defer close(t.c)

	first := make(chan time.Time, 1)
	first <- time.Now()

	for {
		select {
		case tick := <-first:
			t.c <- tick
			t.b.Reset()
			next = t.b.NextBackOff()
			if next == Stop {
				return
			}
			afterC = time.After(next)
		case tick := <-afterC:
			t.c <- tick
			next = t.b.NextBackOff()
			if next == Stop {
				return
			}
			afterC = time.After(next)
		case <-t.stop:
			return
		}
	}
}
