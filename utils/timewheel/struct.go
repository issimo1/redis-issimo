package timewheel

import (
	"container/list"
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
