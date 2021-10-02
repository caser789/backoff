package backoff

import (
	"context"
	"sync"
	"time"
)

type Ticker struct {
	C        <-chan time.Time
	c        chan time.Time
	b        BackOff
	ctx      context.Context
	timer    Timer
	stop     chan struct{}
	stopOnce sync.Once
}

func NewTicker(b BackOff) *Ticker {
	return NewTickerWithTimer(b, &defaultTimer{})
}

func NewTickerWithTimer(b BackOff, timer Timer) *Ticker {
	if timer == nil {
		timer = &defaultTimer{}
	}
	c := make(chan time.Time)
	t := &Ticker{
		C:     c,
		c:     c,
		b:     b,
		ctx:   getContext(b),
		timer: timer,
		stop:  make(chan struct{}),
	}
	t.b.Reset()
	go t.run()
	return t
}

func (t *Ticker) Stop() {
	t.stopOnce.Do(func() { close(t.stop) })
}

func (t *Ticker) run() {
	c := t.c
	defer close(c)

	// Ticker is guaranteed to tick at least once.
	afterC := t.send(time.Now())

	for {
		if afterC == nil {
			return
		}

		select {
		case tick := <-afterC:
			afterC = t.send(tick)
		case <-t.stop:
			t.c = nil // Prevent future ticks from being sent to the channel.
			return
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *Ticker) send(tick time.Time) <-chan time.Time {
	select {
	case t.c <- tick:
	case <-t.stop:
		return nil
	}

	next := t.b.NextBackOff()
	if next == Stop {
		t.Stop()
		return nil
	}

	t.timer.Start(next)
	return t.timer.C()
}
