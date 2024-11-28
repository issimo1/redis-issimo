package timewheel

import "time"

type Delay struct {
	tw *TimeWheel
}

func NewDelay(interval time.Duration, slotNums int) *Delay {
	return &Delay{
		tw: New(interval, slotNums),
	}
}
