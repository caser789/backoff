package backoff

import "time"

type BackOff interface {
	NextBackOff() time.Duration
	Reset()
}

const Stop time.Duration = -1

type ZeroBackOff struct{}

func (b *ZeroBackOff) Reset() {}

func (b *ZeroBackOff) NextBackOff() time.Duration { return 0 }

type StopBackOff struct{}

func (b *StopBackOff) Reset() {}

func (b *StopBackOff) NextBackOff() time.Duration { return Stop }

type ConstantBackoff struct {
	Interval time.Duration
}

func (b *ConstantBackoff) Reset()                     {}
func (b *ConstantBackoff) NextBackOff() time.Duration { return b.Interval }

func NewConstantBackOff(d time.Duration) *ConstantBackoff {
	return &ConstantBackoff{Interval: d}
}
