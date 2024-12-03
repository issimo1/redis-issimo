package timewheel

import "time"

type Delay struct {
	tw *TimeWheel
}

func NewDelay() *Delay {
	delay := &Delay{}
	delay.tw = New(1*time.Second, 3600)
	delay.tw.Start()
	return delay
}

func (d *Delay) AddAt(expire time.Time, key string, call func()) {
	interval := time.Until(expire)
	d.Add(interval, key, call)
}

func (d *Delay) Add(expire time.Duration, key string, call func()) {
	d.tw.Add(expire, key, call)
}

func (d *Delay) Cancel(key string) {
	d.tw.Cancel(key)
}
