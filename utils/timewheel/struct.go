package timewheel

import (
	"container/list"
	"github.com/issimo1/redis-issimo/utils/logger"
	"time"
)

type taskPos struct {
	pos int
	ele *list.Element
}

type task struct {
	delay    time.Duration
	key      string
	circle   int
	callback func()
}

type TimeWheel struct {
	interval   time.Duration // 时间轮间隔
	ticker     *time.Ticker  //定时器
	curSlotPos int           //游标
	slotNum    int           //循环队列大小
	slots      []*list.List
	m          map[string]*taskPos
	addChan    chan *task
	cancelChan chan string
	stopChan   chan struct{}
}

func New(interval time.Duration, slotNum int) *TimeWheel {
	timeWheel := &TimeWheel{
		interval:   interval,
		ticker:     nil,
		curSlotPos: 0,
		slotNum:    slotNum,
		slots:      make([]*list.List, slotNum),
		m:          make(map[string]*taskPos),
		addChan:    make(chan *task),
		cancelChan: make(chan string),
		stopChan:   make(chan struct{}),
	}

	for i := 0; i < slotNum; i++ {
		timeWheel.slots[i] = list.New()
	}
	return timeWheel
}

func (tw *TimeWheel) doTask() {
	for {
		select {
		case <-tw.ticker.C:
			tw.execTask()
		case t := <-tw.addChan:
			tw.addTask(t)
		case key := <-tw.cancelChan:
			tw.cancelTask(key)
		case <-tw.stopChan:
		}
	}
}

func (tw *TimeWheel) execTask() {
	l := tw.slots[tw.curSlotPos]
	if tw.curSlotPos == tw.slotNum-1 {
		tw.curSlotPos = 0
	} else {
		tw.curSlotPos++
	}
	go tw.scanList(l)
}

func (tw *TimeWheel) scanList(l *list.List) {
	for e := l.Front(); e != nil; {
		t := e.Value.(*task)
		if t.circle > 0 {
			t.circle--
			continue
		}

		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(err)
				}
			}()
			call := t.callback
			call()
		}()

		next := e.Next()
		l.Remove(next)
		if t.key != "" {
			delete(tw.m, t.key)
		}
		e = next
	}

}

func (tw *TimeWheel) posAndCircle(d time.Duration) (pos, circle int) {
	delaySecond := int(d.Seconds())
	intervalSecond := int(tw.interval.Seconds())
	pos = (tw.curSlotPos + delaySecond/intervalSecond) % tw.slotNum
	circle = (delaySecond / intervalSecond) / tw.slotNum
	return
}

func (tw *TimeWheel) addTask(t *task) {
	pos, circle := tw.posAndCircle(t.delay)
	t.circle = circle

	element := tw.slots[pos].PushBack(t)
	if t.key != "" {
		if _, ok := tw.m[t.key]; ok {

		}
		tw.m[t.key] = &taskPos{pos: pos, ele: element}
	}
}

func (tw *TimeWheel) cancelTask(key string) {
	taskPos, ok := tw.m[key]
	if !ok {
		return
	}
	tw.slots[taskPos.pos].Remove(taskPos.ele)
	delete(tw.m, key)
}

func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.doTask()
}

func (tw *TimeWheel) Stop() {
	tw.stopChan <- struct{}{}
}

func (tw *TimeWheel) Add(delay time.Duration, key string, call func()) {
	if delay < 0 {
		return
	}
	t := task{
		delay:    delay,
		key:      key,
		callback: call,
	}
	tw.addChan <- &t
}

func (tw *TimeWheel) Cancel(key string) {
	tw.cancelChan <- key
}
