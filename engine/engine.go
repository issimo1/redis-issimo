package engine

import (
	"github.com/issimo1/redis-issimo/utils/config"
	"github.com/issimo1/redis-issimo/utils/timewheel"
	"sync/atomic"
)

type Engine struct {
	dbSet []*atomic.Value
	delay timewheel.Delay
}

func NewEngine() *Engine {
	engine := &Engine{}

	engine.dbSet = make([]*atomic.Value, config.GlobalConfig.DBCount)
	for i := 0; i < config.GlobalConfig.DBCount; i++ {
		db :=
	}
}
